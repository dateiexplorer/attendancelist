// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"html/template"
	"io/ioutil"
	"log"
	"time"

	//"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/secure"
)


func cookieExists(r *http.Request) (*http.Cookie, bool) {
	cookie, err := r.Cookie("user")
	if err != nil {
		return nil, false
	}
	return cookie, true
}

func validateToken(tokenID string, backendURL string, backendPort int) (secure.ValidTokenResponse, bool) {

	res, err := http.Get(fmt.Sprintf("https://%v:%v/tokens/valid?id=%v", backendURL, backendPort, tokenID))
	if err != nil {
		fmt.Errorf("there must be something with the backend server please check: %w", err)
		return secure.ValidTokenResponse{}, false
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("can't read body of response: %w", err)
		return secure.ValidTokenResponse{}, false
	}

	var validTokenRes secure.ValidTokenResponse
	err = json.Unmarshal(body, &validTokenRes)

	if err != nil {
		fmt.Errorf("can't unmarshal body of response: %w", err)
		return secure.ValidTokenResponse{}, false
	}

	if !validTokenRes.Valid {
		return secure.ValidTokenResponse{}, false
	}

	return validTokenRes, true
}

type accessResponse struct {
	Person *journal.Person
	Token  *secure.AccessToken
}

func access(w http.ResponseWriter, r *http.Request, wd string, backendURL string, backendPort int, sessions *secure.OpenSessions) {
	query := r.URL.Query()
	if !query.Has("token") {
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "forbidden.html")))
		t.Execute(w, nil)
		return
	}

	tokenID := query.Get("token")
	validTokenRes, isValid := validateToken(tokenID, backendURL, backendPort)
	// Check if token is valid
	if !isValid {
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "forbidden.html")))
		t.Execute(w, nil)
		return
	}

	cookie, exists := cookieExists(r)

	if !exists {
		obj := accessResponse{
			Person: nil,
			Token:  validTokenRes.Token,
		}
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "login.html")))
		t.Execute(w, obj)
		return
	}

	var userCookie secure.UserCookie
	decode, errEnc := base64.StdEncoding.DecodeString(cookie.Value)
	if errEnc != nil {
		fmt.Errorf("can't decode cookie: %w", errEnc)
	}
	errUn := json.Unmarshal([]byte(decode), &userCookie)
	fmt.Println(*userCookie.Person, errUn)
	if errUn != nil {
		fmt.Errorf("can't unmarshal cookie: %w", errUn)
	}

	obj := accessResponse {
		Person: userCookie.Person,
		Token:  validTokenRes.Token,
	}

	isLoggedIn := false
	sessions.Range(func(key, value interface{}) bool {
		val := value.(secure.Session)
		if val.UserHash == cookie.Value && val.Location == validTokenRes.Token.Location {
			isLoggedIn = true
			return false
		}
		return true
	})

	if isLoggedIn {
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "logout.html")))
		t.Execute(w, obj)
	} else {

		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "login.html")))
		t.Execute(w, obj)
	}

	//has cookie
}

func generateCookie(person journal.Person) *http.Cookie {
	hash, err := secure.Hash(person, "privServerSecret")
	if err != nil {
		return &http.Cookie{}
	}
	cookie := secure.UserCookie{Person: &person, Hash: hash}

	val, err := json.Marshal(cookie)
	if err != nil {
		return &http.Cookie{}
	}
	value := base64.StdEncoding.EncodeToString(val)

	userCookie := &http.Cookie{
		Name:  "user",
		Value: value,
	}

	return userCookie
}

func validateFormData(r *http.Request, backendURL string, backendPort int) (journal.Person, secure.ValidTokenResponse, bool) {
	if err := r.ParseForm(); err != nil {
		fmt.Errorf("cant parse form: %w", err)
		return journal.Person{}, secure.ValidTokenResponse{}, false
	}

	fTokenID := r.FormValue("tokenId")
	fLocation := r.FormValue("location")
	fFirstName := r.FormValue("firstName")
	fLastName := r.FormValue("lastName")
	fStreet := r.FormValue("street")
	fNumber := r.FormValue("number")
	fZipCode := r.FormValue("zipCode")
	fCity := r.FormValue("city")

	// Check if token is valid
	accessToken, isValid := validateToken(fTokenID, backendURL, backendPort)
	if !isValid {
		return journal.Person{}, secure.ValidTokenResponse{}, false
	}

	if fTokenID == "" || fLocation == "" || fFirstName == "" || fLastName == "" || fStreet == "" || fNumber == "" || fZipCode == "" || fCity == "" {
		return journal.Person{}, secure.ValidTokenResponse{}, false
	}

	person := journal.NewPerson(fFirstName, fLastName, fStreet, fNumber, fZipCode, fCity)

	return person, accessToken, true
}

type successResponse struct {
	Action string
	Location journal.Location
}

func login(w http.ResponseWriter, r *http.Request, backendURL string, backendPort int, wd string, sessions *secure.OpenSessions, sessionIds *<-chan string) {

	person, validToken, isValidFormData := validateFormData(r, backendURL, backendPort)

	if !isValidFormData {
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "forbidden.html")))
		t.Execute(w, nil)
		return
	}

	cookie, exists := cookieExists(r)

	if !exists {
		cookie = generateCookie(person)
		http.SetCookie(w, cookie)
	}

	//nicht gleiche location, weil wenn an neuem platz angemeldet die location ja nicht mit übergeben
	//oder machen wir da neue seite?
	sessions.Range(func(key, value interface{}) bool {
		val := value.(secure.Session)
		if val.UserHash == cookie.Value {
			sessions.Delete(key)
			return false
		}
		return true
	})

	sessionId := <-*sessionIds
	session := secure.Session{
		ID:       journal.SessionIdentifier(sessionId),
		UserHash: cookie.Value,
		Location: validToken.Token.Location,
	}
	sessions.Store(sessionId, session)

	newEntry := journal.NewJournalEntry(timeutil.Timestamp{Time: time.Now()}, session.ID, journal.Event(0), validToken.Token.Location, person)
	err := journal.WriteToJournalFile(path.Join(wd, "internal", "journal", "testdata"), newEntry)

	if err != nil {
		return
	}

	obj := successResponse{
		Action: "Login",
		Location: validToken.Token.Location,
	}

	t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "successful.html")))
	t.Execute(w, obj)
	return
}

func logout(w http.ResponseWriter, r *http.Request, sessions *secure.OpenSessions, wd string, backendURL string, backendPort int) {
	person, validToken, isValidFormData := validateFormData(r, backendURL, backendPort)

	if !isValidFormData {
		t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "forbidden.html")))
		t.Execute(w, nil)
		return
	}

	cookie := generateCookie(person)

	var session secure.Session

	sessions.Range(func(key, value interface{}) bool {
		val := value.(secure.Session)
		if val.UserHash == cookie.Value && val.Location == validToken.Token.Location {
			session = val
			sessions.Delete(key)
			return false
		}
		return true
	})

	if (session == secure.Session{}) {
		return
	}

	newEntry := journal.NewJournalEntry(timeutil.Timestamp{Time: time.Now()}, session.ID, journal.Event(1), validToken.Token.Location, person)
	journal.WriteToJournalFile(path.Join(wd, "internal", "journal", "testdata"), newEntry)

	obj := successResponse {
		Action: "Logout",
		Location: validToken.Token.Location,
	}

	t := template.Must(template.ParseFiles(path.Join(wd, "web", "templates", "loginservice", "successful.html")))
	t.Execute(w, obj)
	return

}

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

	flag.IntVar(&port, "port", 8081, "Port this service should running on")
	flag.IntVar(&backendPort, "backend-port", 4443, "Port the backend service should running on")
	flag.StringVar(&url, "url", "localhost", "URL under which this service is available without the https:// prefix")
	flag.StringVar(&backendURL, "backend-url", "localhost", "URL under which the backend service is available without the https:// prefix")

	flag.Parse()

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot get working directory: %w", err))
	}

	sessions := new(secure.OpenSessions)

	sessionIds := secure.RandIDGenerator(16, 10000)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(wd, "web", "static")))))

	http.HandleFunc("/access", func(w http.ResponseWriter, r *http.Request) {
		access(w, r, wd, backendURL, backendPort, sessions)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logout(w, r, sessions, wd, backendURL, backendPort)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r, backendURL, backendPort, wd, sessions, &sessionIds)

		// Check if token is valid

		// Check for session

		// Do login (maybe logout)

		// End check if token is valid


		// Cookie is available


		// Check if data is valid
		/*serverHash, err := secure.Hash(*userCookie.Person, "privServerSecret")
		if err != nil {
			fmt.Errorf("can't hash cookie: %w", err)
		}

		if serverHash != userCookie.Hash {
			w.Write([]byte("Invalid hash"))
			return
		}
		*/

	})

	// Proxy for backend to avoid cors issues.
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		url := strings.Replace(r.URL.RequestURI(), "/api/", "", 1)

		res, err := http.Get(fmt.Sprintf("https://%v:%v/%v", backendURL, backendPort, url))
		if err != nil {
			return
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(body)
	})

	fmt.Fprintf(log.Writer(), "Start login service on %v:%v with params backend-url=%v, backend-port:%v\n", url, port, backendURL, backendPort)
	log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf("%v:%v", url, port),
		path.Join(wd, "certs", "cert.pem"), path.Join(wd, "certs", "key.pem"), nil))
}
