package harness

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"time"
)

type mailpit struct {
	smtpAddr string
	apiAddr  string
	cleanup  func() error
}

func newMailpit(retry int) (*mailpit, error) {
	addrs, err := findOpenPorts(2)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	// detect if ../../../bin/tools/mailpit exists or ../../bin/tools/mailpit exists
	cmdpath := "../../bin/tools/mailpit"
	if _, err := os.Stat(cmdpath); err != nil {
		if _, err := os.Stat("../" + cmdpath); err == nil {
			cmdpath = "../" + cmdpath
		} else {
			return nil, fmt.Errorf("mailpit: %w", err)
		}
	}

	cmd := exec.Command(cmdpath, "-s", addrs[0], "-l", addrs[1])
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	to := time.NewTimer(5 * time.Second)
	defer to.Stop()
	for {
		// check if the process is still running
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			if retry > 0 && strings.Contains(output.String(), "address already in use") {
				// small random delay, in case of conflict
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				return newMailpit(retry - 1)
			}
			return nil, errors.New(output.String())
		}

		select {
		case <-to.C:
			return nil, fmt.Errorf("mailpit: timeout: %s", output.String())
		case <-t.C:
			if isListening(addrs[0]) && isListening(addrs[1]) {
				return &mailpit{
					smtpAddr: addrs[0],
					apiAddr:  addrs[1],
					cleanup:  cmd.Process.Kill,
				}, nil
			}
		}
	}
}

func (m *mailpit) Close() error { return m.cleanup() }

func (m *mailpit) UnreadMessages() ([]emailMessage, error) {
	resp, err := http.Get("http://" + m.apiAddr + "/api/v1/search?query=is:unread")
	if err != nil {
		return nil, fmt.Errorf("mailpit: search messages: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mailpit: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mailpit: search messages (bad response: %s): %s", resp.Status, string(data))
	}

	var body struct {
		Messages []struct {
			To      []mail.Address
			Snippet string
		}
	}
	if err = json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("mailpit: unmarshal response: %w\n%s", err, string(data))
	}

	var result []emailMessage
	for _, msg := range body.Messages {
		var addrs []string
		for _, p := range msg.To {
			addrs = append(addrs, p.Address)
		}
		result = append(result, emailMessage{address: addrs, body: msg.Snippet})
	}

	return result, nil
}

// ReadMessage will return true if an unread message was found and matched the keywords, marking it as read in the process.
func (m *mailpit) ReadMessage(to string, keywords ...string) (bool, error) {
	quotedKeywords := make([]string, len(keywords))
	for i, k := range keywords {
		quotedKeywords[i] = strconv.Quote(k)
	}
	query := fmt.Sprintf("is:unread to:%s %s", strconv.Quote(to), strings.Join(quotedKeywords, " "))

	resp, err := http.Get("http://" + m.apiAddr + "/api/v1/search?query=" + url.QueryEscape(query))
	if err != nil {
		return false, fmt.Errorf("mailpit: search messages: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("mailpit: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("mailpit: search messages (bad response: %s): %s", resp.Status, string(data))
	}

	var body struct {
		Messages []struct{ ID string }
	}
	if err = json.Unmarshal(data, &body); err != nil {
		return false, fmt.Errorf("mailpit: unmarshal response: %w\n%s", err, string(data))
	}

	if len(body.Messages) == 0 {
		return false, nil
	}

	var reqBody struct {
		IDs  []string
		Read bool
	}
	reqBody.IDs = append(reqBody.IDs, body.Messages[0].ID) // only read the first message
	reqBody.Read = true

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("mailpit: marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, "http://"+m.apiAddr+"/api/v1/messages", bytes.NewReader(reqData))
	if err != nil {
		return false, fmt.Errorf("mailpit: create request: %w", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("mailpit: mark message as read: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("mailpit: mark message as read (bad response: %s)", resp.Status)
	}

	return true, nil
}
