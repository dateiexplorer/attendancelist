// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/dateiexplorer/dhbw-attendancelist/internal/journal"
	"github.com/dateiexplorer/dhbw-attendancelist/internal/web"
)

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
