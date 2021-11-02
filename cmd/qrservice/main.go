// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"crypto/tls"
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

	flag.IntVar(&port, "port", 8080, "Port this service should running on")
	flag.IntVar(&backendPort, "backend-port", 4443, "Port the backend service should running on")
	flag.StringVar(&url, "url", "localhost", "URL under which this service is available without the https:// prefix")
	flag.StringVar(&backendURL, "backend-url", "localhost", "URL under which the backend service is available without the https:// prefix")

	flag.Parse()

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot get working directory: %w", err))
	}

	getLocationURL := fmt.Sprintf("https://%v:%v/locations", backendURL, backendPort)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(wd, "web", "static")))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loc := journal.Location(strings.ReplaceAll(r.URL.Path, "/", ""))

		res, err := http.Get(getLocationURL)
		if err != nil {
			return
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return
		}

		var locations secure.Locations
		err = json.Unmarshal(body, &locations)
		if err != nil {
			return
		}

		// Return page for specific location
		if ok := locations.Contains(loc); ok {
			t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "qrservice", "location.html")))

			t.Execute(w, loc)
			return
		}

		// Path is no location, return 404 not found page
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "qrservice", "notFound.html")))
		t.Execute(w, locations)
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

	fmt.Fprintf(log.Writer(), "Start qr service on %v:%v with params backend-url=%v, backend-port:%v\n", url, port, backendURL, backendPort)
	log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf("%v:%v", url, port),
		path.Join(wd, "certs", "cert.pem"), path.Join(wd, "certs", "key.pem"), nil))
}
