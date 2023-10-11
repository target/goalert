package sendit_test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/sendit"
)

func logCmd(t *testing.T, name string, arg ...string) *exec.Cmd {
	t.Helper()
	t.Log("exec:", name, strings.Join(arg, " "))
	return exec.Command(name, arg...)
}

// TestReadme is a test that runs the commands in the README and validates core functionality.
func TestReadme(t *testing.T) {
	const secret = "testing-secret"

	cmd := logCmd(t, "go", "run", "./cmd/sendit-token",
		"-secret", secret)
	tokenData, err := cmd.Output()
	require.NoError(t, err)
	token := strings.TrimSpace(string(tokenData))

	var c jwt.RegisteredClaims
	tok, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithAudience(sendit.TokenAudienceAuth), jwt.WithIssuer(sendit.TokenIssuer))
	require.NoError(t, err, "must be valid jwt")
	assert.True(t, tok.Valid, "token must be valid")

	assert.Equal(t, "sendit", c.Issuer)

	// start server
	cmd = logCmd(t, "go", "run", "./cmd/sendit-server",
		"-secret", secret,
		"-addr", "localhost:0")
	r, w := io.Pipe()
	cmd.Stderr = io.MultiWriter(w, os.Stdout)
	require.NoError(t, cmd.Start())
	defer mustExit(cmd.Process.Kill)

	rd := bufio.NewReader(r)
	s, err := rd.ReadString('\n')
	go func(rd *bufio.Reader) { _, _ = io.Copy(io.Discard, rd) }(rd)
	require.NoError(t, err)

	_, srvAddr, ok := strings.Cut(strings.TrimSpace(s), "Listening: ")
	require.True(t, ok, "must print Listening: <addr>")

	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server-prefix/test", r.URL.Path)
		_, _ = io.WriteString(w, "Hello, world!")
	}))
	defer testSrv.Close()

	srcURL := fmt.Sprintf("http://%s/server-prefix", srvAddr)
	// start client
	cmd = logCmd(t, "go", "run", "./cmd/sendit",
		"-token", token,
		srcURL,
		testSrv.URL,
	)
	r, w = io.Pipe()
	cmd.Stderr = w
	rd = bufio.NewReader(r)
	require.NoError(t, cmd.Start())
	defer mustExit(cmd.Process.Kill)

	for {
		s, err = rd.ReadString('\n')
		require.NoError(t, err)
		t.Logf("client: %s", s)
		if strings.Contains(s, "Ready") {
			break
		}
	}
	go func(rd *bufio.Reader) { _, _ = io.Copy(io.Discard, rd) }(rd)

	resp, err := http.Get(srcURL + "/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "Hello, world!", string(data))
}

// mustExit is a helper function to process error handling correctly on exit
func mustExit(fn func() error) {
	err := fn()
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		panic(err)
	}
}
