// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/secure"
)

const privServerSecret = "privateServerSecret"

func isTokenValid(id string, backendURL string, backendPort int) (*secure.ValidTokenResponse, error) {
	res, err := http.Get(fmt.Sprintf("https://%v:%v/tokens/valid?id=%v", backendURL, backendPort, id))
	if err != nil {
		return nil, fmt.Errorf("get request failed: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read body: %w", err)
	}

	var validTokenRes secure.ValidTokenResponse
	json.Unmarshal(body, &validTokenRes)

	return &validTokenRes, nil
}

func onInvalidCookie(w http.ResponseWriter, wd string, validToken *secure.AccessToken) {
	t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "login.html")))
	t.Execute(w, struct {
		Person *journal.Person
		Token  *secure.AccessToken
	}{
		Person: nil,
		Token:  validToken,
	})
}

func main() {
	// Since this application uses a TLS secured connection with a self signed
	// certificate, communication between the services results in the following
	// error:
	// http: TLS handshake error from 127.0.0.1:*****: remote error: tls: bad certificate
	//
	// This happens because of self signed certificates not trusted.
	// It is highly recommended to comment out this line if using the
	// webservices in production.
	//
	// Credits for this solution to cyberdelia and Matthias and the following post:
	// https://stackoverflow.com/questions/12122159/how-to-do-a-https-request-with-bad-certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	var port, backendPort int
	var url, backendURL string

	flag.IntVar(&port, "port", 8081, "Port this service should running on")
	flag.IntVar(&backendPort, "backend-port", 4443, "Port the backend service should running on")
	flag.StringVar(&url, "url", "localhost", "URL under which this service is available without the https:// prefix")
	flag.StringVar(&backendURL, "backend-url", "localhost", "URL under which the backend service is available without the https:// prefix")

	flag.Parse()

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot get working directory: %w", err))
	}

	// Init journal writer
	journalWriter := make(chan journal.JournalEntry, 64)
	go func() {
		for entry := range journalWriter {
			body, err := json.Marshal(entry)
			if err != nil {
				fmt.Fprint(os.Stderr, fmt.Errorf("cannot marshal JournalEntry: %w", err))
				continue
			}

			res, err := http.Post(fmt.Sprintf("https://%v:%v/entries", backendURL, backendPort), "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Fprint(os.Stderr, fmt.Errorf("failed to post data to backend: %w", err))
				continue
			}

			res.Body.Close()
		}
	}()

	// Init session manager
	openSessions, sessionQueue, sessionIDs := secure.RunSessionManager(journalWriter, 10)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(wd, "web", "static")))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// Check if query param "token" is set.
			query := r.URL.Query()
			if !query.Has("token") {
				return
			}

			qTokenID := query.Get("token")
			validTokenRes, err := isTokenValid(qTokenID, backendURL, backendPort)
			if err != nil {
				// Error while reading = An error occured.
				fmt.Println("error while reading = An error occured.")
				return
			}

			if !validTokenRes.Valid {
				// Access denied.
				t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "forbidden.html")))
				t.Execute(w, nil)
				return
			}

			// Has Cookie?
			cookie, err := r.Cookie("user")
			if err != nil {
				fmt.Println("cookie not found")
				// No cookie found
				// User login for the first time or cookie is lost
				// In both cases the form must be filled in
				onInvalidCookie(w, wd, validTokenRes.Token)
				return
			}

			// Cookie is available
			var userCookie secure.UserCookie

			// Check if cookie is valid
			decodedCookie, err := base64.StdEncoding.DecodeString(cookie.Value)
			if err != nil {
				// Cannot decode = Invalid Cookie
				fmt.Println("cannot decode = Invalid Cookie")
				onInvalidCookie(w, wd, validTokenRes.Token)
				return
			}

			err = json.Unmarshal([]byte(decodedCookie), &userCookie)
			if err != nil {
				// Cannot unmarhsal = Invalid Cookie
				fmt.Println("cannot unmarshal = Invalid Cookie")
				onInvalidCookie(w, wd, validTokenRes.Token)
				return
			}

			// Check if data is valid (hash)
			hash, err := secure.Hash(*userCookie.Person, privServerSecret)
			if err != nil || hash != userCookie.Hash {
				// Hash is invalid = Invalid Cookie
				fmt.Println("hash is invalid = Invalid Cookie")
				onInvalidCookie(w, wd, validTokenRes.Token)
				return
			}

			// Cookie is available and valid
			// Search for UserSession with the same location
			// If found, perform logout
			if userSession, ok := openSessions.GetSessionForUser(hash); ok {
				if userSession.Location == validTokenRes.Token.Location {
					// Same location => perform logout
					sessionQueue <- secure.CloseSession(userSession, userCookie.Person)
					t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "logout.html")))
					t.Execute(w, validTokenRes.Token.Location)
					return
				}
			}

			// If the location isn't the same as the location in the UserSession
			// or UserSession doesn't exists, show filled login form
			t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "login.html")))
			t.Execute(w, struct {
				Person *journal.Person
				Token  *secure.AccessToken
			}{
				Person: userCookie.Person,
				Token:  validTokenRes.Token,
			})
		case "POST":
			// Perform login
			if err := r.ParseForm(); err != nil {
				// Form cannot be parsed, Error
				return
			}

			fTokenID := r.FormValue("tokenId")
			fLocation := r.FormValue("location")
			fFirstName := r.FormValue("firstName")
			fLastName := r.FormValue("lastName")
			fStreet := r.FormValue("street")
			fNumber := r.FormValue("number")
			fZipCode := r.FormValue("zipCode")
			fCity := r.FormValue("city")

			if fTokenID == "" || fLocation == "" || fFirstName == "" || fLastName == "" || fStreet == "" || fNumber == "" || fZipCode == "" || fCity == "" {
				// Form is incomplete, Error
				return
			}

			validTokenRes, err := isTokenValid(fTokenID, backendURL, backendPort)
			if err != nil {
				// Error while readding = An error occured.
				return
			}

			location := journal.Location(fLocation)
			if !validTokenRes.Valid || validTokenRes.Token.Location != location {
				// Access denied.
				t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "forbidden.html")))
				t.Execute(w, nil)
				return
			}

			// Show login page and set new UserCookie
			person := journal.NewPerson(fFirstName, fLastName, fStreet, fNumber, fZipCode, fCity)

			userCookie, hash := secure.CreateUserCookie(&person, privServerSecret)
			http.SetCookie(w, userCookie)

			// Token is valid, Check if session exists
			// If UserSession exists, perform first logout and login afterwards
			if userSession, ok := openSessions.GetSessionForUser(hash); ok {
				sessionQueue <- secure.CloseSession(userSession, &person)
			}

			sessionQueue <- secure.OpenSession(sessionIDs, &person, location, privServerSecret)
			t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "success.html")))
			t.Execute(w, location)
		}
	})

	// Proxy for backend to avoid cors issues.
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		url := strings.Replace(r.URL.RequestURI(), "/api/", "", 1)

		res, err := http.Get(fmt.Sprintf("https://%v:%v/%v", backendURL, backendPort, url))
		if err != nil {
			return
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(body)
	})

	fmt.Fprintf(log.Writer(), "Start login service on %v:%v with params backend-url=%v, backend-port:%v\n", url, port, backendURL, backendPort)
	log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf("%v:%v", url, port),
		path.Join(wd, "certs", "cert.pem"), path.Join(wd, "certs", "key.pem"), nil))
}
