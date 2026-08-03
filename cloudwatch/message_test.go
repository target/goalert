package cloudwatch

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The signing scheme requires alphabetical field order, and the tables are
// hand-written literals rather than sorted at runtime.
func TestSignedFieldsAreSorted(t *testing.T) {
	for msgType, fields := range signedFields {
		assert.True(t, sort.StringsAreSorted(fields), "%s field list is not sorted: %v", msgType, fields)
	}
}

func strPtr(s string) *string { return &s }

func TestCanonicalString(t *testing.T) {
	tests := []struct {
		name    string
		env     envelope
		want    string
		wantErr bool
	}{
		{
			name: "notification with subject",
			env: envelope{
				Type:      typeNotification,
				MessageID: "mid",
				TopicARN:  "arn",
				Subject:   strPtr("subj"),
				Message:   "msg",
				Timestamp: "ts",
			},
			want: "Message\nmsg\nMessageId\nmid\nSubject\nsubj\nTimestamp\nts\nTopicArn\narn\nType\nNotification\n",
		},
		{
			// Absent Subject omits the whole block, it does not contribute an
			// empty value.
			name: "notification without subject",
			env: envelope{
				Type:      typeNotification,
				MessageID: "mid",
				TopicARN:  "arn",
				Message:   "msg",
				Timestamp: "ts",
			},
			want: "Message\nmsg\nMessageId\nmid\nTimestamp\nts\nTopicArn\narn\nType\nNotification\n",
		},
		{
			// A forger cannot fake absence with an empty string: this produces a
			// different canonical string, so verification fails.
			name: "notification with empty subject differs from absent",
			env: envelope{
				Type:      typeNotification,
				MessageID: "mid",
				TopicARN:  "arn",
				Subject:   strPtr(""),
				Message:   "msg",
				Timestamp: "ts",
			},
			want: "Message\nmsg\nMessageId\nmid\nSubject\n\nTimestamp\nts\nTopicArn\narn\nType\nNotification\n",
		},
		{
			name: "subscription confirmation",
			env: envelope{
				Type:         typeSubscriptionConfirmation,
				MessageID:    "mid",
				TopicARN:     "arn",
				Message:      "msg",
				Timestamp:    "ts",
				SubscribeURL: "https://sub",
				Token:        "tok",
			},
			want: "Message\nmsg\nMessageId\nmid\nSubscribeURL\nhttps://sub\nTimestamp\nts\nToken\ntok\nTopicArn\narn\nType\nSubscriptionConfirmation\n",
		},
		{
			name: "unsubscribe confirmation",
			env: envelope{
				Type:         typeUnsubscribeConfirmation,
				MessageID:    "mid",
				TopicARN:     "arn",
				Message:      "msg",
				Timestamp:    "ts",
				SubscribeURL: "https://sub",
				Token:        "tok",
			},
			want: "Message\nmsg\nMessageId\nmid\nSubscribeURL\nhttps://sub\nTimestamp\nts\nToken\ntok\nTopicArn\narn\nType\nUnsubscribeConfirmation\n",
		},
		{
			// Subject is not signed for confirmations, so it must not appear even
			// when present.
			name: "confirmation ignores subject",
			env: envelope{
				Type:         typeSubscriptionConfirmation,
				MessageID:    "mid",
				TopicARN:     "arn",
				Subject:      strPtr("ignored"),
				Message:      "msg",
				Timestamp:    "ts",
				SubscribeURL: "https://sub",
				Token:        "tok",
			},
			want: "Message\nmsg\nMessageId\nmid\nSubscribeURL\nhttps://sub\nTimestamp\nts\nToken\ntok\nTopicArn\narn\nType\nSubscriptionConfirmation\n",
		},
		{name: "unknown type errors", env: envelope{Type: "Bogus"}, wantErr: true},
		{name: "empty type errors", env: envelope{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := canonicalString(&tt.env)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
