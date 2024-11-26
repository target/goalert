package apikey

import (
	"context"
	"net"
	"net/netip"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation/validate"
)

// _updateLastUsed will record usage for the given API key ID, user agent, and IP address.
func (s *Store) _updateLastUsed(ctx context.Context, id uuid.UUID, ua, ip string) error {
	ua = validate.SanitizeText(ua, 1024)
	ip, _, _ = net.SplitHostPort(ip)
	ip = validate.SanitizeText(ip, 255)
	params := gadb.APIKeyRecordUsageParams{
		KeyID:     id,
		UserAgent: ua,
	}

	params.IpAddress, _ = netip.ParseAddr(ip) // best effort
	return gadb.NewCompat(s.db).APIKeyRecordUsage(ctx, params)
}
