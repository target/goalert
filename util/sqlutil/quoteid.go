package sqlutil

import "github.com/jackc/pgx"

func QuoteID(parts ...string) string {
	return pgx.Identifier(parts).Sanitize()
}
