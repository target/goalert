package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func buildPause() (func(error), PauseResumer) {
	ch := make(chan error)

	return func(err error) {
			ch <- err
		},
		PauseResumerFunc(
			func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case err := <-ch:
					return err
				}
			},
			func(ctx context.Context) error {
				return nil
			},
		)
}

func TestMultiPauseResume(t *testing.T) {
	t.Run("simple success", func(t *testing.T) {
		to := time.NewTimer(time.Second)
		defer to.Stop()
		done1, pr1 := buildPause()
		done2, pr2 := buildPause()
		ctx := context.Background()
		errCh := make(chan error)
		go func() { errCh <- MultiPauseResume(pr1, pr2).Pause(ctx) }()

		done1(nil)
		done2(nil)

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("got %v; want nil", err)
			}
		case <-to.C:
			t.Fatal("never returned")
		}

	})
	t.Run("external cancellation", func(t *testing.T) {
		to := time.NewTimer(time.Second)
		defer to.Stop()

		_, pr1 := buildPause()
		_, pr2 := buildPause()
		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error)
		go func() { errCh <- MultiPauseResume(pr1, pr2).Pause(ctx) }()

		cancel()

		select {
		case err := <-errCh:
			if err == nil {
				t.Error("got nil; want err")
			}
		case <-to.C:
			t.Fatal("never returned")
		}
	})
	t.Run("external cancellation", func(t *testing.T) {
		to := time.NewTimer(time.Second)
		defer to.Stop()

		done1, pr1 := buildPause()
		_, pr2 := buildPause()
		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error)
		go func() { errCh <- MultiPauseResume(pr1, pr2).Pause(ctx) }()

		done1(nil)
		cancel()

		select {
		case err := <-errCh:
			if err == nil {
				t.Error("got nil; want err")
			}
		case <-to.C:
			t.Fatal("never returned")
		}
	})
	t.Run("external cancellation", func(t *testing.T) {
		to := time.NewTimer(time.Second)
		defer to.Stop()

		done1, pr1 := buildPause()
		_, pr2 := buildPause()
		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error)
		go func() { errCh <- MultiPauseResume(pr1, pr2).Pause(ctx) }()

		done1(errors.New("okay"))
		cancel()

		select {
		case err := <-errCh:
			if err == nil {
				t.Error("got nil; want err")
			}
		case <-to.C:
			t.Fatal("never returned")
		}
	})
}
