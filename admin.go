// Copyright 2013 Mathias Monnerville and Anthony Baillard.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloudinary

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	baseAdminUrl = "https://api.cloudinary.com/v1_1"
)

const (
	pathListAllImages = "/resources/image"
	pathListAllRaws   = "/resources/raw"
	pathListSingleImage = "/resources/image/upload/"
  pathListAllVideos = "/resources/video"
)

const (
	maxResults = 2048
)

func (s *Service) dropAllResources(rtype ResourceType, w io.Writer) error {
	qs := url.Values{
		"max_results": []string{strconv.FormatInt(maxResults, 10)},
	}
	path := pathListAllImages
	if rtype == RawType {
		path = pathListAllRaws
	}
	for {
		resp, err := http.Get(fmt.Sprintf("%s%s?%s", s.adminURI, path, qs.Encode()))
		m, err := handleHttpResponse(resp)
		if err != nil {
			return err
		}
		for _, v := range m["resources"].([]interface{}) {
			publicId := v.(map[string]interface{})["public_id"].(string)
			if w != nil {
				fmt.Fprintf(w, "Deleting %s ... ", publicId)
			}
			if err := s.Delete(publicId, "", rtype); err != nil {
				// Do not return. Report the error but continue through the list.
				fmt.Fprintf(w, "Error: %s: %s\n", publicId, err.Error())
			}
		}
		if e, ok := m["next_cursor"]; ok {
			qs.Set("next_cursor", e.(string))
		} else {
			break
		}
	}

	return nil
}

// DropAllImages deletes all remote images from Cloudinary. File names are
// written to io.Writer if available.
func (s *Service) DropAllImages(w io.Writer) error {
	return s.dropAllResources(ImageType, w)
}

// DropAllRaws deletes all remote raw files from Cloudinary. File names are
// written to io.Writer if available.
func (s *Service) DropAllRaws(w io.Writer) error {
	return s.dropAllResources(RawType, w)
}

// DropAll deletes all remote resources (both images and raw files) from Cloudinary.
// File names are written to io.Writer if available.
func (s *Service) DropAll(w io.Writer) error {
	if err := s.DropAllImages(w); err != nil {
		return err
	}
	if err := s.DropAllRaws(w); err != nil {
		return err
	}
	return nil
}

func (s *Service) doGetResources(rtype ResourceType) ([]*Resource, error) {
	qs := url.Values{
		"max_results": []string{strconv.FormatInt(maxResults, 10)},
	}
	path := pathListAllImages
	if rtype == RawType {
		path = pathListAllRaws
	} else if rtype == VideoType {
		path = pathListAllVideos
	}

	allres := make([]*Resource, 0)
	for {
		resp, err := http.Get(fmt.Sprintf("%s%s?%s", s.adminURI, path, qs.Encode()))
		if err != nil {
			return nil, err
		}

		rs := new(resourceList)
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(rs); err != nil {
			return nil, err
		}
		for _, res := range rs.Resources {
			allres = append(allres, res)
		}
		if rs.NextCursor > 0 {
			qs.Set("next_cursor", strconv.FormatInt(rs.NextCursor, 10))
		} else {
			break
		}
	}
	return allres, nil
}

func (s *Service) doGetResourceDetails(publicId string) (*ResourceDetails, error) {
	path := pathListSingleImage

	resp, err := http.Get(fmt.Sprintf("%s%s%s", s.adminURI, path, publicId))
	if err != nil {
		return nil, err
	}
	details := new(ResourceDetails)
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(details); err != nil {
		return nil, err
	}
	return details, nil
}

// Resources returns a list of all uploaded resources. They can be
// images or raw files, depending on the resource type passed in rtype.
// Cloudinary can return a limited set of results. Pagination is supported,
// so the full set of results is returned.
func (s *Service) Resources(rtype ResourceType) ([]*Resource, error) {
	return s.doGetResources(rtype)
}

// GetResourceDetails gets the details of a single resource that is specified by publicId.
func (s *Service) ResourceDetails(publicId string) (*ResourceDetails, error) {
	return s.doGetResourceDetails(publicId)
}
