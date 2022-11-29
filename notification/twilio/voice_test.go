package twilio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
	ctx := mockConfig.Context(context.Background())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callResp := Call{
			SID:       "CA0123456789abcdef",
			To:        "+16125551234",
			From:      "+16125550000",
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
	result, err := voiceObj.Send(ctx, notification.Test{
		Dest: notification.Dest{
			ID:    "1",
			Type:  notification.DestTypeVoice,
			Value: "+16125551234",
		},
		CallbackID: "test2",
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, nil, result)
}
