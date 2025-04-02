package main

import (
	"context"
	"net"
	"net/url"
	"time"
)

func init() {
	register(func(ctx context.Context, urlStr string) error {
		u, err := url.Parse(urlStr)
		if err != nil {
			return err
		}

		d, ok := ctx.Deadline()
		if !ok {
			panic("no deadline")
		}

		c, err := net.DialTimeout("tcp", u.Host, time.Until(d))
		if err != nil {
			return err
		}
		_ = c.Close()

		return nil
	}, "tcp")
}
