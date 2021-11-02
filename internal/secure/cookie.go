package secure

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/dateiexplorer/attendancelist/internal/journal"
)

const UserCookieName = "user"

type UserCookie struct {
	Person *journal.Person `json:"person"`
	Hash   string          `json:"hash"`
}

func CreateUserCookie(person *journal.Person, privkey string) (*http.Cookie, string) {
	hash, _ := Hash(*person, privkey)
	cookie := UserCookie{Person: person, Hash: hash}

	json, _ := json.Marshal(cookie)
	encodedCookie := base64.StdEncoding.EncodeToString(json)

	return &http.Cookie{
		Name:  UserCookieName,
		Value: encodedCookie,
	}, hash
}
