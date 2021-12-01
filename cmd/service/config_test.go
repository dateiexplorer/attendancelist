// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLValueString(t *testing.T) {
	v := URLValue{nil}
	assert.Equal(t, "", v.String())

	url, err := url.Parse("https://login")
	assert.NoError(t, err)

	v = URLValue{url}
	assert.Equal(t, "https://login", url.String())
}

func TestConfigValidate(t *testing.T) {
	url, err := url.Parse("https://login")
	assert.NoError(t, err)

	config := config{
		qrPort: 4443, loginPort: 4444, expireDuration: 30,
		loginURL:      url,
		locationsPath: "locations.xml", certPath: "cert.pem", keyPath: "key.pem",
	}

	valid, errs := config.validate()
	assert.Equal(t, 0, len(errs))
	assert.True(t, valid)
}

func TestConfigValidateFail(t *testing.T) {
	url, err := url.Parse("https://login")
	assert.NoError(t, err)

	config := config{
		qrPort: 4443, loginPort: 4444, expireDuration: 0,
		loginURL:      url,
		locationsPath: "", certPath: "", keyPath: "",
	}

	valid, errs := config.validate()
	assert.Equal(t, 4, len(errs))
	assert.False(t, valid)
}

func TestConfigValidateExpireDurationIsNegative(t *testing.T) {
	url, err := url.Parse("https://login")
	assert.NoError(t, err)

	config := config{
		qrPort: 4443, loginPort: 4444, expireDuration: -1,
		loginURL:      url,
		locationsPath: "locations.xml", certPath: "cert.pem", keyPath: "key.pem",
	}

	valid, errs := config.validate()
	assert.Equal(t, 1, len(errs))
	assert.False(t, valid)
}
