// Copyright 2013 Mathias Monnerville and Anthony Baillard.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gotsunami/go-cloudinary"
	"github.com/outofpluto/goconfig/config"
)

type Config struct {
	// Url to the CLoudinary service.
	CloudinaryURI *url.URL
	// Url to a MongoDB instance, used to track files and upload
	// only changed. Optional.
	MongoURI *url.URL
	// Regexp pattern to prevent remote file deletion.
	KeepFilesPattern string
	// An optional remote prepend path, used to generate a unique
	// data path to a remote resource. This can be useful if public
	// ids are not random (i.e provided as request arguments) to solve
	// any caching issue: a different prepend path generates a new path
	// to the remote resource.
	PrependPath string
	// ProdTag is an alias to PrependPath. If PrependPath is empty but
	// ProdTag is set (with at prodtag= line in the [global] section of
	// the config file), PrependPath is set to ProdTag. For example, it
	// can be used with a DVCS commit tag to force new remote data paths
	// to remote resources.
	ProdTag string
}

var service *cloudinary.Service

// Parses all structure fields values, looks for any
// variable defined as ${VARNAME} and substitute it by
// calling os.Getenv().
//
// The reflect package is not used here since we cannot
// set a private field (not exported) within a struct using
// reflection.
func (c *Config) handleEnvVars() error {
	// [cloudinary]
	if c.CloudinaryURI != nil {
		curi, err := handleQuery(c.CloudinaryURI)
		if err != nil {
			return err
		}
		c.CloudinaryURI = curi
	}
	if len(c.PrependPath) == 0 {
		// [global]
		if len(c.ProdTag) > 0 {
			ptag, err := replaceEnvVars(c.ProdTag)
			if err != nil {
				return err
			}
			c.PrependPath = cloudinary.EnsureTrailingSlash(ptag)
		}
	}

	// [database]
	if c.MongoURI != nil {
		muri, err := handleQuery(c.MongoURI)
		if err != nil {
			return err
		}
		c.MongoURI = muri
	}
	return nil
}

// LoadConfig parses a config file and sets global settings
// variables to be used at runtime. Note that returning an error
// will cause the application to exit with code error 1.
func LoadConfig(path string) (*Config, error) {
	settings := &Config{}

	c, err := config.ReadDefault(path)
	if err != nil {
		return nil, err
	}

	// Cloudinary settings
	var cURI *url.URL
	var uri string

	if uri, err = c.String("cloudinary", "uri"); err != nil {
		return nil, err
	}
	if cURI, err = url.Parse(uri); err != nil {
		return nil, errors.New(fmt.Sprint("cloudinary URI: ", err.Error()))
	}
	settings.CloudinaryURI = cURI

	// An optional remote prepend path
	if prepend, err := c.String("cloudinary", "prepend"); err == nil {
		settings.PrependPath = cloudinary.EnsureTrailingSlash(prepend)
	}
	settings.ProdTag, _ = c.String("global", "prodtag")

	// Keep files regexp? (optional)
	var pattern string
	pattern, _ = c.String("cloudinary", "keepfiles")
	if pattern != "" {
		settings.KeepFilesPattern = pattern
	}

	// mongodb section (optional)
	uri, _ = c.String("database", "uri")
	if uri != "" {
		var mURI *url.URL
		if mURI, err = url.Parse(uri); err != nil {
			return nil, errors.New(fmt.Sprint("mongoDB URI: ", err.Error()))
		}
		settings.MongoURI = mURI
	} else {
		fmt.Fprintf(os.Stderr, "Warning: database not set (upload sync disabled)\n")
	}

	// Looks for env variables, perform substitutions if needed
	if err := settings.handleEnvVars(); err != nil {
		return nil, err
	}
	return settings, nil
}

func fail(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

func printResources(res []*cloudinary.Resource, err error) {
	if err != nil {
		fail(err.Error())
	}
	if len(res) == 0 {
		fmt.Println("No resource found.")
		return
	}
	fmt.Printf("%-30s %-10s %-5s %s\n", "public_id", "Version", "Type", "Size")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range res {
		fmt.Printf("%-30s %d %s %10d\n", r.PublicId, r.Version, r.ResourceType, r.Size)
	}
}

func perror(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(1)
}

func step(caption string) {
	fmt.Printf("==> %s\n", caption)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Usage: %s [options] action settings.conf \n", os.Args[0]))
		fmt.Fprintf(os.Stderr, `
Actions:
ls          list all remote resources
rm          delete a remote resource
up          upload a local resource
url         get the URL of of a remote resource

The config file is a plain text file with a [cloudinary] section, e.g
[cloudinary]
uri=cloudinary://api_key:api_secret@cloud_name
`)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	optRaw := flag.String("r", "", "raw filename or public id")
	optImg := flag.String("i", "", "image filename or public id")
	optVerbose := flag.Bool("v", false, "verbose output")
	optSimulate := flag.Bool("s", false, "simulate, do nothing (dry run)")
	optAll := flag.Bool("a", false, "applies to all resource files")
	flag.Parse()

	if len(flag.Args()) != 2 {
		flag.Usage()
	}

	action := flag.Arg(0)
	supportedAction := func(act string) bool {
		switch act {
		case "ls", "rm", "up", "url":
			return true
		}
		return false
	}(action)
	if !supportedAction {
		fmt.Fprintf(os.Stderr, "Unknown action '%s'\n", action)
		flag.Usage()
	}

	var err error
	settings, err := LoadConfig(flag.Arg(1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", flag.Arg(1), err.Error())
		os.Exit(1)
	}

	service, err = cloudinary.Dial(settings.CloudinaryURI.String())
	service.Verbose(*optVerbose)
	service.Simulate(*optSimulate)
	service.KeepFiles(settings.KeepFilesPattern)
	if settings.MongoURI != nil {
		if err := service.UseDatabase(settings.MongoURI.String()); err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to mongoDB: %s\n", err.Error())
			os.Exit(1)
		}
	}

	if err != nil {
		fail(err.Error())
	}

	if *optSimulate {
		fmt.Println("*** DRY RUN MODE ***")
	}

	if len(settings.PrependPath) > 0 {
		fmt.Println("/!\\ Remote prepend path set to: ", settings.PrependPath)
	} else {
		fmt.Println("/!\\ No remote prepend path set")
	}

	switch action {
	case "up":
		if *optRaw == "" && *optImg == "" {
			fail("Missing -i or -r option.")
		}
		if *optRaw != "" {
			step("Uploading as raw data")
			if _, err := service.UploadStaticRaw(*optRaw, nil, settings.PrependPath); err != nil {
				perror(err)
			}
		} else {
			step("Uploading as images")
			if _, err := service.UploadStaticImage(*optImg, nil, settings.PrependPath); err != nil {
				perror(err)
			}
		}
		break

	case "rm":
		if *optAll {
			step(fmt.Sprintf("Deleting all resources..."))
			if err := service.DropAll(os.Stdout); err != nil {
				perror(err)
			}
		} else {
			if *optRaw == "" && *optImg == "" {
				fail("Missing -i or -r option.")
			}
			if *optRaw != "" {
				step(fmt.Sprintf("Deleting raw file %s", *optRaw))
				if err := service.Delete(*optRaw, settings.PrependPath, cloudinary.RawType); err != nil {
					perror(err)
				}
			} else {
				step(fmt.Sprintf("Deleting image %s", *optImg))
				if err := service.Delete(*optImg, settings.PrependPath, cloudinary.ImageType); err != nil {
					perror(err)
				}
			}
		}

	case "ls":
		fmt.Println("==> Raw resources:")
		printResources(service.Resources(cloudinary.RawType))
		fmt.Println("==> Images:")
		printResources(service.Resources(cloudinary.ImageType))

	case "url":
		if *optRaw == "" && *optImg == "" {
			fail("Missing -i or -r option.")
		}
		if *optRaw != "" {
			fmt.Println(service.Url(*optRaw, cloudinary.RawType))
		} else {
			fmt.Println(service.Url(*optImg, cloudinary.ImageType))
		}
	}

	fmt.Println("")
	if err != nil {
		fail(err.Error())
	}
}
