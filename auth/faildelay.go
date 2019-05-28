package auth

import (
	"context"
	cRand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"time"
)

var failRand *rand.Rand

const (
	delayMax = 10 * time.Millisecond
)

func init() {
	var seed int64
	err := binary.Read(cRand.Reader, binary.BigEndian, &seed)
	if err != nil {
		panic(err)
	}
	failRand = rand.New(rand.NewSource(seed))
}

// Delay will block for a random delay (or until the context is Done).
//
// It is useful in situations where there has been an auth failure.
func Delay(ctx context.Context) {
	dur := time.Duration(failRand.Int63n(int64(delayMax)))
	t := time.NewTicker(dur)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
