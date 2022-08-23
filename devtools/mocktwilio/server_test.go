package mocktwilio_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/mocktwilio"
)

func TestServer(t *testing.T) {
	t.Run("SMS", func(t *testing.T) {
		cfg := mocktwilio.Config{
			AccountSID: "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			AuthToken:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			OnError: func(ctx context.Context, err error) {
				t.Errorf("mocktwilio: error: %v", err)
			},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			t.Errorf("mocktwilio: unexpected request to %s", req.URL.String())
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
		srv := mocktwilio.NewServer(cfg)
		twHTTP := httptest.NewServer(srv)
		defer appHTTP.Close()
		defer srv.Close()
		defer twHTTP.Close()

		appPhone := mocktwilio.Number{
			Number:          mocktwilio.NewPhoneNumber(),
			VoiceWebhookURL: appHTTP.URL + "/voice",
			SMSWebhookURL:   appHTTP.URL + "/sms",
		}
		require.NoError(t, srv.AddNumber(appPhone))

		// send device to app
		devNum := mocktwilio.NewPhoneNumber()
		_, err := srv.SendMessage(context.Background(), devNum, appPhone.Number, "hello")
		require.NoError(t, err)

		// send app to device
		v := make(url.Values)
		v.Set("From", appPhone.Number)
		v.Set("To", devNum)
		v.Set("Body", "world")
		v.Set("StatusCallback", appHTTP.URL+"/status")
		resp, err := http.PostForm(twHTTP.URL+"/2010-04-01/Accounts/"+cfg.AccountSID+"/Messages.json", v)
		require.NoError(t, err)
		var msgStatus struct {
			SID    string
			Status string
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&msgStatus))
		assert.Equal(t, "queued", msgStatus.Status)

		msg := <-srv.Messages()
		assert.Equal(t, msg.ID(), msgStatus.SID)

		resp, err = http.Get(twHTTP.URL + "/2010-04-01/Accounts/" + cfg.AccountSID + "/Messages/" + msg.ID() + ".json")
		require.NoError(t, err)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&msgStatus))
		require.Equal(t, "sending", msgStatus.Status)

		require.NoError(t, msg.SetStatus(context.Background(), mocktwilio.MessageDelivered))

		resp, err = http.Get(twHTTP.URL + "/2010-04-01/Accounts/" + cfg.AccountSID + "/Messages/" + msg.ID() + ".json")
		require.NoError(t, err)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&msgStatus))
		require.Equal(t, "delivered", msgStatus.Status)

		t.Fail()
		require.NoError(t, srv.Close())
	})

	t.Run("Voice", func(t *testing.T) {
		return
		cfg := mocktwilio.Config{
			AccountSID: "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			AuthToken:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			OnError: func(ctx context.Context, err error) {
				t.Errorf("mocktwilio: error: %v", err)
			},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			t.Errorf("mocktwilio: unexpected request to %s", req.URL.String())
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
		srv := mocktwilio.NewServer(cfg)
		twHTTP := httptest.NewServer(srv)
		defer appHTTP.Close()
		defer srv.Close()
		defer twHTTP.Close()

		appPhone := mocktwilio.Number{
			Number:          mocktwilio.NewPhoneNumber(),
			VoiceWebhookURL: appHTTP.URL + "/voice",
			SMSWebhookURL:   appHTTP.URL + "/sms",
		}
		require.NoError(t, srv.AddNumber(appPhone))
	})
}
