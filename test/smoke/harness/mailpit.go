package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mailpit struct {
	t        *testing.T
	smtpAddr string
	apiAddr  string
	cleanup  func() error
}

func newMailpit(t *testing.T, retry int) (mp *mailpit) {
	t.Helper()

	addrs, err := findOpenPorts(2)
	require.NoError(t, err, "expected to find open ports for mailpit")

	var output bytes.Buffer
	// detect if ../../../bin/tools/mailpit exists or ../../bin/tools/mailpit exists
	cmdpath := "../../bin/tools/mailpit"
	if _, err := os.Stat(cmdpath); err != nil {
		_, err = os.Stat("../" + cmdpath)
		require.NoError(t, err, "expected to find mailpit binary")
		cmdpath = "../" + cmdpath
	}

	cmd := exec.Command(cmdpath, "-s", addrs[0], "-l", addrs[1])
	cmd.Stdout, cmd.Stderr = &output, &output
	require.NoError(t, cmd.Start(), "expected to start mailpit")

	require.Eventually(t, func() bool {
		// check if the process is still running
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			if retry > 0 && strings.Contains(output.String(), "address already in use") {
				// small random delay, in case of conflict
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				mp = newMailpit(t, retry-1)
				return true
			}
			t.Error(output.String())
			return false
		}
		return isListening(addrs[0]) && isListening(addrs[1])
	}, 5*time.Second, 100*time.Millisecond, "expected to find mailpit listening on ports")

	t.Cleanup(func() { _ = cmd.Process.Kill() })
	return &mailpit{
		t:        t,
		smtpAddr: addrs[0],
		apiAddr:  addrs[1],
		cleanup:  cmd.Process.Kill,
	}
}

func doJSON(t *testing.T, method, url string, reqBody, respBody any) {
	t.Helper()

	var data []byte
	if reqBody != nil {
		var err error
		data, err = json.Marshal(reqBody)
		require.NoError(t, err, "expected to marshal request")
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	require.NoError(t, err, "expected to create request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "expected to send request")
	data, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "expected to read response")
	require.Equalf(t, http.StatusOK, resp.StatusCode, "expected status OK; data=\n%s", string(data))
	if respBody == nil {
		return
	}
	require.NoErrorf(t, json.Unmarshal(data, respBody), "expected to unmarshal response from:\n%s", string(data))
}

func (m *mailpit) UnreadMessages() []emailMessage {
	m.t.Helper()

	var body struct {
		Messages []struct {
			To      []mail.Address
			Snippet string
		}
	}
	doJSON(m.t, "GET", "http://"+m.apiAddr+"/api/v1/search?query=is:unread", nil, &body)

	var result []emailMessage
	for _, msg := range body.Messages {
		var addrs []string
		for _, p := range msg.To {
			addrs = append(addrs, p.Address)
		}
		result = append(result, emailMessage{address: addrs, body: msg.Snippet})
	}

	return result
}

// ReadMessage will return true if an unread message was found and matched the keywords, marking it as read in the process.
func (m *mailpit) ReadMessage(to string, keywords ...string) bool {
	quotedKeywords := make([]string, len(keywords))
	for i, k := range keywords {
		quotedKeywords[i] = strconv.Quote(k)
	}
	query := fmt.Sprintf("is:unread to:%s %s", strconv.Quote(to), strings.Join(quotedKeywords, " "))

	var body struct {
		Messages []struct{ ID string }
	}
	doJSON(m.t, "GET", "http://"+m.apiAddr+"/api/v1/search?query="+url.QueryEscape(query), nil, &body)

	if len(body.Messages) == 0 {
		return false
	}

	var reqBody struct {
		IDs  []string
		Read bool
	}
	reqBody.IDs = append(reqBody.IDs, body.Messages[0].ID) // only read the first message
	reqBody.Read = true

	doJSON(m.t, http.MethodPut, "http://"+m.apiAddr+"/api/v1/messages", reqBody, nil)

	return true
}
