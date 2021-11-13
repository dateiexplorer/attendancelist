// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"github.com/dateiexplorer/attendancelist/internal/web"
)

const privServerSecret = "privateServerSecret"

//go:embed web/*
var content embed.FS

func getToken(w http.ResponseWriter, r *http.Request, validTokens *web.ValidTokens) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	// Without location query param get all tokens
	if !query.Has("location") {
		tokens := validTokens.GetAll()

		res, _ := json.Marshal(tokens)
		w.Write(res)
		return
	}

	location := journal.Location(query.Get("location"))

	if token, ok := validTokens.GetCurrentForLocation(location); ok {
		res, _ := json.Marshal(token)
		w.Write(res)
		return
	}

	err := fmt.Errorf("no valid token found for this location")
	w.Write([]byte(fmt.Sprintf("{\"err\":\"%v\"}", err)))
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

func main() {
	var loginURL, _ = url.Parse("https://localhost:4444/login")
	var locationsPath, certPath, keyPath string
	var loginPort, qrPort, expireDuration int

	flag.IntVar(&expireDuration, "expire", 60, "The expire duration for an access token in seconds")
	flag.IntVar(&qrPort, "qr-port", 4443, "The port the QR code service should running on")
	flag.IntVar(&loginPort, "login-port", 4444, "The port the login service should running on")
	flag.Var(&URLValue{loginURL}, "login-url", "The `url` encoded in the QR code for the login page")
	flag.StringVar(&locationsPath, "locations", "", "The locations XML file `path` in the file system")
	flag.StringVar(&certPath, "cert", "", "The `path` to the SSL/TLS certificate file")
	flag.StringVar(&keyPath, "key", "", "The `path` to the SSL/TLS key file")
	flag.Parse()

	config := config{
		qrPort:         qrPort,
		loginPort:      loginPort,
		expireDuration: expireDuration,
		loginURL:       loginURL,
		locationsPath:  locationsPath,
		certPath:       certPath,
		keyPath:        keyPath,
	}

	// Validate configuration
	if valid, errs := config.validate(); !valid {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}

		os.Exit(1)
	}

	// Load locations from XML file
	locations, err := web.ReadLocationsFromXML(locationsPath)
	if err != nil {
		panic(fmt.Errorf("locations not loaded: %w", err))
	}

	// Initialize journal writer
	journalWriter := make(chan journal.JournalEntry, 1000)
	go func() {
		for entry := range journalWriter {
			path := "data"
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// Create empty directory
				if err = os.MkdirAll(path, os.ModePerm); err != nil {
					fmt.Fprintln(os.Stderr, "error while create directories: %w", err)
				}
			}

			if err = journal.WriteToJournalFile(path, &entry); err != nil {
				fmt.Fprintln(os.Stderr, "error while write journal file: %w", err)
			}
		}
	}()

	// Init session manager
	openSessions, sessionQueue, sessionIDs := web.RunSessionManager(journalWriter, 10)

	// Initialize token map
	// Tokens update automatically
	validTokens := web.RunTokenManager(&locations, time.Duration(expireDuration)*time.Second, loginURL, 10)

	runQRService(config, &locations, validTokens)
	runLoginService(config, loginURL, validTokens, openSessions, sessionIDs, sessionQueue)

	// Block main thread forever
	block := make(chan bool)
	block <- true
}

func runQRService(config config, locs *web.Locations, validTokens *web.ValidTokens) {
	mux := http.NewServeMux()

	// Set up web assets
	assets, _ := fs.Sub(content, "web/static")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loc := journal.Location(strings.ReplaceAll(r.URL.Path, "/", ""))

		// Return page for specific location
		if ok := locs.Contains(loc); ok {
			t := template.Must(template.ParseFS(content, "web/templates/qrservice/location.html"))
			t.Execute(w, loc)
			return
		}

		// Path is no location, return 404 not found page
		t := template.Must(template.ParseFS(content, "web/templates/qrservice/404.html"))
		t.Execute(w, locs)
	})

	mux.HandleFunc("/api/tokens", func(w http.ResponseWriter, r *http.Request) {
		getToken(w, r, validTokens)
	})

	go func() {
		log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf(":%v", config.qrPort), config.certPath, config.keyPath, mux))
	}()
}

func runLoginService(config config, url *url.URL, validTokens *web.ValidTokens, openSessions *web.OpenSessions, sessionIDs <-chan string, sessionQueue chan<- web.SessionQueueItem) {
	mux := http.NewServeMux()

	// Set up web assets
	assets, _ := fs.Sub(content, "web/static")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	mux.HandleFunc(url.Path, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
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

			var userCookie web.UserCookie

			// Check if cookie is valid
			decodedCookie, err := base64.StdEncoding.DecodeString(cookie.Value)
			if err != nil {
				// Cannot decode = Invalid Cookie
				onInvalidCookie(w, token)
				return
			}

			err = json.Unmarshal([]byte(decodedCookie), &userCookie)
			if err != nil {
				onInvalidCookie(w, token)
				return
			}

			hash, err := web.Hash(*userCookie.Person, privServerSecret)
			if err != nil || hash != userCookie.Hash {
				onInvalidCookie(w, token)
				return
			}
			// End of cookie validation check.

			// Cookie is available and valid
			// Search for UserSession with the same location
			// If found, perform logout
			if userSession, ok := openSessions.GetSessionForUser(hash); ok {
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
		case "POST":
			// Perform login
			// Check if data is complete and user have access to perform this login.
			if err := r.ParseForm(); err != nil {
				// Form cannot be parsed, Error
				w.WriteHeader(http.StatusInternalServerError)
				t := template.Must(template.ParseFS(content, "web/templates/loginservice/500.html"))
				t.Execute(w, err.Error())
				return
			}

			tokenID := r.FormValue("tokenId")
			location := journal.Location(r.FormValue("location"))
			firstName := r.FormValue("firstName")
			lastName := r.FormValue("lastName")
			street := r.FormValue("street")
			number := r.FormValue("number")
			zipCode := r.FormValue("zipCode")
			city := r.FormValue("city")
			if tokenID == "" || location == "" || firstName == "" || lastName == "" || street == "" || number == "" || zipCode == "" || city == "" {
				w.WriteHeader(http.StatusInternalServerError)
				t := template.Must(template.ParseFS(content, "web/templates/loginservice/500.html"))
				t.Execute(w, fmt.Errorf("form is not complete"))
				return
			}

			// Check if user can do perform a login.
			token, ok := validTokens.GetByID(tokenID)
			if !ok || token.Location != location {
				// Access denied.
				onAccessDenied(w)
				return
			}

			// Show login page and set new UserCookie
			person := journal.NewPerson(firstName, lastName, street, number, zipCode, city)
			userCookie, hash := web.CreateUserCookie(&person, privServerSecret)
			http.SetCookie(w, userCookie)

			// Token is valid, check if session exists
			// If UserSession exists, perform first logout and login afterwards
			if userSession, ok := openSessions.GetSessionForUser(hash); ok {
				sessionQueue <- web.CloseSession(timeutil.Now(), userSession, &person)
			}

			sessionQueue <- web.OpenSession(sessionIDs, timeutil.Now(), &person, location, privServerSecret)
			t := template.Must(template.ParseFS(content, "web/templates/loginservice/login.html"))
			t.Execute(w, location)
		}
	})

	mux.HandleFunc("/faq", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.ParseFS(content, "web/templates/loginservice/faq.html"))
		t.Execute(w, nil)
	})

	go func() {
		log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf(":%v", config.loginPort), config.certPath, config.keyPath, mux))
	}()
}
