package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// replaceEnvVars replaces all ${VARNAME} with their value
// using os.Getenv().
func replaceEnvVars(src string) (string, error) {
	r, err := regexp.Compile(`\${([A-Z_]+)}`)
	if err != nil {
		return "", err
	}
	envs := r.FindAllString(src, -1)
	for _, varname := range envs {
		evar := os.Getenv(varname[2 : len(varname)-1])
		if evar == "" {
			return "", errors.New(fmt.Sprintf("error: env var %s not defined", varname))
		}
		src = strings.Replace(src, varname, evar, -1)
	}
	return src, nil
}

func handleQuery(uri *url.URL) (*url.URL, error) {
	qs, err := url.QueryUnescape(uri.String())
	if err != nil {
		return nil, err
	}
	r, err := replaceEnvVars(qs)
	if err != nil {
		return nil, err
	}
	wuri, err := url.Parse(r)
	if err != nil {
		return nil, err
	}
	return wuri, nil
}
