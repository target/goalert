package slack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/target/goalert/config"
)

func waitContext(ctx context.Context, delay time.Duration) error {
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// withClient is a wrapper for slack.Client that adds retry logic.
func (cs *ChannelSender) withClient(ctx context.Context, withFn func(*slack.Client) error) error {
	opts := []slack.Option{
		slack.OptionHTTPClient(http.DefaultClient),
	}

	cfg := config.FromContext(ctx)
	if cs.cfg.BaseURL != "" {
		base := cs.cfg.BaseURL
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		opts = append(opts, slack.OptionAPIURL(base))
	}

	cli := slack.New(cfg.Slack.AccessToken, opts...)

	var err error
	for i := 0; i < 3; i++ {
		err = withFn(cli)

		var rateErr *slack.RateLimitedError
		if errors.As(err, &rateErr) && rateErr.RetryAfter > 0 {
			err = waitContext(ctx, rateErr.RetryAfter)
			if err != nil {
				return err
			}

			// retry
			continue
		}

		return err
	}

	return fmt.Errorf("failed after 3 attempts: %w", err)
}
