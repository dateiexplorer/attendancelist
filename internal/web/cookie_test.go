// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

// Package web provides all functionality which is necessary for the
// service communication, such as cookies or user session management.
package web

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/dateiexplorer/dhbw-attendancelist/internal/journal"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserCookie(t *testing.T) {
	// Prepare cookie
	person := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "74821", "Mosbach")
	privkey := "privServerSecret"
	hash, err := Hash(&person, privkey)
	assert.NoError(t, err)

	cookie := UserCookie{&person, hash}

	// Make JSON object from cookie
	json, err := json.Marshal(cookie)
	assert.NoError(t, err)

	// Encode JSON object for cookie value
	encodedValue := base64.StdEncoding.EncodeToString(json)

	expected := &http.Cookie{
		Name:  UserCookieName,
		Value: encodedValue,
	}

	actualCookie, actualHash := CreateUserCookie(&person, privkey)

	assert.Equal(t, expected, actualCookie)
	assert.Equal(t, hash, actualHash)
}
