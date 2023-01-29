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

	"github.com/dateiexplorer/attendancelist/internal/journal"
)

// The name of the UserCookie in the browser.
const UserCookieName = "user"

// A UserCookie represents a cookie which identifies a journal.Person.
type UserCookie struct {
	Person *journal.Person `json:"person"`
	Hash   string          `json:"hash"`
}

// CreateUserCookie creates a new http.Cookie which can be send to the client.
// It takes a journal.Person for which the cookie is issued and the private secret
// of the server to generate a unique hash value.
//
// The function returns a pointer to the http.Cookie and a string wich contains the
// generated hash value as a hexadecimal string.
func CreateUserCookie(person *journal.Person, privkey string) (*http.Cookie, string) {
	hash, _ := Hash(*person, privkey)
	cookie := UserCookie{Person: person, Hash: hash}

	json, _ := json.Marshal(cookie)
	encodedCookie := base64.StdEncoding.EncodeToString(json)

	return &http.Cookie{
		Name:  UserCookieName,
		Value: encodedCookie,
	}, hash
}
