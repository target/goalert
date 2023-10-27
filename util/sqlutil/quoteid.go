package sqlutil

import "github.com/jackc/pgx/v5"

func QuoteID(parts ...string) string {
	return pgx.Identifier(parts).Sanitize()
}
