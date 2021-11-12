package secure

import (
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"sync"

	"github.com/dateiexplorer/attendancelist/internal/journal"
)

type OpenSessions struct {
	sync.Map
}

type Session struct {
	ID       journal.SessionIdentifier
	UserHash string
	Location journal.Location
	Logintime timeutil.Timestamp
}
