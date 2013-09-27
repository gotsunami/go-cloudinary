// Copyright 2013 Mathias Monnerville. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/matm/go-cloudinary"
	"github.com/outofpluto/goconfig/config"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	CloudinaryURI    *url.URL
	MongoURI         *url.URL
	KeepFilesPattern string
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

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Usage: %s [options] settings.conf \n", os.Args[0]))
		fmt.Fprintf(os.Stderr, `
The config file is a plain text file with a [cloudinary] section, e.g

[cloudinary]
uri=cloudinary://api_key:api_secret@cloud_name
`)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	uploadAsRaw := flag.String("upr", "", "path to the file or directory to upload as raw files")
	uploadAsImg := flag.String("upi", "", "path to the file or directory to upload as image files")
	dropImg := flag.String("rmi", "", "delete remote image by public_id")
	dropRaw := flag.String("rmr", "", "delete remote raw file by public_id")
	dropAll := flag.Bool("rmall", false, "delete all (images and raw) remote files")
	dropAllImages := flag.Bool("rmalli", false, "delete all remote images files")
	dropAllRaws := flag.Bool("rmallr", false, "delete all remote raw files")
	listImages := flag.Bool("lsi", false, "List all remote images")
	listRaws := flag.Bool("lsr", false, "List all remote raw files")
	urlImg := flag.String("urli", "", "URL to the uploaded image")
	urlRaw := flag.String("urlr", "", "URL to the uploaded raw file")
	verbose := flag.Bool("v", false, "verbose output")
	simulate := flag.Bool("s", false, "simulate, do nothing (dry run)")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprint(os.Stderr, "Missing config file\n")
		flag.Usage()
	}

	var err error
	settings, err := LoadConfig(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", flag.Arg(0), err.Error())
		os.Exit(1)
	}

	service, err = cloudinary.Dial(settings.CloudinaryURI.String())
	service.Verbose(*verbose)
	service.Simulate(*simulate)
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

	if *simulate {
		fmt.Println("*** DRY RUN MODE ***")
	}

	if *uploadAsRaw != "" {
		fmt.Println("Uploading as raw data ...")
		if err := service.Upload(*uploadAsRaw, nil, false, cloudinary.RawType); err != nil {
			perror(err)
		}
	} else if *uploadAsImg != "" {
		fmt.Println("Uploading as images ...")
		if err := service.Upload(*uploadAsImg, nil, false, cloudinary.ImageType); err != nil {
			perror(err)
		}
	} else if *dropImg != "" {
		fmt.Printf("Deleting image %s ...\n", *dropImg)
		if err := service.Delete(*dropImg, cloudinary.ImageType); err != nil {
			perror(err)
		}
	} else if *dropRaw != "" {
		fmt.Printf("Deleting raw file %s ...\n", *dropRaw)
		if err := service.Delete(*dropRaw, cloudinary.RawType); err != nil {
			perror(err)
		}
	} else if *dropAll {
		fmt.Println("Drop all")
		if err := service.DropAll(os.Stdout); err != nil {
			perror(err)
		}
	} else if *dropAllImages {
		fmt.Println("Drop all images")
		if err := service.DropAllImages(os.Stdout); err != nil {
			perror(err)
		}
	} else if *dropAllRaws {
		fmt.Println("Drop all raw files")
		if err := service.DropAllRaws(os.Stdout); err != nil {
			perror(err)
		}
	} else if *listImages {
		printResources(service.Resources(cloudinary.ImageType))
	} else if *listRaws {
		printResources(service.Resources(cloudinary.RawType))
	} else if *urlImg != "" {
		fmt.Println(service.Url(*urlImg, cloudinary.ImageType))
	} else if *urlRaw != "" {
		fmt.Println(service.Url(*urlRaw, cloudinary.RawType))
	}

	if err != nil {
		fail(err.Error())
	}
}
