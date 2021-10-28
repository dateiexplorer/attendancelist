package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
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
	loc := journal.Location(query.Get("loc"))
	log.Println(loc)

	token, ok := validTokens.GetAccessTokenForLocation(loc)
	if !ok {
		log.Fatalln("not found")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(token)
	if err != nil {
		return
		// fmt.Errorf("cannot format AccessToken to json: %w", err)
	} else {
		w.Write(res)
	}
}

func serveLocations(w http.ResponseWriter, r *http.Request, loc token.Locations) {
	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(loc.Locations)
	if err != nil {
		fmt.Errorf("cant format AccessToken to json: %w", err)
	} else {
		w.Write(res)
	}
}

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

	fs := http.FileServer(http.Dir(path.Join(wd, "web", "locSrc")))
	http.Handle("/", fs)

	http.HandleFunc("/loc", func(w http.ResponseWriter, r *http.Request) {
		serveLocations(w, r, locations)
	})

	http.HandleFunc("/newAccessTk", func(w http.ResponseWriter, r *http.Request) {
		accessTokenDispenser(w, r, validTokens)
	})

	log.Fatalln(http.ListenAndServeTLS(":4443",
		path.Join(wd, "cmd", "service", "cert.pem"), path.Join(wd, "cmd", "service", "key.pem"), nil))

	// log.Fatalln(http.ListenAndServeTLS(":4443", path.Join("cert.pem"), path.Join("key.pem"), nil))
	// log.Fatalln(http.ListenAndServe(":4443", nil))

}
