package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func init() {
	register(func(ctx context.Context, connStr string) error {
		conn, err := pgx.Connect(ctx, connStr)
		if err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		defer conn.Close(ctx)

		err = conn.Ping(ctx)
		if err != nil {
			return fmt.Errorf("ping: %w", err)
		}

		return nil
	}, "postgres", "postgresql")
}
