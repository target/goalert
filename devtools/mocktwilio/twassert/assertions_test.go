package twassert_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/mocktwilio"
	"github.com/target/goalert/devtools/mocktwilio/twassert"
)

func TestAssertSMS(t *testing.T) {
	s := mocktwilio.NewServer(mocktwilio.Config{
		AccountSID: "AC123",
		AuthToken:  "abc123",
	})
	srv := httptest.NewServer(s)
	defer srv.Close()

	s.AddUpdateNumber(mocktwilio.Number{Number: "+12345678901"})

	a := twassert.NewAssertions(t, twassert.Config{
		ServerAPI:      s,
		AppPhoneNumber: "+12345678901",
	})

	v := make(url.Values)
	v.Set("Body", "Hello, world!")
	v.Set("From", "+12345678901")
	v.Set("To", "+23456789012")
	resp, err := http.PostForm(srv.URL+"/2010-04-01/Accounts/AC123/Messages.json", v)
	require.NoError(t, err)
	if !assert.Equal(t, 201, resp.StatusCode) {
		data, _ := io.ReadAll(resp.Body)
		t.Log(string(data))
		return
	}

	a.Device("+23456789012").ExpectSMS("Hello, world!")
}
