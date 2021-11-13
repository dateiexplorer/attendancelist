// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"fmt"
	"net/url"
)

type URLValue struct {
	URL *url.URL
}

func (v URLValue) String() string {
	if v.URL != nil {
		return v.URL.String()
	}
	return ""
}

func (v URLValue) Set(s string) error {
	if u, err := url.Parse(s); err != nil {
		return err
	} else {
		*v.URL = *u
	}
	return nil
}

type config struct {
	qrPort, loginPort, expireDuration int
	loginURL                          *url.URL
	locationsPath, certPath, keyPath  string
}

func (c *config) validate() (bool, []error) {
	errs := make([]error, 0)
	if c.expireDuration <= 0 {
		errs = append(errs, fmt.Errorf("the expire time for access token must be greater than zero"))
	}

	if c.locationsPath == "" {
		errs = append(errs, fmt.Errorf("the path to the locations XML file must be set, e.g. -locations locations.xml"))
	}

	if c.certPath == "" {
		errs = append(errs, fmt.Errorf("the path to the SSL/TLS certificate file must be set, e.g. -cert cert.pem"))
	}
	if c.keyPath == "" {
		errs = append(errs, fmt.Errorf("the path to the SSL/TLS key file must be set, e.g. -key key.pem"))
	}

	return len(errs) == 0, errs
}
