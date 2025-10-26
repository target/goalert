package app

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func withTimeout(t *testing.T, name string, fn func() error) error {
	t.Helper()
	errCh := make(chan error, 1)
	go func() {
		errCh <- fn()
	}()
	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()
	select {
	case err := <-errCh:
		return err
	case <-timeout.C:

	}

	t.Fatalf("%s: timeout", name)
	return nil // never runs
}

func TestMultiListener_Close(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer l.Close()

	m := newMultiListener(l)

	c, err := net.Dial("tcp", l.Addr().String())
	assert.NoError(t, err)
	defer c.Close()

	err = withTimeout(t, "close", m.Close)
	assert.NoError(t, err)
}

func TestMultiListener_Accept(t *testing.T) {
	t.Run("multiple listeners", func(t *testing.T) {
		l1, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NoError(t, err)
		defer l1.Close()

		l2, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NoError(t, err)
		defer l2.Close()

		m := newMultiListener(l1, l2)

		c1, err := net.Dial("tcp", l1.Addr().String())
		assert.NoError(t, err)
		defer c1.Close()

		ac1, err := m.Accept()
		assert.NoError(t, err)
		defer ac1.Close()

		assert.Equal(t, l1.Addr().String(), ac1.LocalAddr().String())
		assert.Equal(t, c1.LocalAddr().String(), ac1.RemoteAddr().String())

		c2, err := net.Dial("tcp", l2.Addr().String())
		assert.NoError(t, err)
		defer c2.Close()

		ac2, err := m.Accept()
		assert.NoError(t, err)
		defer ac2.Close()

		assert.Equal(t, l2.Addr().String(), ac2.LocalAddr().String())
		assert.Equal(t, c2.LocalAddr().String(), ac2.RemoteAddr().String())

		err = withTimeout(t, "close", m.Close)
		assert.NoError(t, err)

		err = withTimeout(t, "accept", func() error { _, err := m.Accept(); return err })
		assert.Error(t, err)
	})
	t.Run("return on accept pending", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NoError(t, err)
		defer l.Close()

		m := newMultiListener(l)

		go func() {
			time.Sleep(10 * time.Millisecond) // wait until Accept is called
			_ = m.Close()
		}()

		err = withTimeout(t, "accept", func() error { _, err := m.Accept(); return err })
		assert.Error(t, err)
	})
}
