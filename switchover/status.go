package switchover

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

	dbNext []byte
}

func ParseStatus(str string) (*Status, error) {
	v, err := url.ParseQuery(str)
	if err != nil {
		return nil, err
	}
	dbNext, err := base64.StdEncoding.DecodeString(v.Get("DBNext"))
	if err != nil {
		return nil, err
	}
	reqs, err := strconv.Atoi(v.Get("ActiveRequests"))
	if err != nil {
		return nil, err
	}

	s := &Status{
		NodeID: v.Get("NodeID"),
		State:  State(v.Get("State")),
		At:     time.Now(),
		dbNext: dbNext,

		ActiveRequests: reqs,
	}

	s.Offset, err = time.ParseDuration(v.Get("Offset"))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func stripAppName(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Del("application_name")
	u.RawQuery = q.Encode()
	return u.String()
}

// MatchDBNext will return true if the Status indicates a
// matching db-next-url.
func (s Status) MatchDBNext(dbNextURL string) bool {
	return hmac.Equal(s.dbNext, dbNext(s.NodeID, dbNextURL))
}
func dbNext(id, url string) []byte {
	return hmac.New(sha256.New, []byte(stripAppName(url))).Sum([]byte(id))
}
func (s Status) serialize() string {
	v := make(url.Values)
	v.Set("NodeID", s.NodeID)
	v.Set("State", string(s.State))
	v.Set("Offset", s.Offset.String())
	v.Set("DBNext", base64.StdEncoding.EncodeToString(s.dbNext))
	v.Set("ActiveRequests", strconv.Itoa(s.ActiveRequests))
	return v.Encode()
}
