package mockslack

import (
	"regexp"
)

var (
	scopeRx           = regexp.MustCompile(`^[a-z]+(:[a-z]+)*$`)
	clientIDRx        = regexp.MustCompile(`^[0-9]{12}.[0-9]{12}$`)
	clientSecretRx    = regexp.MustCompile(`^[a-z0-9]{32}$`)
	userAccessTokenRx = regexp.MustCompile(`^xoxp-[0-9]{12}-[0-9]{12}-[0-9]{12}-[a-z0-9]{32}$`)
)
