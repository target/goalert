package mockslack

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	idChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	botChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type idGen struct {
	ts  time.Time
	mx  sync.Mutex
	tok map[string]struct{}
}

func newIDGen() *idGen {
	return &idGen{
		ts:  time.Now(),
		tok: make(map[string]struct{}),
	}
}

func genID(s string, n int) string {
	buf := make([]byte, n)
	l := len(s)
	for i := range buf {
		buf[i] = s[rand.Intn(l)]
	}
	return string(buf)
}

func genTeamID() string {
	return "T" + genID(idChars, 8)
}

var hexSrc = rand.New(rand.NewSource(0))

func genHex(n int) string {
	buf := make([]byte, n)
	hexSrc.Read(buf)
	return hex.EncodeToString(buf)
}
func timeString(t time.Time) string { return fmt.Sprintf("%d", t.UnixNano()/1000000)[1:] }

func (gen *idGen) next(fn func() string) string {
	for {
		id := fn()
		gen.mx.Lock()
		if _, ok := gen.tok[id]; !ok {
			gen.tok[id] = struct{}{}
			gen.mx.Unlock()
			return id
		}
		gen.mx.Unlock()
	}
}

func (gen *idGen) ID(p string) string {
	return gen.next(func() string { return p + genID(idChars, 8) })
}
func (gen *idGen) UserID() string        { return gen.ID("W") }
func (gen *idGen) AppID() string         { return gen.ID("A") }
func (gen *idGen) ChannelID() string     { return gen.ID("D") }
func (gen *idGen) GroupID() string       { return gen.ID("G") }
func (gen *idGen) ClientSecret() string  { return gen.next(func() string { return genHex(16) }) }
func (gen *idGen) SigningSecret() string { return gen.next(func() string { return genHex(16) }) }

func (gen *idGen) TokenCode() string {
	return gen.next(func() string {
		return fmt.Sprintf("%s.%s.%s", timeString(gen.ts), timeString(time.Now().AddDate(2, 0, 0)), genHex(32))
	})
}

func (gen *idGen) ClientID() string {
	return gen.next(func() string { return timeString(gen.ts) + "." + timeString(time.Now()) })
}

func (gen *idGen) UserAccessToken() string {
	return gen.next(func() string {
		return fmt.Sprintf("xoxp-%s-%s-%s-%s", timeString(gen.ts), timeString(time.Now()), timeString(time.Now().AddDate(1, 0, 0)), genHex(16))
	})
}

func (gen *idGen) BotAccessToken() string {
	return gen.next(func() string {
		return fmt.Sprintf("xoxb-%s-%s-%s", timeString(gen.ts), timeString(time.Now()), genID(botChars, 24))
	})
}

func contains(strs []string, val string) bool {
	for _, s := range strs {
		if val == s {
			return true
		}
	}
	return false
}
