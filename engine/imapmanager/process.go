package imapmanager

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/gadb"
)


// processMessageForService processes a single email message for a specific service.
func (db *DB) processMessageForService(ctx context.Context, c *client.Client, msg *imap.Message, svc IMAPServiceConfig) error {
	// Parse the email
	email := parseEnvelope(msg)
	if email == nil {
		db.logger.Warn("IMAP failed to parse message envelope", "service_id", svc.ServiceID)
		return nil
	}

	if email.MessageID == "" {
		db.logger.Warn("IMAP message has no Message-ID, skipping", "service_id", svc.ServiceID)
		return nil
	}

	// Check if already processed (within a transaction)
	var alreadyProcessed bool
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		queries := gadb.New(tx)

		processed, err := queries.IMAPMessageProcessed(ctx, email.MessageID)
		if err != nil {
			return fmt.Errorf("check processed: %w", err)
		}

		alreadyProcessed = processed
		return nil
	})

	if err != nil {
		return err
	}

	if alreadyProcessed {
		db.logger.Debug("IMAP message already processed", "service_id", svc.ServiceID, "message_id", email.MessageID)
		// Still mark as read/delete if configured, but don't create alerts
		db.updateMessageFlagsForService(c, msg.SeqNum, svc)
		return nil
	}

	// Check if email matches any filter rules for this service
	var matched bool
	err = db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		queries := gadb.New(tx)

		// Get filter rules for this service
		rules, err := queries.IMAPFilterRulesForService(ctx, svc.ServiceID)
		if err != nil {
			return fmt.Errorf("get filter rules: %w", err)
		}

		// Check if any rule matches (OR logic across rules)
		for _, rule := range rules {
			ruleMatched, err := matchesFilter(email, rule)
			if err != nil {
				db.logger.Error("filter matching error", "service_id", svc.ServiceID, "rule_id", rule.ID, "error", err)
				continue
			}

			if ruleMatched {
				matched = true
				break
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("check filter rules: %w", err)
	}

	if !matched {
		db.logger.Debug("IMAP no matching filter rules", "service_id", svc.ServiceID, "message_id", email.MessageID)
		// Mark as processed even if no matches (prevent reprocessing)
		err = db.markProcessed(ctx, email.MessageID)
		if err != nil {
			return err
		}
		db.updateMessageFlagsForService(c, msg.SeqNum, svc)
		return nil
	}

	// Create alert for this service
	err = db.createAlert(ctx, email, svc)
	if err != nil {
		db.logger.Error("IMAP failed to create alert", "service_id", svc.ServiceID, "error", err)
		return err
	}

	db.logger.Info("IMAP created alert", "service_id", svc.ServiceID, "message_id", email.MessageID)

	// Mark message as processed
	err = db.markProcessed(ctx, email.MessageID)
	if err != nil {
		return err
	}

	// Update IMAP flags (mark as read, delete if configured)
	db.updateMessageFlagsForService(c, msg.SeqNum, svc)

	return nil
}

// createAlert creates a GoAlert alert from an email based on service configuration.
func (db *DB) createAlert(ctx context.Context, email *emailMessage, svc IMAPServiceConfig) error {
	// Sanitize and truncate fields
	summary := sanitizeText(email.Subject, 1024)

	// Build details based on configuration
	var detailsParts []string

	if svc.IncludeHeaders {
		// Include basic email headers
		if email.Date != "" {
			detailsParts = append(detailsParts, fmt.Sprintf("Date: %s", email.Date))
		}
		if email.MessageID != "" {
			detailsParts = append(detailsParts, fmt.Sprintf("Message-ID: %s", email.MessageID))
		}
	}

	if svc.IncludeFrom && email.From != "" {
		detailsParts = append(detailsParts, fmt.Sprintf("From: %s", email.From))
	}

	if svc.IncludeTo && email.To != "" {
		detailsParts = append(detailsParts, fmt.Sprintf("To: %s", email.To))
	}

	if svc.IncludeSubject && email.Subject != "" {
		detailsParts = append(detailsParts, fmt.Sprintf("Subject: %s", email.Subject))
	}

	if svc.IncludeBody && email.Body != "" {
		if len(detailsParts) > 0 {
			detailsParts = append(detailsParts, "") // Empty line before body
		}
		detailsParts = append(detailsParts, sanitizeText(email.Body, 6000))
	}

	details := strings.Join(detailsParts, "\n")

	// If no parts were included, at least include "From" as fallback
	if details == "" {
		details = fmt.Sprintf("From: %s", email.From)
	}

	// Create alert
	a := &alert.Alert{
		Source:    alert.SourceEmail,
		Summary:   summary,
		Details:   details,
		ServiceID: svc.ServiceID.String(),
		Dedup:     alert.NewUserDedup(email.MessageID),
		Status:    alert.StatusTriggered,
	}

	_, _, err := db.alertStore.CreateOrUpdate(ctx, a)
	if err != nil {
		return fmt.Errorf("create alert: %w", err)
	}

	return nil
}

// markProcessed marks a message as processed in the database.
func (db *DB) markProcessed(ctx context.Context, messageID string) error {
	return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		queries := gadb.New(tx)
		return queries.IMAPMarkMessageProcessed(ctx, messageID)
	})
}


// updateMessageFlagsForService updates IMAP message flags for a specific service.
func (db *DB) updateMessageFlagsForService(c *client.Client, seqNum uint32, svc IMAPServiceConfig) {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNum)

	// Mark as read
	if svc.MarkAsRead {
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		err := c.Store(seqSet, item, flags, nil)
		if err != nil {
			db.logger.Error("IMAP failed to mark message as read", "service_id", svc.ServiceID, "error", err)
		}
	}

	// Delete if configured
	if svc.DeleteAfter {
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.DeletedFlag}
		err := c.Store(seqSet, item, flags, nil)
		if err != nil {
			db.logger.Error("IMAP failed to mark message for deletion", "service_id", svc.ServiceID, "error", err)
		}

		// Expunge to permanently delete
		err = c.Expunge(nil)
		if err != nil {
			db.logger.Error("IMAP failed to expunge messages", "service_id", svc.ServiceID, "error", err)
		}
	}
}

// sanitizeText sanitizes and truncates text to a maximum length.
func sanitizeText(text string, maxLen int) string {
	// Remove all non-printable characters using unicode.IsPrint (same as validation)
	text = strings.Map(func(r rune) rune {
		// Keep printable characters, tabs, and newlines
		if unicode.IsPrint(r) || r == '\t' || r == '\n' {
			return r
		}
		// Remove non-printable characters
		return -1
	}, text)

	// Trim leading and trailing whitespace (validation requires this)
	text = strings.TrimSpace(text)

	// Truncate if needed
	if len(text) > maxLen {
		text = text[:maxLen] + "..."
	}

	return text
}
