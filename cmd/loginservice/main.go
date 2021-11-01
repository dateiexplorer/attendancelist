// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/secure"
)

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

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(wd, "web", "static")))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			// Perform login
			if err := r.ParseForm(); err != nil {
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
				// TODO: Form is not complete
				return
			}

			// Check if token is valid
			// Check for session
			// Do login (maybe logout)
		case "GET":
			// Login/Logout
		}

		query := r.URL.Query()
		if !query.Has("token") {
			return
		}

		tokenID := query.Get("token")

		// Check if token is valid
		res, err := http.Get(fmt.Sprintf("https://%v:%v/tokens/valid?id=%v", backendURL, backendPort, tokenID))
		if err != nil {
			return
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}

		var validTokenRes secure.ValidTokenResponse
		err = json.Unmarshal(body, &validTokenRes)

		if err != nil {
			return
		}

		if !validTokenRes.Valid {
			t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "forbidden.html")))
			t.Execute(w, nil)
			return
		}

		// End check if token is valid

		cookie, err := r.Cookie("user")
		var userCookie secure.UserCookie
		if err != nil {
			// Cookie not found

			// TODO: Set Cookie
			test := journal.NewPerson("Max", "Mustermann", "Musterstadt", "20", "74722", "Buchen")
			hash, _ := secure.Hash(test, "privServerSecret")
			testCookie := secure.UserCookie{Person: &test, Hash: hash}

			val, _ := json.Marshal(testCookie)
			value := base64.StdEncoding.EncodeToString(val)

			userCookie := &http.Cookie{
				Name:  "user",
				Value: value,
			}

			http.SetCookie(w, userCookie)
			// w.Write([]byte("Cookie set."))

			t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "login.html")))
			t.Execute(w, struct {
				Person *journal.Person
				Token  *secure.AccessToken
			}{nil, validTokenRes.Token})
			return
		}

		// Cookie is available
		decode, err := base64.StdEncoding.DecodeString(cookie.Value)

		err = json.Unmarshal([]byte(decode), &userCookie)
		fmt.Println(*userCookie.Person, err)
		if err != nil {
			return
		}

		// Check if data is valid
		serverHash, err := secure.Hash(*userCookie.Person, "privServerSecret")
		if serverHash != userCookie.Hash {
			w.Write([]byte("Invalid hash"))
			return
		}

		obj := struct {
			Person *journal.Person
			Token  *secure.AccessToken
		}{
			Person: userCookie.Person,
			Token:  validTokenRes.Token,
		}

		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "login.html")))
		t.Execute(w, obj)
	})

	// Proxy for backend to avoid cors issues.
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		url := strings.Replace(r.URL.RequestURI(), "/api/", "", 1)

		res, err := http.Get(fmt.Sprintf("https://%v:%v/%v", backendURL, backendPort, url))
		if err != nil {
			return
		}

		body, err := ioutil.ReadAll(res.Body)
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
