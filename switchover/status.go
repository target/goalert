package switchover

import (
	"net/url"
	"strconv"
	"time"
)

// Status represents the status of an individual node.
type Status struct {
	NodeID string
	State  State
	Offset time.Duration
	At     time.Time

	ActiveRequests int

	DBID     string
	DBNextID string
}

func ParseStatus(str string) (*Status, error) {
	v, err := url.ParseQuery(str)
	if err != nil {
		return nil, err
	}
	reqs, err := strconv.Atoi(v.Get("ActiveRequests"))
	if err != nil {
		return nil, err
	}

	s := &Status{
		NodeID:   v.Get("NodeID"),
		State:    State(v.Get("State")),
		At:       time.Now(),
		DBID:     v.Get("DBID"),
		DBNextID: v.Get("DBNextID"),

		ActiveRequests: reqs,
	}

	s.Offset, err = time.ParseDuration(v.Get("Offset"))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s Status) serialize() string {
	v := make(url.Values)
	v.Set("NodeID", s.NodeID)
	v.Set("State", string(s.State))
	v.Set("Offset", s.Offset.String())
	v.Set("DBID", s.DBID)
	v.Set("DBNextID", s.DBNextID)
	v.Set("ActiveRequests", strconv.Itoa(s.ActiveRequests))
	return v.Encode()
}
