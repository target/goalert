package timezone

import (
	_ "embed"
	"sort"
	"strings"
)

//go:embed zones.txt
var zoneData string

//go:embed aliases.txt
var aliasData string

var zones = strings.Split(strings.TrimSpace(zoneData), "\n")[1:] // skip header

var aliases = make(map[string]string)

func init() {
	sort.Strings(zones)

	for _, zone := range zones {
		aliases[zone] = zone
	}

	for _, line := range strings.Split(strings.TrimSpace(aliasData), "\n")[1:] { // skip header
		alias, zone, _ := strings.Cut(line, "=")
		aliases[alias] = zone
	}

	// check for unknown zones
	for _, zone := range aliases {
		if aliases[zone] != zone {
			panic("unknown zone: " + zone)
		}
	}
}

// Zones returns a list of all known time zones.
func Zones() []string { return zones }

// CanonicalZone returns the canonical name for the given zone.
func CanonicalZone(zone string) string { return aliases[zone] }
