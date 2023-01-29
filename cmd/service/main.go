// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"embed"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/web"
)

const privServerSecret = "privateServerSecret"

const journalStorePath = "data"
const maxConcurrentRequests = 1000
const tokenLength = 10

//go:embed web/*
var content embed.FS

func main() {
	// Configuration
	var loginURL, _ = url.Parse("https://localhost:4444/access")
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

	// Init journal writer
	// Journals written automatically
	journalWriter := runJournalWriter(maxConcurrentRequests, journalStorePath)

	// Init session manager
	openSessions, sessionQueue, sessionIDs := web.RunSessionManager(journalWriter, tokenLength)

	// Init token map
	// Tokens update automatically
	validTokens := web.RunTokenManager(&locations, time.Duration(expireDuration)*time.Second, loginURL, tokenLength)

	runQRService(config, &locations, validTokens)
	runLoginService(config, loginURL, validTokens, openSessions, sessionIDs, sessionQueue)

	// Block main thread forever
	block := make(chan bool)
	block <- true
}

func runJournalWriter(maxConcurrentRequests int, path string) chan<- journal.JournalEntry {
	journalWriter := make(chan journal.JournalEntry, maxConcurrentRequests)

	go func() {
		for entry := range journalWriter {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// Create empty directory
				if err = os.MkdirAll(path, os.ModePerm); err != nil {
					fmt.Fprintln(os.Stderr, "error while create directories: %w", err)
				}
			}

			if err := journal.WriteToJournalFile(path, &entry); err != nil {
				fmt.Fprintln(os.Stderr, "error while write journal file: %w", err)
			}
		}
	}()

	return journalWriter
}
