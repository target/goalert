package sqldrv

import (
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/target/goalert/retry"
)

// SetConfigRetries will set the LookupFunc and DialFunc on the provided pgxpool.Config to use retry.Wrap.
func SetConfigRetries(cfg *pgxpool.Config) {
	cfg.ConnConfig.LookupFunc = retry.Wrap(net.DefaultResolver.LookupHost, 12, 100*time.Millisecond)

	var d net.Dialer
	cfg.ConnConfig.DialFunc = retry.Wrap2(d.DialContext, 12, 100*time.Millisecond)
}
