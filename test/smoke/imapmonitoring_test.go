package smoke

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestIMAPMonitoring tests that IMAP configuration and filter rules can be created and managed.
//
// - Create IMAP config with GraphQL
// - Create filter rules with GraphQL
// - Query and verify configuration
// - Update and delete filter rules
func TestIMAPMonitoring(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	`

	h := harness.NewHarness(t, sql, "imap-integration")
	defer h.Close()

	serviceID := h.UUID("sid")

	// Create IMAP configuration
	res := h.GraphQLQuery2(`
		mutation {
			createServiceIMAPConfig(input: {
				serviceID: "` + serviceID + `"
				enabled: true
				host: "imap.gmail.com"
				port: 993
				username: "test@example.com"
				useTLS: true
				mailbox: "INBOX"
				pollIntervalMinutes: 5
				markAsRead: false
				deleteAfter: false
				oauthClientID: "test-client-id"
				oauthClientSecret: "test-client-secret"
				oauthRefreshToken: "test-refresh-token"
				includeHeaders: false
				includeFrom: true
				includeTo: true
				includeSubject: true
				includeBody: true
			}) {
				serviceID
				enabled
				host
				port
				username
				useTLS
				mailbox
				pollIntervalMinutes
			}
		}
	`)
	require.Empty(t, res.Errors, "create IMAP config should not error")

	// Verify IMAP config was created
	var createResult struct {
		CreateServiceIMAPConfig struct {
			ServiceID           string
			Enabled             bool
			Host                string
			Port                int
			Username            string
			UseTLS              bool
			Mailbox             string
			PollIntervalMinutes int
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &createResult))
	assert.Equal(t, serviceID, createResult.CreateServiceIMAPConfig.ServiceID)
	assert.True(t, createResult.CreateServiceIMAPConfig.Enabled)
	assert.Equal(t, "imap.gmail.com", createResult.CreateServiceIMAPConfig.Host)
	assert.Equal(t, 993, createResult.CreateServiceIMAPConfig.Port)
	assert.Equal(t, "test@example.com", createResult.CreateServiceIMAPConfig.Username)
	assert.True(t, createResult.CreateServiceIMAPConfig.UseTLS)
	assert.Equal(t, "INBOX", createResult.CreateServiceIMAPConfig.Mailbox)
	assert.Equal(t, 5, createResult.CreateServiceIMAPConfig.PollIntervalMinutes)

	// Create filter rule for error emails
	res = h.GraphQLQuery2(`
		mutation {
			createIMAPFilterRule(input: {
				serviceID: "` + serviceID + `"
				name: "Error Alerts"
				subjectPattern: "ERROR"
				matchMode: contains
				excludeReplies: true
			}) {
				id
				name
				enabled
				subjectPattern
				matchMode
				excludeReplies
			}
		}
	`)
	require.Empty(t, res.Errors, "create filter rule should not error")

	var filterResult struct {
		CreateIMAPFilterRule struct {
			ID             string
			Name           string
			Enabled        bool
			SubjectPattern string
			MatchMode      string
			ExcludeReplies bool
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &filterResult))
	filterRuleID := filterResult.CreateIMAPFilterRule.ID
	assert.Equal(t, "Error Alerts", filterResult.CreateIMAPFilterRule.Name)
	assert.True(t, filterResult.CreateIMAPFilterRule.Enabled)
	assert.Equal(t, "ERROR", filterResult.CreateIMAPFilterRule.SubjectPattern)
	assert.Equal(t, "contains", filterResult.CreateIMAPFilterRule.MatchMode)
	assert.True(t, filterResult.CreateIMAPFilterRule.ExcludeReplies)

	// Create second filter rule with regex pattern
	res = h.GraphQLQuery2(`
		mutation {
			createIMAPFilterRule(input: {
				serviceID: "` + serviceID + `"
				name: "Critical Alerts"
				fromPattern: "alerts@.*\\.com"
				subjectPattern: "CRITICAL|URGENT"
				matchMode: regex
				excludeReplies: false
			}) {
				id
				name
			}
		}
	`)
	require.Empty(t, res.Errors, "create second filter rule should not error")

	// Query service to verify config and rules
	res = h.GraphQLQuery2(`
		query {
			service(id: "` + serviceID + `") {
				id
				name
				imapConfig {
					enabled
					host
					port
					username
					mailbox
					pollIntervalMinutes
					markAsRead
					deleteAfter
					includeHeaders
					includeFrom
					includeTo
					includeSubject
					includeBody
				}
				imapFilterRules {
					id
					name
					enabled
					fromPattern
					subjectPattern
					toPattern
					matchMode
					excludeReplies
				}
			}
		}
	`)
	require.Empty(t, res.Errors, "query service should not error")

	var serviceResult struct {
		Service struct {
			ID         string
			Name       string
			ImapConfig struct {
				Enabled             bool
				Host                string
				Port                int
				Username            string
				Mailbox             string
				PollIntervalMinutes int
				MarkAsRead          bool
				DeleteAfter         bool
				IncludeHeaders      bool
				IncludeFrom         bool
				IncludeTo           bool
				IncludeSubject      bool
				IncludeBody         bool
			}
			ImapFilterRules []struct {
				ID             string
				Name           string
				Enabled        bool
				FromPattern    *string
				SubjectPattern *string
				ToPattern      *string
				MatchMode      string
				ExcludeReplies bool
			}
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &serviceResult))

	// Verify IMAP config
	cfg := serviceResult.Service.ImapConfig
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "imap.gmail.com", cfg.Host)
	assert.Equal(t, 993, cfg.Port)
	assert.Equal(t, "test@example.com", cfg.Username)
	assert.Equal(t, "INBOX", cfg.Mailbox)
	assert.Equal(t, 5, cfg.PollIntervalMinutes)
	assert.False(t, cfg.MarkAsRead)
	assert.False(t, cfg.DeleteAfter)
	assert.False(t, cfg.IncludeHeaders)
	assert.True(t, cfg.IncludeFrom)
	assert.True(t, cfg.IncludeTo)
	assert.True(t, cfg.IncludeSubject)
	assert.True(t, cfg.IncludeBody)

	// Verify filter rules
	require.Len(t, serviceResult.Service.ImapFilterRules, 2)

	// Find rules by name (order not guaranteed)
	var errorRule, criticalRule *struct {
		ID             string
		Name           string
		Enabled        bool
		FromPattern    *string
		SubjectPattern *string
		ToPattern      *string
		MatchMode      string
		ExcludeReplies bool
	}

	for i := range serviceResult.Service.ImapFilterRules {
		switch serviceResult.Service.ImapFilterRules[i].Name {
		case "Error Alerts":
			errorRule = &serviceResult.Service.ImapFilterRules[i]
		case "Critical Alerts":
			criticalRule = &serviceResult.Service.ImapFilterRules[i]
		}
	}

	require.NotNil(t, errorRule, "Error Alerts rule should exist")
	require.NotNil(t, criticalRule, "Critical Alerts rule should exist")

	assert.True(t, errorRule.Enabled)
	assert.Nil(t, errorRule.FromPattern)
	assert.NotNil(t, errorRule.SubjectPattern)
	assert.Equal(t, "ERROR", *errorRule.SubjectPattern)
	assert.Nil(t, errorRule.ToPattern)
	assert.Equal(t, "contains", errorRule.MatchMode)
	assert.True(t, errorRule.ExcludeReplies)

	assert.True(t, criticalRule.Enabled)
	assert.NotNil(t, criticalRule.FromPattern)
	assert.Equal(t, "alerts@.*\\.com", *criticalRule.FromPattern)
	assert.NotNil(t, criticalRule.SubjectPattern)
	assert.Equal(t, "CRITICAL|URGENT", *criticalRule.SubjectPattern)
	assert.Equal(t, "regex", criticalRule.MatchMode)
	assert.False(t, criticalRule.ExcludeReplies)

	// Update filter rule
	res = h.GraphQLQuery2(`
		mutation {
			updateIMAPFilterRule(input: {
				id: "` + filterRuleID + `"
				enabled: false
				subjectPattern: "CRITICAL ERROR"
			})
		}
	`)
	require.Empty(t, res.Errors, "update filter rule should not error")

	// Verify update
	res = h.GraphQLQuery2(`
		query {
			service(id: "` + serviceID + `") {
				imapFilterRules {
					id
					name
					enabled
					subjectPattern
				}
			}
		}
	`)
	require.Empty(t, res.Errors)

	var updateResult struct {
		Service struct {
			ImapFilterRules []struct {
				ID             string
				Name           string
				Enabled        bool
				SubjectPattern *string
			}
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &updateResult))

	// Find the updated rule
	var updatedRule *struct {
		ID             string
		Name           string
		Enabled        bool
		SubjectPattern *string
	}
	for i := range updateResult.Service.ImapFilterRules {
		if updateResult.Service.ImapFilterRules[i].ID == filterRuleID {
			updatedRule = &updateResult.Service.ImapFilterRules[i]
			break
		}
	}
	require.NotNil(t, updatedRule, "updated rule should exist")
	assert.False(t, updatedRule.Enabled)
	assert.NotNil(t, updatedRule.SubjectPattern)
	assert.Equal(t, "CRITICAL ERROR", *updatedRule.SubjectPattern)

	// Delete filter rule
	res = h.GraphQLQuery2(`
		mutation {
			deleteIMAPFilterRule(id: "` + filterRuleID + `")
		}
	`)
	require.Empty(t, res.Errors, "delete filter rule should not error")

	// Verify deletion
	res = h.GraphQLQuery2(`
		query {
			service(id: "` + serviceID + `") {
				imapFilterRules {
					id
					name
				}
			}
		}
	`)
	require.Empty(t, res.Errors)

	var deleteResult struct {
		Service struct {
			ImapFilterRules []struct {
				ID   string
				Name string
			}
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &deleteResult))
	require.Len(t, deleteResult.Service.ImapFilterRules, 1, "should have 1 rule after deletion")
	assert.Equal(t, "Critical Alerts", deleteResult.Service.ImapFilterRules[0].Name)

	// Update IMAP config
	res = h.GraphQLQuery2(`
		mutation {
			updateServiceIMAPConfig(input: {
				serviceID: "` + serviceID + `"
				enabled: false
				pollIntervalMinutes: 10
				markAsRead: true
			})
		}
	`)
	require.Empty(t, res.Errors, "update IMAP config should not error")

	// Verify config update
	res = h.GraphQLQuery2(`
		query {
			service(id: "` + serviceID + `") {
				imapConfig {
					enabled
					pollIntervalMinutes
					markAsRead
					host
				}
			}
		}
	`)
	require.Empty(t, res.Errors)

	var configUpdateResult struct {
		Service struct {
			ImapConfig struct {
				Enabled             bool
				PollIntervalMinutes int
				MarkAsRead          bool
				Host                string
			}
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &configUpdateResult))
	assert.False(t, configUpdateResult.Service.ImapConfig.Enabled)
	assert.Equal(t, 10, configUpdateResult.Service.ImapConfig.PollIntervalMinutes)
	assert.True(t, configUpdateResult.Service.ImapConfig.MarkAsRead)
	assert.Equal(t, "imap.gmail.com", configUpdateResult.Service.ImapConfig.Host, "unchanged field should remain")

	// Delete IMAP config
	res = h.GraphQLQuery2(`
		mutation {
			deleteServiceIMAPConfig(serviceID: "` + serviceID + `")
		}
	`)
	require.Empty(t, res.Errors, "delete IMAP config should not error")

	// Verify config deletion
	res = h.GraphQLQuery2(`
		query {
			service(id: "` + serviceID + `") {
				imapConfig {
					enabled
				}
			}
		}
	`)
	require.Empty(t, res.Errors)

	var configDeleteResult struct {
		Service struct {
			ImapConfig *struct {
				Enabled bool
			}
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &configDeleteResult))
	assert.Nil(t, configDeleteResult.Service.ImapConfig, "IMAP config should be null after deletion")
}
