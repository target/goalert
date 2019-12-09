package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortLongPath(t *testing.T) {
	check := func(long, short string) {
		t.Helper()

		if long != "" {
			assert.Equalf(t, short, ShortPath(long), `ShortPath("%s")`, long)
		}
		if short != "" {
			assert.Equalf(t, long, LongPath(short), `LongPath("%s")`, short)
		}
	}

	check("/unknown/path", "")
	check("", "/unknown/path")
	check("/alerts/123", "/a/ew")
	check("/alerts/123456", "/a/wMQH")
	check("/alerts/123456789", "/a/lZrvOg")
	check("/alerts/1234567890123", "/a/y4nsj.cj")
	check("/services/00000000-0000-0000-0000-000000000000/alerts", "/s/AAAAAAAAAAAAAAAAAAAAAA")
	check("/services/14ab7066-7371-4e06-ac59-ad488932fe36/alerts", "/s/FKtwZnNxTgasWa1IiTL-Ng")
}
