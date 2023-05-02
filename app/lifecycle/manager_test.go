package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_PauseingShutdown(t *testing.T) {

	_, pr := buildPause()
	ran := make(chan struct{})
	run := func(ctx context.Context) error { <-ctx.Done(); close(ran); return ctx.Err() }
	shut := func(ctx context.Context) error { return nil }
	mgr := NewManager(run, shut)
	require.NoError(t, mgr.SetPauseResumer(pr))

	go func() { assert.ErrorIs(t, mgr.Run(context.Background()), context.Canceled) }()

	var err error
	errCh := make(chan error)
	pauseErr := make(chan error)

	tc := time.NewTimer(time.Second)
	defer tc.Stop()

	go func() { pauseErr <- mgr.Pause(context.Background()) }()
	tc.Reset(time.Second)
	select {
	case <-mgr.PauseWait():
	case <-tc.C:
		t.Fatal("pause didn't start")
	}
	// done(nil)

	go func() { errCh <- mgr.Shutdown(context.Background()) }()

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("shutdown never finished")
	case err = <-errCh:
	}
	if err != nil {
		t.Fatalf("shutdown error: got %v; want nil", err)
	}

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("run never got canceled")
	case <-ran:
	}

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("pause never finished")
	case <-pauseErr:
	}

}

func TestManager_PauseShutdown(t *testing.T) {
	done, pr := buildPause()
	ran := make(chan struct{})
	run := func(ctx context.Context) error { <-ctx.Done(); close(ran); return ctx.Err() }
	shut := func(ctx context.Context) error { return nil }
	mgr := NewManager(run, shut)
	require.NoError(t, mgr.SetPauseResumer(pr))

	go func() { assert.ErrorIs(t, mgr.Run(context.Background()), context.Canceled) }()

	var err error
	errCh := make(chan error)
	go func() { errCh <- mgr.Pause(context.Background()) }()
	done(nil)

	tc := time.NewTimer(time.Second)
	defer tc.Stop()
	select {
	case <-tc.C:
		t.Fatal("pause never finished")
	case err = <-errCh:
	}
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	go func() { errCh <- mgr.Shutdown(context.Background()) }()

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("shutdown never finished")
	case err = <-errCh:
	}
	if err != nil {
		t.Fatalf("shutdown error: got %v; want nil", err)
	}

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("run never got canceled")
	case <-ran:
	}

}

func TestManager_PauseResume(t *testing.T) {
	done, pr := buildPause()
	run := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	shut := func(ctx context.Context) error { return nil }
	mgr := NewManager(run, shut)
	require.NoError(t, mgr.SetPauseResumer(pr))

	go func() { assert.ErrorIs(t, mgr.Run(context.Background()), context.Canceled) }()

	var err error
	errCh := make(chan error)
	go func() { errCh <- mgr.Pause(context.Background()) }()
	done(nil)

	tc := time.NewTimer(time.Second)
	defer tc.Stop()
	select {
	case <-tc.C:
		t.Fatal("pause never finished")
	case err = <-errCh:
	}
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}

	go func() { errCh <- mgr.Resume(context.Background()) }()

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("resume never finished")
	case err = <-errCh:
	}
	if err != nil {
		t.Fatalf("resume error: got %v; want nil", err)
	}

}

func TestManager_PauseingResume(t *testing.T) {

	_, pr := buildPause()
	ran := make(chan struct{})
	run := func(ctx context.Context) error { <-ctx.Done(); close(ran); return ctx.Err() }
	shut := func(ctx context.Context) error { return nil }
	mgr := NewManager(run, shut)
	require.NoError(t, mgr.SetPauseResumer(pr))

	go func() { assert.ErrorIs(t, mgr.Run(context.Background()), context.Canceled) }()

	var err error
	errCh := make(chan error)
	pauseErr := make(chan error)

	tc := time.NewTimer(time.Second)
	defer tc.Stop()

	go func() { pauseErr <- mgr.Pause(context.Background()) }()
	tc.Reset(time.Second)
	select {
	case <-mgr.PauseWait():
	case <-tc.C:
		t.Fatal("pause didn't start")
	}
	// done(nil)

	go func() { errCh <- mgr.Resume(context.Background()) }()

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("resume never finished")
	case err = <-errCh:
	}
	if err != nil {
		t.Fatalf("resume error: got %v; want nil", err)
	}

	tc.Reset(time.Second)
	select {
	case <-tc.C:
		t.Fatal("pause never finished")
	case <-pauseErr:
	}

}
