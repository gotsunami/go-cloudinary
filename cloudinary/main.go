package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/matm/go-cloudinary"
	"github.com/outofpluto/goconfig/config"
	"log"
	"mime"
	"net/url"
	"os"
	"path/filepath"
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

func walkIt(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	ftype := mime.TypeByExtension(filepath.Ext(path))
	if strings.HasPrefix(ftype, "image") {
		fmt.Printf("Uploading %s (%s)\n", path, mime.TypeByExtension(filepath.Ext(path)))
		idx := strings.LastIndex(path, "/") //FIXME
		if idx != -1 {
			idx = strings.LastIndex(path[:idx], "/")
		}
		publicId := path[idx+1:]
		fmt.Println(publicId)
		/*
			if err := service.Upload(path, ""); err != nil {
				log.Fatal(err)
			}
		*/
	}
	return nil
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

	uploadPath := flag.String("u", "", "path to the file or directory to upload")
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
		log.Fatal(err)
	}

	// Upload file
	if *uploadPath != "" {
		info, err := os.Stat(*uploadPath)
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() {
			if err := filepath.Walk(*uploadPath, walkIt); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := service.Upload(*uploadPath, ""); err != nil {
				log.Fatal(err)
			}
		}
	}
}
