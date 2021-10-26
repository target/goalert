package slack

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/target/goalert/config"
	"github.com/target/goalert/util"
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
		base, err := util.JoinURL(cs.cfg.BaseURL, "/api/")
		if err != nil {
			return fmt.Errorf("invalid Slack.BaseURL: %w", err)
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
