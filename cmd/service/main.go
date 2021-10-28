package main

import (
	"encoding/json"
	"fmt"
	"github.com/dateiexplorer/attendancelist/internal/locations"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)


func accessTokenDispenser(w http.ResponseWriter, r *http.Request, tokens *locations.ValidTokens) {
	q := r.URL.Query()
	loc := q.Get("loc")
	fmt.Println(loc)
	var token locations.AccessToken
	for _, val := range *tokens{
		if (*val.Location) == locations.Location(loc) && val.Valid ==2 {
			token = *val
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	resp, err := json.Marshal(token)
	if err != nil {
		fmt.Errorf("cant format AccessToken to json: %w",err)
	} else if token.Id != "" {
		w.Write(resp)
	}
}

func  serveLocations(w http.ResponseWriter, r *http.Request, loc locations.Locations) {
	resp, err := json.Marshal(loc.Locations)
	if err != nil {
		fmt.Errorf("cant format AccessToken to json: %w",err)
	} else{
		w.Write(resp)
	}
}

func main() {
	tokens := make(locations.ValidTokens, 1000)
	loc := locations.AccessTokenValidator(&tokens, path.Join("internal","locations","testdata","locations.xml"), time.Duration(5)*time.Second)
	wd, _ :=os.Getwd()
	fs:= http.FileServer(http.Dir(path.Join(wd, "web", "locSrc")))
	http.Handle("/", fs)

	http.HandleFunc("/loc", func(w http.ResponseWriter, r *http.Request) {
		serveLocations(w,r,loc)
	})

	http.HandleFunc("/newAccessTk", func(w http.ResponseWriter, r *http.Request) {
		accessTokenDispenser(w, r, &tokens)
	})
	log.Fatalln(http.ListenAndServeTLS(":4443",
		path.Join("cmd","service","cert.pem"), path.Join("cmd","service","key.pem"), nil))
}