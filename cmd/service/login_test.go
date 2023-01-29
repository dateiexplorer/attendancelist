// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
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

func TestPerformAccessTokenNotSetOrInvalid(t *testing.T) {
	var expected bytes.Buffer

	template := template.Must(template.ParseFiles(path.Join("web", "templates", "loginservice", "403.html")))
	template.Execute(&expected, nil)

	validTokens := new(web.ValidTokens)
	openSessions := new(web.OpenSessions)
	sessionQueue := make(chan web.SessionQueueItem)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performAccess(w, r, validTokens, openSessions, sessionQueue)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	// Check if body equals the Forbidden (403.html) template.
	assert.Equal(t, expected.Bytes(), body)

	res, err = http.Get(ts.URL + "/?token=a")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	body, err = io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)
}

func TestPerformAccessCookie(t *testing.T) {
	tokens := []*web.AccessToken{
		{ID: "a", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 0).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Valid: 0, Location: "DHBW Mosbach", QR: []byte{}},
	}

	validTokens := new(web.ValidTokens)
	validTokens.Add(tokens)

	var expected bytes.Buffer

	validToken := tokens[0]
	template := template.Must(template.ParseFiles(path.Join("web", "templates", "loginservice", "form.html")))
	template.Execute(&expected, struct {
		Person *journal.Person
		Token  *web.AccessToken
	}{
		Person: nil,
		Token:  validToken,
	})

	openSessions := new(web.OpenSessions)
	sessionQueue := make(chan web.SessionQueueItem, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performAccess(w, r, validTokens, openSessions, sessionQueue)
	}))
	defer ts.Close()

	// Make request with valid token but without cookie
	res, err := http.Get(ts.URL + "/?token=a")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)

	// Make request with invalid cookie
	req, err := http.NewRequest("GET", ts.URL+"/?token=a", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	p := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74782", "Mosbach")
	cookie, _ := web.CreateUserCookie(&p, strings.Repeat(privServerSecret, 2))
	req.AddCookie(cookie)

	res, err = ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err = io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)

	// Make request with valid cookie
	req, err = http.NewRequest("GET", ts.URL+"/?token=a", nil)
	assert.NoError(t, err)

	cookie, _ = web.CreateUserCookie(&p, privServerSecret)
	req.AddCookie(cookie)

	res, err = ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err = io.ReadAll(res.Body)
	assert.NoError(t, err)

	expected.Reset()
	template.Execute(&expected, struct {
		Person *journal.Person
		Token  *web.AccessToken
	}{
		Person: &p,
		Token:  validToken,
	})

	assert.Equal(t, expected.Bytes(), body)
}

func TestPerformAccessLogin(t *testing.T) {
	tokens := []*web.AccessToken{
		{ID: "a", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 0).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Valid: 0, Location: "DHBW Mosbach", QR: []byte{}},
	}

	validTokens := new(web.ValidTokens)
	validTokens.Add(tokens)

	var expected bytes.Buffer

	p := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74782", "Mosbach")

	// Create valid cookie
	cookie, _ := web.CreateUserCookie(&p, privServerSecret)

	validToken := tokens[0]
	template := template.Must(template.ParseFiles(path.Join("web", "templates", "loginservice", "form.html")))
	template.Execute(&expected, struct {
		Person *journal.Person
		Token  *web.AccessToken
	}{
		Person: &p,
		Token:  validToken,
	})

	openSessions := new(web.OpenSessions)
	sessionQueue := make(chan web.SessionQueueItem, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performAccess(w, r, validTokens, openSessions, sessionQueue)
	}))
	defer ts.Close()

	// Make request
	req, err := http.NewRequest("GET", ts.URL+"/?token=a", nil)
	assert.NoError(t, err)
	req.AddCookie(cookie)

	res, err := ts.Client().Do(req)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)
}

func TestPerformAccessLogout(t *testing.T) {
	tokens := []*web.AccessToken{
		{ID: "a", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 0).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Valid: 0, Location: "DHBW Mosbach", QR: []byte{}},
	}

	validTokens := new(web.ValidTokens)
	validTokens.Add(tokens)

	var expected bytes.Buffer

	p := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74782", "Mosbach")

	// Create valid cookie
	cookie, hash := web.CreateUserCookie(&p, privServerSecret)

	template := template.Must(template.ParseFiles(path.Join("web", "templates", "loginservice", "logout.html")))
	template.Execute(&expected, "DHBW Mosbach")

	openSessions := new(web.OpenSessions)
	session := web.NewSession("a", hash, "DHBW Mosbach")
	openSessions.Store(hash, &session)
	sessionQueue := make(chan web.SessionQueueItem, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performAccess(w, r, validTokens, openSessions, sessionQueue)
	}))
	defer ts.Close()

	// Make request
	req, err := http.NewRequest("GET", ts.URL+"/?token=a", nil)
	assert.NoError(t, err)
	req.AddCookie(cookie)

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	// Get logout message
	assert.Equal(t, expected.Bytes(), body)

	sessionItem := <-sessionQueue
	assert.Equal(t, journal.Logout, sessionItem.Action)
	assert.Equal(t, p, *sessionItem.Person)
	assert.Equal(t, session, *sessionItem.Session)
}

func TestPerformLoginInvalidForm(t *testing.T) {
	var expected bytes.Buffer

	template := template.Must(template.ParseFiles(path.Join("web", "templates", "loginservice", "500.html")))
	template.Execute(&expected, "form is not complete")

	validTokens := new(web.ValidTokens)
	openSessions := new(web.OpenSessions)
	sessionQueue := make(chan web.SessionQueueItem)
	sessionIDs := make(chan string)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performLogin(w, r, validTokens, openSessions, sessionQueue, sessionIDs)
	}))
	defer ts.Close()

	form := url.Values{}
	form.Add("firstName", "Max")

	req, err := http.NewRequest("POST", ts.URL, strings.NewReader(form.Encode()))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)
}

func TestPerformLoginDenyAccess(t *testing.T) {
	var expected bytes.Buffer

	template := template.Must(template.ParseFiles(path.Join("web", "templates", "loginservice", "403.html")))
	template.Execute(&expected, nil)

	tokens := []*web.AccessToken{
		{ID: "a", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 0).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Valid: 0, Location: "DHBW Mosbach", QR: []byte{}},
	}

	validTokens := new(web.ValidTokens)
	validTokens.Add(tokens)

	openSessions := new(web.OpenSessions)
	sessionQueue := make(chan web.SessionQueueItem)
	sessionIDs := make(chan string)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performLogin(w, r, validTokens, openSessions, sessionQueue, sessionIDs)
	}))
	defer ts.Close()

	// Invalid token
	form := url.Values{}
	form.Add("firstName", "Max")
	form.Add("lastName", "Mustermann")
	form.Add("street", "Musterstraße")
	form.Add("number", "20")
	form.Add("zipCode", "74782")
	form.Add("city", "Mosbach")
	form.Add("tokenId", "b")
	form.Add("location", "DHBW Mosbach")

	req, err := http.NewRequest("POST", ts.URL, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)

	// Invalid location
	form = url.Values{}
	form.Add("firstName", "Max")
	form.Add("lastName", "Mustermann")
	form.Add("street", "Musterstraße")
	form.Add("number", "20")
	form.Add("zipCode", "74782")
	form.Add("city", "Mosbach")
	form.Add("tokenId", "a")
	form.Add("location", "Alte Mälzerei")

	req, err = http.NewRequest("POST", ts.URL, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err = ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	body, err = io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)
}

func TestPerformLogin(t *testing.T) {
	var expected bytes.Buffer

	template := template.Must(template.ParseFiles(path.Join("web", "templates", "loginservice", "login.html")))
	template.Execute(&expected, "DHBW Mosbach")

	tokens := []*web.AccessToken{
		{ID: "a", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 0).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Valid: 0, Location: "DHBW Mosbach", QR: []byte{}},
	}

	validTokens := new(web.ValidTokens)
	validTokens.Add(tokens)

	p := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "74782", "DHBW Mosbach")

	cookie, hash := web.CreateUserCookie(&p, privServerSecret)
	cookie.Raw = fmt.Sprintf("%v=%v", cookie.Name, cookie.Value)

	session := web.NewSession("a", hash, "DHBW Mosbach")

	openSessions := new(web.OpenSessions)
	openSessions.Store("a", &session)
	sessionQueue := make(chan web.SessionQueueItem, 2)
	sessionIDs := make(chan string, 2)
	sessionIDs <- "b"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		performLogin(w, r, validTokens, openSessions, sessionQueue, sessionIDs)
	}))
	defer ts.Close()

	// Form
	form := url.Values{}
	form.Add("firstName", p.FirstName)
	form.Add("lastName", p.LastName)
	form.Add("street", p.Address.Street)
	form.Add("number", p.Address.Number)
	form.Add("zipCode", p.Address.ZipCode)
	form.Add("city", p.Address.City)
	form.Add("tokenId", "a")
	form.Add("location", "DHBW Mosbach")

	req, err := http.NewRequest("POST", ts.URL, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, expected.Bytes(), body)

	// Check if cookie was set
	cookies := res.Cookies()
	assert.Equal(t, 1, len(cookies))

	assert.Equal(t, cookie, cookies[0])

	sessionItem := <-sessionQueue
	assert.Equal(t, journal.Logout, sessionItem.Action)
	assert.Equal(t, p, *sessionItem.Person)
	assert.Equal(t, session, *sessionItem.Session)

	sessionItem = <-sessionQueue
	assert.Equal(t, journal.Login, sessionItem.Action)
	assert.Equal(t, p, *sessionItem.Person)
	assert.Equal(t, web.NewSession("b", hash, "DHBW Mosbach"), *sessionItem.Session)
}
