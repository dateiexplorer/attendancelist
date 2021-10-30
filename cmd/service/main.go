package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/token"
)

// func locationViewHandler(w http.ResponseWriter, r *http.Request) {
// 	/*title := r.URL.Path[len("/view/"):]
// 	p, _ := loadPage(title)
// 	t, _ := template.ParseFiles("view.html")
// 	t.Execute(w, p)*/
// }

func accessTokenDispenser(w http.ResponseWriter, r *http.Request, validTokens *token.ValidTokens) {
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
		// fmt.Errorf("cannot format AccessToken to json: %w", err)
	} else {
		w.Write(res)
	}
}

// func serveLocations(w http.ResponseWriter, r *http.Request, loc token.Locations) {
// 	w.Header().Set("Content-Type", "application/json")

// 	res, err := json.Marshal(loc.Locations)
// 	if err != nil {
// 		fmt.Errorf("cant format AccessToken to json: %w", err)
// 	} else {
// 		w.Write(res)
// 	}
// }

func main() {
	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot get working directory: %w", err))
	}

	// Load locations from XML file
	locations, err := token.ReadLocationsFromXML(path.Join(wd, "cmd", "service", "locations.xml"))
	if err != nil {
		panic(fmt.Errorf("locations not loaded: %w", err))
	}

	// Initialize token map
	// Tokens update automatically
	validTokens := locations.GenerateAccessTokens(10, time.Duration(5)*time.Second, "localhost", 8081)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(wd, "web", "static")))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loc := journal.Location(strings.ReplaceAll(r.URL.Path, "/", ""))

		// Return page for specific location
		if ok := locations.Contains(loc); ok {
			t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "qrcode.html")))

			t.Execute(w, loc)
			return
		}

		// Path is no location, return 404 not found page
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "notFound.html")))
		t.Execute(w, locations)
	})

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	split := strings.Split(r.URL.Path, "/")

	// 	if len(split) != 2 {
	// 		temp := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "locations.html")))
	// 		temp.Execute(w, locations)
	// 		return
	// 	}

	// 	location := split[1]
	// 	if location == "" {
	// 		// temp := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "locations.html")))
	// 		// temp.Execute(w, locations)
	// 		w.WriteHeader(http.StatusNotFound)
	// 		fmt.Fprint(w, "custom 404")
	// 		return
	// 	}

	// 	valid := false
	// 	for _, l := range locations.Locations {
	// 		if l == journal.Location(location) {
	// 			valid = true
	// 			break
	// 		}
	// 	}

	// 	if !valid {
	// 		temp := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "locations.html")))
	// 		temp.Execute(w, locations)
	// 		return
	// 	}

	// 	temp := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "qrcode.html")))
	// 	temp.Execute(w, location)
	// })

	// http.HandleFunc("/loc", func(w http.ResponseWriter, r *http.Request) {
	// 	serveLocations(w, r, locations)
	// })

	http.HandleFunc("/api/token/get", func(w http.ResponseWriter, r *http.Request) {
		accessTokenDispenser(w, r, validTokens)
	})

	log.Fatalln(http.ListenAndServeTLS(":4443",
		path.Join(wd, "cmd", "service", "cert.pem"), path.Join(wd, "cmd", "service", "key.pem"), nil))

	// log.Fatalln(http.ListenAndServeTLS(":4443", path.Join("cert.pem"), path.Join("key.pem"), nil))
	// log.Fatalln(http.ListenAndServe(":4443", nil))

}
