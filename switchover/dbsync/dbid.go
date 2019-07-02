package dbsync

import (
	"crypto/rand"
	"encoding/hex"
)

func newDBID() string {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
