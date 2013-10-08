// Copyright 2013 Mathias Monnerville and Anthony Baillard.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloudinary

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
)

// Returns SHA1 file checksum
func fileChecksum(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha1.New()
	io.WriteString(hash, string(data))
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
