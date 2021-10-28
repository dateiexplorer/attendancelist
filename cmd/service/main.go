package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/token"
)

var validTokens *token.ValidTokens

// func locationViewHandler(w http.ResponseWriter, r *http.Request) {
// 	/*title := r.URL.Path[len("/view/"):]
// 	p, _ := loadPage(title)
// 	t, _ := template.ParseFiles("view.html")
// 	t.Execute(w, p)*/
// }

func accessTokenDispenser(w http.ResponseWriter, r *http.Request) {
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

func main() {
	// Load locations from XML file
	locations, err := token.ReadLocationsFromXML(path.Join("locations.xml"))
	if err != nil {
		panic("locations not loaded")
	}

	// Initialize token map
	// Tokens update automatically
	validTokens = locations.GenerateAccessTokens(10, time.Duration(5)*time.Second, "localhost", 8081)

	http.HandleFunc("/newAccessTk", accessTokenDispenser)
	// log.Fatalln(http.ListenAndServeTLS(":4443", path.Join("cert.pem"), path.Join("key.pem"), nil))
	log.Fatalln(http.ListenAndServe(":4443", nil))
}
