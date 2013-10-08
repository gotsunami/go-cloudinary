// Copyright 2013 Mathias Monnerville and Anthony Baillard.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package cloudinary

import (
	"fmt"
	"testing"
)

func TestDial(t *testing.T) {
	if _, err := Dial("baduri::"); err == nil {
		t.Error("should fail on bad uri")
	}

	// Not a cloudinary:// URL scheme
	if _, err := Dial("http://localhost"); err == nil {
		t.Error("should fail if URL scheme different of cloudinary://")
	}

	// Missing API secret (password)?
	if _, err := Dial("cloudinary://login@cloudname"); err == nil {
		t.Error("should fail when no API secret is provided")
	}

	k := &Service{
		cloudName: "cloudname",
		apiKey:    "login",
		apiSecret: "secret",
	}
	s, err := Dial(fmt.Sprintf("cloudinary://%s:%s@%s", k.apiKey, k.apiSecret, k.cloudName))
	if err != nil {
		t.Error("expect a working service at this stage but got an error.")
	}
	if s.cloudName != k.cloudName || s.apiKey != k.apiKey || s.apiSecret != k.apiSecret {
		t.Errorf("wrong service instance. Expect %v, got %v", k, s)
	}
	uexp := fmt.Sprintf("%s/%s/image/upload/", baseUploadUrl, s.cloudName)
	if s.uploadURI.String() != uexp {
		t.Errorf("wrong upload URI. Expect %s, got %s", uexp, s.uploadURI.String())
	}

}

func TestVerbose(t *testing.T) {
	s := new(Service)
	s.Verbose(true)
	if !s.verbose {
		t.Errorf("wrong verbose attribute. Expect %v, got %v", true, s.verbose)
	}
}

func TestSimulate(t *testing.T) {
	s := new(Service)
	s.Simulate(true)
	if !s.simulate {
		t.Errorf("wrong simulate attribute. Expect %v, got %v", true, s.simulate)
	}
}

func TestKeepFiles(t *testing.T) {
	s := new(Service)
	if err := s.KeepFiles(""); err != nil {
		t.Error("empty pattern should not raise an error")
	}
	pat := "[[;"
	if err := s.KeepFiles(pat); err == nil {
		t.Errorf("wrong pattern %s should raise an error", pat)
	}
	pat = "images/\\.jpg$"
	err := s.KeepFiles(pat)
	if err != nil {
		t.Errorf("valid pattern should return no error", pat)
	}
	if s.keepFilesPattern == nil {
		t.Errorf(".keepFilesPattern attribute is still nil with a valid pattern")
	}
}

func TestUseDatabase(t *testing.T) {
	s := new(Service)
	if err := s.UseDatabase("baduri::"); err == nil {
		t.Error("should fail on bad uri")
	}

}
