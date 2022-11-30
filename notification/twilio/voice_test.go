package twilio

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
)

func TestSpellNumber(t *testing.T) {
	// Test the spell number function
	assert.Equal(t, "1. 2. 3. 4. 5. 6", spellNumber(123456))
}

func TestSend(t *testing.T) {
	mockConfig := config.Config{}
	mockConfig.Twilio.Enable = true
	mockConfig.Twilio.VoiceName = "Polly.Joanna-Neural"
	mockConfig.Twilio.VoiceLanguage = "en-US"
	mockConfig.Twilio.FromNumber = "+16125551111"
	ctx := mockConfig.Context(context.Background())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)
		to := ""
		for k, v := range r.Form {
			switch k {
			case "To":
				assert.Equal(t, "+16125551234", v[0])
				to = v[0]
			case "From":
				assert.Equal(t, mockConfig.Twilio.FromNumber, v[0])
			case "Url":
				u, err := url.Parse(v[0])
				require.NoError(t, err)
				for queryKey, queryValue := range u.Query() {
					switch queryKey {
					case "msgBody":
						assert.Len(t, queryValue, 3)
						for i, msg := range queryValue {
							b64enc = base64.URLEncoding.WithPadding(base64.NoPadding)
							data, err := b64enc.DecodeString(msg)
							require.NoError(t, err)
							switch i {
							case 0:
								assert.Equal(t, "Hello! This is GoAlert with an alert notification.", string(data))
							case 1:
								assert.Equal(t, PAUSE, string(data))
							case 2:
								assert.Equal(t, "something happened.", string(data))
							}
						}
					}
				}
			}
		}

		callResp := Call{
			SID:       "CA0123456789abcdef",
			To:        to,
			From:      mockConfig.Twilio.FromNumber,
			Status:    CallStatus("queued"),
			Direction: "outbound",
		}
		resp, _ := json.Marshal(callResp)
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	}))
	defer svr.Close()

	cfg := &Config{BaseURL: svr.URL}

	voiceObj, err := NewVoice(ctx, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	result, err := voiceObj.Send(ctx, notification.Alert{
		Dest: notification.Dest{
			ID:    "1",
			Type:  notification.DestTypeVoice,
			Value: "+16125551234",
		},
		CallbackID: "test2",
		AlertID:    2,
		Summary:    "something happened",
		Details:    "something bad happened",
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, &notification.SentMessage{
		ExternalID:   "CA0123456789abcdef",
		State:        1,
		StateDetails: "queued",
		SrcValue:     mockConfig.Twilio.FromNumber,
	}, result)
}

func TestTwilioCallback(t *testing.T) {
	//todo see if we can get this working
	t.Skip()
	rec := httptest.NewRecorder()
	mockConfig := config.Config{}
	mockConfig.Twilio.Enable = true
	mockConfig.Twilio.VoiceName = "Polly.Joanna-Neural"
	mockConfig.Twilio.VoiceLanguage = "en-US"
	mockConfig.Twilio.FromNumber = "+16125551111"
	ctx := mockConfig.Context(context.Background())

	req := httptest.NewRequest("GET", "/api/v2/twilio/call?type=alert", nil)
	cfg := &Config{BaseURL: ""}
	voiceObj, err := NewVoice(ctx, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(voiceObj.ServeAlert)
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	//	assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
	//	data, err := io.ReadAll(resp.Body)
	//	assert.NoError(t, err)
	//	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
	//
	// <Response>
	//
	//	<Say>
	//		<prosody rate="slow">Hello</prosody>
	//	</Say>
	//	<Say>
	//		<prosody rate="slow">Goodbye.</prosody>
	//	</Say>
	//	<Hangup></Hangup>
	//
	// </Response>`, string(data))
}
