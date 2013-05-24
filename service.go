package cloudinary

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	baseUploadUrl = "http://api.cloudinary.com/v1_1"
)

type Service struct {
	cloudName string
	apiKey    string
	apiSecret string
	uploadURI *url.URL
}

// Dial will use the url to connect to the cloudinary service.
// The uri parameter must be a valid URI with the cloudinary:// scheme,
// i.e. cloudinary://api_key:api_secret@clound_name.
func Dial(uri string) (*Service, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "cloudinary" {
		return nil, errors.New("Missing cloudinary:// scheme in URI")
	}
	secret, exists := u.User.Password()
	if !exists {
		return nil, errors.New("No API secret provided in URI.")
	}
	s := &Service{
		cloudName: u.Host,
		apiKey:    u.User.Username(),
		apiSecret: secret,
	}
	// Upload URI to the service
	up, err := url.Parse(fmt.Sprintf("%s/%s/image/upload/", baseUploadUrl, s.cloudName))
	if err != nil {
		return nil, err
	}
	s.uploadURI = up
	return s, nil
}

// CloudName returns the cloud name used to access the Cloudinary service.
func (s *Service) CloudName() string {
	return s.cloudName
}

// ApiKey returns the API key used to access the Cloudinary service.
func (s *Service) ApiKey() string {
	return s.apiKey
}

// UploadURI returns the URI used to upload content to the Cloudinary service.
func (s *Service) UploadURI() *url.URL {
	return s.uploadURI
}

// Upload a file name in the cloud. publicId is the unique identifier of the
// ressource in the cloud. It can be an empty string.
func (s *Service) Upload(path, publicId string) error {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	// Write public ID
	if publicId != "" {
		pi, err := w.CreateFormField("public_id")
		if err != nil {
			return err
		}
		pi.Write([]byte(publicId))
	}

	// Write API key
	ak, err := w.CreateFormField("api_key")
	if err != nil {
		return err
	}
	ak.Write([]byte(s.apiKey))

	// Write timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	ts, err := w.CreateFormField("timestamp")
	if err != nil {
		return err
	}
	ts.Write([]byte(timestamp))

	// Write signature
	hash := sha1.New()
	part := fmt.Sprintf("timestamp=%s%s", timestamp, s.apiSecret)
	if publicId != "" {
		part = fmt.Sprintf("public_id=%s&%s", publicId, part)
	}
	io.WriteString(hash, part)
	signature := fmt.Sprintf("%x", hash.Sum(nil))

	si, err := w.CreateFormField("signature")
	if err != nil {
		return err
	}
	si.Write([]byte(signature))

	// Write file field
	fw, err := w.CreateFormFile("file", path)
	if err != nil {
		return err
	}
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = io.Copy(fw, fd)
	if err != nil {
		return err
	}
	// Don't forget to close the multipart writer to get a terminating boundary
	w.Close()

	req, err := http.NewRequest("POST", s.uploadURI.String(), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	fmt.Println(resp.Status, s.uploadURI)
	io.Copy(os.Stderr, resp.Body)

	return nil
}

// Upload all files in a directory, recursively.
func (s *Service) UploadDir(path string) error {
	return nil
}

// Url returns the complete access path in the cloud to the
// ressource designed by publicId or the empty string if
// no match.
func (s *Service) Url(publicId string) string {
	return ""
}
