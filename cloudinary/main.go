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
	CloudinaryURI *url.URL
	MongoURI      *url.URL
}

var service *cloudinary.Service

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
	return settings, nil
}

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

func printResources(res []*cloudinary.Resource, err error) {
	if err != nil {
		fatal(err.Error())
	}
	fmt.Printf("%-30s %-10s %-5s %s\n", "public_id", "Version", "Type", "Size")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range res {
		fmt.Printf("%-30s %d %s %10d\n", r.PublicId, r.Version, r.ResourceType, r.Size)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Usage: %s [options] settings.conf \n", os.Args[0]))
		fmt.Fprintf(os.Stderr, `
Without any option supplied, it will read the config file and check
ressource (cloudinary, mongodb) availability.

`)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	uploadAsRaw := flag.String("uploadasraw", "", "path to the file or directory to upload as raw files")
	uploadAsImg := flag.String("uploadasimg", "", "path to the file or directory to upload as image files")
	dropImg := flag.String("dropimg", "", "delete remote image by public_id")
	dropRaw := flag.String("dropraw", "", "delete remote raw file by public_id")
	dropAll := flag.Bool("dropall", false, "delete all (images and raw) remote files")
	dropAllImages := flag.Bool("dropallimages", false, "delete all remote images files")
	dropAllRaws := flag.Bool("dropallraws", false, "delete all remote raw files")
	listImages := flag.Bool("listimages", false, "List all remote images")
	listRaws := flag.Bool("listraws", false, "List all remote raw files")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprint(os.Stderr, "Missing config file\n")
		flag.Usage()
	}

	settings, err := LoadConfig(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", flag.Arg(0), err.Error())
		os.Exit(1)
	}

	service, err = cloudinary.Dial(settings.CloudinaryURI.String())
	if err != nil {
		fatal(err.Error())
	}

	// Upload file
	if *uploadAsRaw != "" {
		fmt.Println("Uploading as raw data ...")
		if err := service.Upload(*uploadAsRaw, false, cloudinary.RawType); err != nil {
			fatal(err.Error())
		}
	} else if *uploadAsImg != "" {
		fmt.Println("Uploading as images ...")
		if err := service.Upload(*uploadAsImg, false, cloudinary.ImageType); err != nil {
			fatal(err.Error())
		}
	} else if *dropImg != "" {
		fmt.Printf("Deleting image %s ...\n", *dropImg)
		if err := service.Delete(*dropImg, cloudinary.ImageType); err != nil {
			fatal(err.Error())
		}
	} else if *dropRaw != "" {
		fmt.Printf("Deleting raw file %s ...\n", *dropRaw)
		if err := service.Delete(*dropRaw, cloudinary.RawType); err != nil {
			fatal(err.Error())
		}
	} else if *dropAll {
		fmt.Println("Drop all")
		if err := service.DropAll(os.Stdout); err != nil {
			fatal(err.Error())
		}
	} else if *dropAllImages {
		fmt.Println("Drop all images")
		if err := service.DropAllImages(os.Stdout); err != nil {
			fatal(err.Error())
		}
	} else if *dropAllRaws {
		fmt.Println("Drop all raw files")
		if err := service.DropAllRaws(os.Stdout); err != nil {
			fatal(err.Error())
		}
	} else if *listImages {
		printResources(service.Images())
	} else if *listRaws {
		printResources(service.RawFiles())
	}
}
