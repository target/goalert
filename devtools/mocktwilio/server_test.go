package mocktwilio_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/mocktwilio"
)

func TestServer(t *testing.T) {
	cfg := mocktwilio.Config{
		AccountSID: "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		AuthToken:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		OnError: func(ctx context.Context, err error) {
			t.Errorf("mocktwilio: error: %v", err)
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/voice", func(w http.ResponseWriter, req *http.Request) {
		t.Error("mocktwilio: unexpected voice request")
		w.WriteHeader(204)
	})
	mux.HandleFunc("/sms", func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "hello", req.FormValue("Body"))
		w.WriteHeader(204)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		t.Log(req.URL.Path)
		w.WriteHeader(204)
	})
	appHTTP := httptest.NewServer(mux)
	defer appHTTP.Close()

	srv := mocktwilio.NewServer(cfg)
	twHTTP := httptest.NewServer(srv)
	defer twHTTP.Close()

	appPhone := mocktwilio.Number{
		Number:          mocktwilio.NewPhoneNumber(),
		VoiceWebhookURL: appHTTP.URL + "/voice",
		SMSWebhookURL:   appHTTP.URL + "/sms",
	}
	err := srv.AddNumber(appPhone)
	require.NoError(t, err)

	// send device to app
	devNum := mocktwilio.NewPhoneNumber()
	_, err = srv.SendMessage(context.Background(), devNum, appPhone.Number, "hello")
	require.NoError(t, err)

	// send app to device
	v := make(url.Values)
	v.Set("From", appPhone.Number)
	v.Set("To", devNum)
	v.Set("Body", "world")
	v.Set("StatusCallback", appHTTP.URL+"/status")
	resp, err := http.PostForm(twHTTP.URL+"/2010-04-01/Accounts/"+cfg.AccountSID+"/Messages.json", v)
	require.NoError(t, err)

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Log("Response:", string(data))
	require.Equal(t, 201, resp.StatusCode)

	require.NoError(t, err)
	var res struct {
		SID string
	}
	err = json.Unmarshal(data, &res)
	require.NoError(t, err)

	msg := <-srv.Messages()
	assert.Equal(t, res.SID, msg.ID())

	resp, err = http.Get(twHTTP.URL + "/2010-04-01/Accounts/" + cfg.AccountSID + "/Messages/" + res.SID + ".json")
	require.NoError(t, err)

	data, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Log("Response:", string(data))
	require.Equal(t, 200, resp.StatusCode)

	err = srv.Close()
	require.NoError(t, err)
}
