package secure

import "github.com/dateiexplorer/attendancelist/internal/journal"

type UserCookie struct {
	Person *journal.Person `json:"person"`
	Hash   string          `json:"hash"`
}
