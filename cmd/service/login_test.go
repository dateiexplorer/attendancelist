// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"strings"
	"testing"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/web"
	"github.com/stretchr/testify/assert"
)

func TestIsFormValid(t *testing.T) {
	form := LoginForm{
		tokenID:   "1234",
		location:  "DHBW Mosbach",
		firstName: "Max",
		lastName:  "Mustermann",
		street:    "Musterstraße",
		number:    "20",
		zipCode:   "1234",
		city:      "Muster",
	}

	valid := form.isValid()
	assert.True(t, valid)
}

func TestIsFormValidFail(t *testing.T) {
	form := LoginForm{
		tokenID:   "",
		location:  "DHBW Mosbach",
		firstName: "Max",
		lastName:  "Mustermann",
		street:    "Musterstraße",
		number:    "20",
		zipCode:   "1234",
		city:      "Muster",
	}

	valid := form.isValid()
	assert.False(t, valid)

	form = LoginForm{
		tokenID:   "1234",
		location:  "",
		firstName: "Max",
		lastName:  "Mustermann",
		street:    "Musterstraße",
		number:    "20",
		zipCode:   "1234",
		city:      "Muster",
	}

	valid = form.isValid()
	assert.False(t, valid)

	form = LoginForm{
		tokenID:   "1234",
		location:  "DHBW Mosbach",
		firstName: "",
		lastName:  "Mustermann",
		street:    "Musterstraße",
		number:    "20",
		zipCode:   "1234",
		city:      "Muster",
	}

	valid = form.isValid()
	assert.False(t, valid)

	form = LoginForm{
		tokenID:   "1234",
		location:  "DHBW Mosbach",
		firstName: "Max",
		lastName:  "",
		street:    "Musterstraße",
		number:    "20",
		zipCode:   "1234",
		city:      "Muster",
	}

	valid = form.isValid()
	assert.False(t, valid)

	form = LoginForm{
		tokenID:   "1234",
		location:  "DHBW Mosbach",
		firstName: "Max",
		lastName:  "Mustermann",
		street:    "",
		number:    "20",
		zipCode:   "1234",
		city:      "Muster",
	}

	valid = form.isValid()
	assert.False(t, valid)

	form = LoginForm{
		tokenID:   "1234",
		location:  "DHBW Mosbach",
		firstName: "Max",
		lastName:  "Mustermann",
		street:    "Musterstraße",
		number:    "",
		zipCode:   "1234",
		city:      "Muster",
	}

	valid = form.isValid()
	assert.False(t, valid)

	form = LoginForm{
		tokenID:   "1234",
		location:  "DHBW Mosbach",
		firstName: "Max",
		lastName:  "Mustermann",
		street:    "Musterstraße",
		number:    "20",
		zipCode:   "",
		city:      "Muster",
	}

	valid = form.isValid()
	assert.False(t, valid)

	form = LoginForm{
		tokenID:   "1234",
		location:  "DHBW Mosbach",
		firstName: "Max",
		lastName:  "Mustermann",
		street:    "Musterstraße",
		number:    "20",
		zipCode:   "1234",
		city:      "",
	}

	valid = form.isValid()
	assert.False(t, valid)
}

func TestValidateCookie(t *testing.T) {
	p := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "1234", "Muster")
	cookie, hash := web.CreateUserCookie(&p, privServerSecret)

	actual, err := validateCookie(cookie)
	assert.NoError(t, err)

	expected := &web.UserCookie{Person: &p, Hash: hash}
	assert.Equal(t, expected, actual)
}

func TestValidateCookieFalseEncoding(t *testing.T) {
	p := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "1234", "Muster")
	cookie, _ := web.CreateUserCookie(&p, privServerSecret)
	cookie.Value = "\x01"

	actual, err := validateCookie(cookie)
	assert.EqualError(t, err, "cannot decode cookie: illegal base64 data at input byte 0")
	assert.Nil(t, actual)
}

func TestValidateCookieFalseHash(t *testing.T) {
	p := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "1234", "Muster")
	// Repeat the server secret 2x to generate an invalid hash
	cookie, _ := web.CreateUserCookie(&p, strings.Repeat(privServerSecret, 2))

	actual, err := validateCookie(cookie)
	assert.EqualError(t, err, "cookie is invalid")
	assert.Nil(t, actual)
}
