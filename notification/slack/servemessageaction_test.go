package slack

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
)

func TestValidateRequestSignature(t *testing.T) {
	// Values pulled directly from: https://api.slack.com/authentication/verifying-requests-from-slack
	var cfg config.Config
	cfg.Slack.SigningSecret = "8f742231b10e8888abcd99yyyzzz85a5"

	req, err := http.NewRequestWithContext(cfg.Context(context.Background()), "POST", "http://example.com", strings.NewReader("token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c"))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Slack-Request-Timestamp", "1531420618")
	req.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")

	err = validateRequestSignature(time.Unix(1531420618, 0), req)
	assert.NoError(t, err)

	req, err = http.NewRequestWithContext(cfg.Context(context.Background()), "POST", "http://example.com", strings.NewReader("token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c"))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Slack-Request-Timestamp", "15314206189") // changed timestamp
	req.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")

	// different timestamp should invalidate the signature
	err = validateRequestSignature(time.Unix(1531420618, 0), req)
	assert.True(t, permission.IsUnauthorized(err), "expected unauthorized error, got: %v", err)
}

// Test cases for the Show Details feature
func TestParseDetailsFromActionValue(t *testing.T) {
	testCases := []struct {
		name           string
		actionValue    string
		expectError    bool
		expectID       int
		expectSummary  string
		expectDetails  string
		expectCallback string
	}{
		{
			name:           "Valid JSON payload",
			actionValue:    `{"type":"show_details","alert_id":123,"summary":"Test Alert","details":"This is a test alert","callback_id":"alert:123:unacknowledged"}`,
			expectError:    false,
			expectID:       123,
			expectSummary:  "Test Alert",
			expectDetails:  "This is a test alert",
			expectCallback: "alert:123:unacknowledged",
		},
		{
			name:           "Valid JSON with escaped details",
			actionValue:    `{"type":"show_details","alert_id":456,"summary":"Multi-line Alert","details":"Line 1\nLine 2\nLine 3","callback_id":"alert:456:acknowledged"}`,
			expectError:    false,
			expectID:       456,
			expectSummary:  "Multi-line Alert",
			expectDetails:  "Line 1\nLine 2\nLine 3",
			expectCallback: "alert:456:acknowledged",
		},
		{
			name:        "Invalid JSON",
			actionValue: `{"type":"show_details","alert_id":123,"summary":}`,
			expectError: true,
		},
		{
			name:           "Missing required field",
			actionValue:    `{"type":"show_details","summary":"Test Alert","details":"Test details","callback_id":"alert:123:unacknowledged"}`,
			expectError:    false, // Function doesn't strictly validate - it uses default values
			expectID:       0,     // Default value for missing alert_id
			expectSummary:  "Test Alert",
			expectDetails:  "Test details",
			expectCallback: "alert:123:unacknowledged",
		},
		{
			name:        "Wrong type",
			actionValue: `{"type":"other_action","alert_id":123,"summary":"Test Alert","details":"Test details","callback_id":"alert:123:unacknowledged"}`,
			expectError: true,
		},
		{
			name:        "Empty string",
			actionValue: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			details, err := parseDetailsFromActionValue(tc.actionValue)

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
				assert.Nil(t, details, "Expected nil details on error")
			} else {
				assert.NoError(t, err, "Expected no error")
				require.NotNil(t, details, "Expected non-nil details")

				assert.Equal(t, tc.expectID, details.ID)
				assert.Equal(t, tc.expectSummary, details.Summary)
				assert.Equal(t, tc.expectDetails, details.Details)
				assert.Equal(t, tc.expectCallback, details.CallbackID)
			}
		})
	}
}

func TestBuildExpandedAlertBlocks(t *testing.T) {
	ctx := context.Background()
	var cfg config.Config
	ctx = cfg.Context(ctx)

	sender := &ChannelSender{}

	testCases := []struct {
		name          string
		alert         *AlertDetails
		expectBlocks  int // minimum expected number of blocks
		expectButtons []string
	}{
		{
			name: "Unacknowledged alert with details",
			alert: &AlertDetails{
				ID:         123,
				Summary:    "Test Alert",
				Details:    "This is the full alert details that was previously collapsed",
				Status:     "unacknowledged",
				CallbackID: "alert:123:unacknowledged",
			},
			expectBlocks:  3, // title, details, action buttons
			expectButtons: []string{"Acknowledge", "Close"},
		},
		{
			name: "Acknowledged alert with details",
			alert: &AlertDetails{
				ID:         456,
				Summary:    "Acknowledged Alert",
				Details:    "Multi-line details\nWith several lines\nOf important information",
				Status:     "acknowledged",
				CallbackID: "alert:456:acknowledged",
			},
			expectBlocks:  3,
			expectButtons: []string{"Close"},
		},
		{
			name: "Alert without details",
			alert: &AlertDetails{
				ID:         789,
				Summary:    "Simple Alert",
				Details:    "",
				Status:     "unacknowledged",
				CallbackID: "alert:789:unacknowledged",
			},
			expectBlocks:  2, // title, action buttons (no details section)
			expectButtons: []string{"Acknowledge", "Close"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blocks := sender.buildExpandedAlertBlocks(ctx, tc.alert)

			assert.GreaterOrEqual(t, len(blocks), tc.expectBlocks, "Expected at least %d blocks", tc.expectBlocks)

			// Find and verify action buttons
			var foundButtons []string
			for _, block := range blocks {
				if actionBlock, ok := block.(*slack.ActionBlock); ok {
					for _, element := range actionBlock.Elements.ElementSet {
						if button, ok := element.(*slack.ButtonBlockElement); ok {
							foundButtons = append(foundButtons, button.Text.Text)
						}
					}
				}
			}

			assert.ElementsMatch(t, tc.expectButtons, foundButtons, "Expected buttons don't match")

			// Verify details section presence
			detailsFound := false
			for _, block := range blocks {
				if sectionBlock, ok := block.(*slack.SectionBlock); ok {
					if sectionBlock.Text != nil && strings.Contains(sectionBlock.Text.Text, "*Details:*") {
						detailsFound = true
						if tc.alert.Details != "" {
							assert.Contains(t, sectionBlock.Text.Text, tc.alert.Details)
						}
					}
				}
			}

			if tc.alert.Details != "" {
				assert.True(t, detailsFound, "Expected to find details section")
			}

			// Verify status context shows "expanded"
			statusFound := false
			for _, block := range blocks {
				if contextBlock, ok := block.(*slack.ContextBlock); ok {
					for _, element := range contextBlock.ContextElements.Elements {
						if textObj, ok := element.(*slack.TextBlockObject); ok {
							if strings.Contains(textObj.Text, "Details expanded") {
								statusFound = true
								break
							}
						}
					}
				}
			}

			assert.True(t, statusFound, "Expected to find 'Details expanded' status")
		})
	}
}
