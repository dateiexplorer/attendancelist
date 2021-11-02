// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/secure"
)

func getAccessToken(w http.ResponseWriter, r *http.Request, validTokens *secure.ValidTokens) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
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

	if token, ok := validTokens.GetAccessTokenForLocation(location); ok {
		res, _ := json.Marshal(token)
		w.Write(res)
		return
	}

	err := fmt.Errorf("no valid token found for this location")
	w.Write([]byte(fmt.Sprintf("{\"err\":\"%v\"}", err)))
}

func isValidToken(w http.ResponseWriter, r *http.Request, validTokens *secure.ValidTokens) {
	query := r.URL.Query()
	id := query.Get("id")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var data secure.ValidTokenResponse

	if value, ok := validTokens.Load(id); ok {
		t, _ := value.(*secure.AccessToken)
		data = secure.ValidTokenResponse{Valid: true, Token: t}
	} else {
		data = secure.ValidTokenResponse{Valid: false, Token: nil}
	}

	res, _ := json.Marshal(data)
	w.Write(res)
}

func getLocations(w http.ResponseWriter, r *http.Request, locations *secure.Locations) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(locations.Locations)
	if err != nil {
		return
	}

	w.Write(res)
}

func addJournalEntry(w http.ResponseWriter, r *http.Request, journalWriter chan<- journal.JournalEntry) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	var entry journal.JournalEntry
	err = json.Unmarshal(data, &entry)
	if err != nil {
		return
	}

	journalWriter <- entry
}

func main() {
	var port, expireTime, loginPort int
	var url, loginURL string

	flag.IntVar(&port, "port", 4443, "Port this service should running on")
	flag.IntVar(&loginPort, "login-port", 8081, "Port the login service should running on")
	flag.IntVar(&expireTime, "expire-time", 30, "Expire time for access tokens in seconds")
	flag.StringVar(&url, "url", "localhost", "URL under which this service is available without the https:// prefix")
	flag.StringVar(&loginURL, "login-url", "localhost", "URL under which the login service is available without the https:// prefix")

	flag.Parse()

	if expireTime <= 0 {
		fmt.Fprintln(os.Stderr, "The expire time for access token must be greater than 0.")
		os.Exit(1)
	}

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot get working directory: %w", err))
	}

	// Load locations from XML file
	locations, err := secure.ReadLocationsFromXML(path.Join(wd, "cmd", "backend", "locations.xml"))
	if err != nil {
		panic(fmt.Errorf("locations not loaded: %w", err))
	}

	// Initialize token map
	// Tokens update automatically
	validTokens := locations.GenerateAccessTokens(10, time.Duration(expireTime)*time.Second, loginURL, loginPort)

	// Initialize journal writer
	journalWriter := make(chan journal.JournalEntry, 64)
	go func() {
		for entry := range journalWriter {
			journal.WriteToJournalFile(path.Join(wd, "data"), entry)
		}
	}()

	http.HandleFunc("/tokens", func(w http.ResponseWriter, r *http.Request) {
		getAccessToken(w, r, validTokens)
	})

	http.HandleFunc("/tokens/valid", func(w http.ResponseWriter, r *http.Request) {
		isValidToken(w, r, validTokens)
	})

	http.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		getLocations(w, r, &locations)
	})

	http.HandleFunc("/entries", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			addJournalEntry(w, r, journalWriter)
		}
	})

	fmt.Fprintf(log.Writer(), "Start backend server on %v:%v with params expire-time=%v, login-url=%v, login-port:%v\n", url, port, expireTime, loginURL, loginPort)
	log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf("%v:%v", url, port),
		path.Join(wd, "certs", "cert.pem"), path.Join(wd, "certs", "key.pem"), nil))
}
