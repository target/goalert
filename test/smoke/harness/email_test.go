package harness

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindOpenPorts(t *testing.T) {
	// This test is more of a sanity check than anything else.
	//
	// Finding an open port is dependent on the system's network state, so it's
	// difficult to write a deterministic test for it. This test is just to
	// ensure that the function doesn't panic and returns a valid port.
	ports, err := findOpenPorts(1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ports[0], ":") {
		t.Fatalf("expected port to contain colon, got %s", ports[0])
	}

	// Ensure the port is actually open
	l, err := net.Listen("tcp", ports[0])
	require.NoError(t, err)
	defer l.Close()
}
