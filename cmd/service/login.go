// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"github.com/dateiexplorer/attendancelist/internal/web"
)

func runLoginService(config config, url *url.URL, validTokens *web.ValidTokens, openSessions *web.OpenSessions, sessionIDs <-chan string, sessionQueue chan<- web.SessionQueueItem) {
	mux := http.NewServeMux()

	// Set up web assets
	assets, _ := fs.Sub(content, "web/static")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	mux.HandleFunc(url.Path, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			performAccess(w, r, validTokens, openSessions, sessionQueue)
		case "POST":
			performLogin(w, r, validTokens, openSessions, sessionQueue, sessionIDs)
		}
	})

	go func() {
		log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf(":%v", config.loginPort), config.certPath, config.keyPath, mux))
	}()
}

func onAccessDenied(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	t := template.Must(template.ParseFS(content, "web/templates/loginservice/403.html"))
	t.Execute(w, nil)
}

func onInvalidCookie(w http.ResponseWriter, validToken *web.AccessToken) {
	t := template.Must(template.ParseFS(content, "web/templates/loginservice/form.html"))
	t.Execute(w, struct {
		Person *journal.Person
		Token  *web.AccessToken
	}{
		Person: nil,
		Token:  validToken,
	})
}

func performAccess(w http.ResponseWriter, r *http.Request, validTokens *web.ValidTokens, openSessions *web.OpenSessions, sessionQueue chan<- web.SessionQueueItem) {
	// Parameter "token" must be set.
	tokenID := r.URL.Query().Get("token")
	if tokenID == "" {
		onAccessDenied(w)
		return
	}

	token, ok := validTokens.GetByID(tokenID)
	// If token is not in validTokens, deny access.
	if !ok {
		onAccessDenied(w)
		return
	}

	// Check if user cookie is set
	cookie, err := r.Cookie("user")
	if err != nil {
		// No cookie found
		// User login for the first time or cookie is lost
		// In both cases the form must be filled in
		onInvalidCookie(w, token)
		return
	}

	// Check if cookie is valid
	userCookie, err := validateCookie(cookie)
	if err != nil {
		onInvalidCookie(w, token)
		return
	}

	// Cookie is available and valid
	// Search for UserSession with the same location
	// If found, perform logout
	if userSession, ok := openSessions.GetSessionForUser(userCookie.Hash); ok {
		if userSession.Location == token.Location {
			// Same location => perform logout
			sessionQueue <- web.CloseSession(timeutil.Now(), userSession, userCookie.Person)
			t := template.Must(template.ParseFS(content, "web/templates/loginservice/logout.html"))
			t.Execute(w, token.Location)
			return
		}
	}

	// If the location isn't the same as the location in the UserSession
	// or UserSession doesn't exists, show filled login form
	t := template.Must(template.ParseFS(content, "web/templates/loginservice/form.html"))
	t.Execute(w, struct {
		Person *journal.Person
		Token  *web.AccessToken
	}{
		Person: userCookie.Person,
		Token:  token,
	})
}

type LoginForm struct {
	tokenID             string
	location            journal.Location
	firstName, lastName string
	street, number      string
	zipCode, city       string
}

func (f *LoginForm) isValid() bool {
	return f.tokenID != "" && f.location != "" && f.firstName != "" && f.lastName != "" && f.street != "" && f.number != "" && f.zipCode != "" && f.city != ""
}

func validateCookie(cookie *http.Cookie) (*web.UserCookie, error) {
	var userCookie web.UserCookie

	// Check if cookie is valid
	decodedCookie, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		// Cannot decode = Invalid Cookie
		return nil, fmt.Errorf("cannot decode cookie: %w", err)
	}

	err = json.Unmarshal([]byte(decodedCookie), &userCookie)
	if err != nil {
		// Cannot unmarshal = Invalid Cookie
		return nil, fmt.Errorf("cannot unmarshal cookie: %w", err)
	}

	hash, err := web.Hash(*userCookie.Person, privServerSecret)
	if err != nil || hash != userCookie.Hash {
		return nil, fmt.Errorf("cookie is invalid")
	}

	return &userCookie, nil
}

func performLogin(w http.ResponseWriter, r *http.Request, validTokens *web.ValidTokens, openSessions *web.OpenSessions, sessionQueue chan<- web.SessionQueueItem, sessionIDs <-chan string) {
	// Perform login
	// Check if data is complete and user have access to perform this login.
	if err := r.ParseForm(); err != nil {
		// Form cannot be parsed, Error
		w.WriteHeader(http.StatusInternalServerError)
		t := template.Must(template.ParseFS(content, "web/templates/loginservice/500.html"))
		t.Execute(w, err.Error())
		return
	}

	form := LoginForm{
		tokenID:   r.FormValue("tokenId"),
		location:  journal.Location(r.FormValue("location")),
		firstName: r.FormValue("firstName"),
		lastName:  r.FormValue("lastName"),
		street:    r.FormValue("street"),
		number:    r.FormValue("number"),
		zipCode:   r.FormValue("zipCode"),
		city:      r.FormValue("city"),
	}

	if !form.isValid() {
		w.WriteHeader(http.StatusInternalServerError)
		t := template.Must(template.ParseFS(content, "web/templates/loginservice/500.html"))
		t.Execute(w, fmt.Errorf("form is not complete"))
		return
	}

	// Check if user can do perform a login.
	token, ok := validTokens.GetByID(form.tokenID)
	if !ok || token.Location != form.location {
		// Access denied.
		onAccessDenied(w)
		return
	}

	// Show login page and set new UserCookie
	person := journal.NewPerson(form.firstName, form.lastName, form.street, form.number, form.zipCode, form.city)
	userCookie, hash := web.CreateUserCookie(&person, privServerSecret)
	http.SetCookie(w, userCookie)

	// Token is valid, check if session exists
	// If UserSession exists, perform first logout and login afterwards
	if userSession, ok := openSessions.GetSessionForUser(hash); ok {
		sessionQueue <- web.CloseSession(timeutil.Now(), userSession, &person)
	}

	sessionQueue <- web.OpenSession(sessionIDs, timeutil.Now(), &person, form.location, privServerSecret)
	t := template.Must(template.ParseFS(content, "web/templates/loginservice/login.html"))
	t.Execute(w, form.location)
}
