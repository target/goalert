package apikey

import (
	"context"
	"net"

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
	params.IpAddress.IPNet.IP = net.ParseIP(ip)
	params.IpAddress.IPNet.Mask = net.CIDRMask(32, 32)
	if params.IpAddress.IPNet.IP != nil {
		params.IpAddress.Valid = true
	}
	return gadb.New(s.db).APIKeyRecordUsage(ctx, params)
}
