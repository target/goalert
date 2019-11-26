package sqlutil

import "github.com/jackc/pgx/v4"

func QuoteID(parts ...string) string {
	return pgx.Identifier(parts).Sanitize()
}
