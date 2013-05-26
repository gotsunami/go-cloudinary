package cloudinary

import (
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
)

const (
	maxResults = 10
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
				fmt.Fprintf(w, "Deleting %s ...\n", publicId)
			}
			if err := s.Delete(publicId, rtype); err != nil {
				return err
			}
			// TODO: also delete resource entry from database (if used)
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
