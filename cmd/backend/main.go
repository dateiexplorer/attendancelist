package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/token"
)

func getAccessToken(w http.ResponseWriter, r *http.Request, validTokens *token.ValidTokens) {
	query := r.URL.Query()
	location := journal.Location(query.Get("location"))

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token, ok := validTokens.GetAccessTokenForLocation(location)
	if !ok {
		err := fmt.Errorf("no valid token found for this location")
		w.Write([]byte(fmt.Sprintf("{\"err\":\"%v\"}", err)))
		return
	}

	res, err := json.Marshal(token)
	if err != nil {
		return
	}

	w.Write(res)
}

func getLocations(w http.ResponseWriter, r *http.Request, locations *token.Locations) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(locations.Locations)
	if err != nil {
		return
	}

	w.Write(res)
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
	locations, err := token.ReadLocationsFromXML(path.Join(wd, "cmd", "backend", "locations.xml"))
	if err != nil {
		panic(fmt.Errorf("locations not loaded: %w", err))
	}

	// Initialize token map
	// Tokens update automatically
	validTokens := locations.GenerateAccessTokens(10, time.Duration(expireTime)*time.Second, loginURL, loginPort)

	http.HandleFunc("/tokens/access", func(w http.ResponseWriter, r *http.Request) {
		getAccessToken(w, r, validTokens)
	})

	http.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		getLocations(w, r, &locations)
	})

	fmt.Fprintf(log.Writer(), "Start backend server on %v:%v with params expire-time=%v, login-url=%v, login-port:%v\n", url, port, expireTime, loginURL, loginPort)
	log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf("%v:%v", url, port),
		path.Join(wd, "certs", "cert.pem"), path.Join(wd, "certs", "key.pem"), nil))
}
