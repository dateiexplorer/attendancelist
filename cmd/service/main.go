package main

import (
	"encoding/json"
	"fmt"
	"github.com/dateiexplorer/attendancelist/internal/locations"
	"log"
	"net/http"
	"path"
	"time"
)

var tokens = make(locations.ValidTokens, 1000)

func locationViewHandler(w http.ResponseWriter, r *http.Request) {
	/*title := r.URL.Path[len("/view/"):]
	p, _ := loadPage(title)
	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)*/
}

func accessTokenDispenser(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	loc := q.Get("loc")
	var token locations.AccessToken
	for key, val := range tokens{
		if (*val.Location) == locations.Location(loc) && val.Valid ==2 {
			token = *tokens[key]
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

func main() {
	locations.AccessTokenValidator(&tokens, path.Join("internal","locations","testdata","locations.xml"), time.Duration(5)*time.Second)

	http.HandleFunc("/newAccessTk", accessTokenDispenser)
	log.Fatalln(http.ListenAndServeTLS(":4443",
		path.Join("cmd","service","cert.pem"), path.Join("cmd","service","key.pem"), nil))
}