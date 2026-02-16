package imapmanager

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
)

// PollArgs are the arguments for the IMAP polling job.
type PollArgs struct{}

func (PollArgs) Kind() string { return "imap-poll" }

// IMAPServiceConfig holds the IMAP configuration for a service.
type IMAPServiceConfig struct {
	ServiceID           uuid.UUID
	ServiceName         string
	Enabled             bool
	OAuthClientID       sql.NullString
	OAuthClientSecret   sql.NullString
	OAuthRefreshToken   sql.NullString
	Host                string
	Port                int
	Username            string
	UseTLS              bool
	Mailbox             string
	PollIntervalMinutes int
	MarkAsRead          bool
	DeleteAfter         bool
	IncludeHeaders      bool
	IncludeFrom         bool
	IncludeTo           bool
	IncludeSubject      bool
	IncludeBody         bool
	LastPolledAt        sql.NullTime
}

// PollIMAP connects to the IMAP server and polls for new messages.
func (db *DB) PollIMAP(ctx context.Context, job *river.Job[PollArgs]) error {
	// Get all services with IMAP enabled
	var services []IMAPServiceConfig
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		queries := gadb.New(tx)
		activeServices, err := queries.IMAPGetActiveServices(ctx)
		if err != nil {
			return fmt.Errorf("get active services: %w", err)
		}

		for _, svc := range activeServices {
			services = append(services, IMAPServiceConfig{
				ServiceID:           svc.ID,
				ServiceName:         svc.Name,
				Enabled:             svc.Enabled,
				OAuthClientID:       svc.OauthClientID,
				OAuthClientSecret:   svc.OauthClientSecret,
				OAuthRefreshToken:   svc.OauthRefreshToken,
				Host:                svc.Host,
				Port:                int(svc.Port),
				Username:            svc.Username,
				UseTLS:              svc.UseTls,
				Mailbox:             svc.Mailbox,
				PollIntervalMinutes: int(svc.PollIntervalMinutes),
				MarkAsRead:          svc.MarkAsRead,
				DeleteAfter:         svc.DeleteAfter,
				IncludeHeaders:      svc.IncludeHeaders,
				IncludeFrom:         svc.IncludeFrom,
				IncludeTo:           svc.IncludeTo,
				IncludeSubject:      svc.IncludeSubject,
				IncludeBody:         svc.IncludeBody,
				LastPolledAt:        svc.LastPolledAt,
			})
		}
		return nil
	})

	if err != nil {
		db.logger.Error("failed to get active IMAP services", "error", err)
		return fmt.Errorf("get active services: %w", err)
	}

	if len(services) == 0 {
		// No services with IMAP enabled
		return nil
	}

	// Poll each service independently
	for _, svc := range services {
		// Check if enough time has elapsed since the last poll
		if svc.LastPolledAt.Valid {
			nextPollTime := svc.LastPolledAt.Time.Add(time.Duration(svc.PollIntervalMinutes) * time.Minute)
			if time.Now().Before(nextPollTime) {
				db.logger.Debug("skipping poll - interval not elapsed", "service_id", svc.ServiceID, "next_poll", nextPollTime)
				continue
			}
		}

		err = db.pollService(ctx, svc)
		if err != nil {
			db.logger.Error("failed to poll service", "service_id", svc.ServiceID, "error", err)
			// Continue with other services
			continue
		}

		// Update last_polled_at timestamp after successful poll
		err = db.updateLastPolled(ctx, svc.ServiceID)
		if err != nil {
			db.logger.Error("failed to update last polled timestamp", "service_id", svc.ServiceID, "error", err)
		}
	}

	return nil
}

// pollService polls IMAP for a single service.
func (db *DB) pollService(ctx context.Context, svc IMAPServiceConfig) error {
	// Connect to IMAP server with service-specific config
	c, err := db.connectIMAPForService(ctx, svc)
	if err != nil {
		db.logger.Error("IMAP connection failed", "service_id", svc.ServiceID, "error", err)
		return fmt.Errorf("connect to IMAP: %w", err)
	}
	defer func() {
		if err := c.Logout(); err != nil {
			db.logger.Debug("IMAP logout error", "error", err)
		}
	}()

	// Select mailbox
	mailbox := svc.Mailbox
	if mailbox == "" {
		mailbox = "INBOX"
	}

	_, err = c.Select(mailbox, false)
	if err != nil {
		db.logger.Error("IMAP select mailbox failed", "service_id", svc.ServiceID, "mailbox", mailbox, "error", err)
		return fmt.Errorf("select mailbox: %w", err)
	}

	// Search for unseen messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	seqNums, err := c.Search(criteria)
	if err != nil {
		db.logger.Error("IMAP search failed", "service_id", svc.ServiceID, "error", err)
		return fmt.Errorf("search messages: %w", err)
	}

	if len(seqNums) == 0 {
		// No new messages
		return nil
	}

	db.logger.Info("IMAP found new messages", "service_id", svc.ServiceID, "count", len(seqNums))

	// Fetch messages in batches
	const batchSize = 50
	for i := 0; i < len(seqNums); i += batchSize {
		end := i + batchSize
		if end > len(seqNums) {
			end = len(seqNums)
		}

		batch := seqNums[i:end]
		err = db.processBatchForService(ctx, c, batch, svc)
		if err != nil {
			db.logger.Error("IMAP process batch failed", "service_id", svc.ServiceID, "error", err)
			// Continue with next batch
		}
	}

	return nil
}

// connectIMAPForService establishes a connection to the IMAP server for a specific service.
func (db *DB) connectIMAPForService(ctx context.Context, svc IMAPServiceConfig) (*client.Client, error) {
	// Determine port
	port := svc.Port
	if port == 0 {
		if svc.UseTLS {
			port = 993 // Default TLS port
		} else {
			port = 143 // Default STARTTLS port
		}
	}

	addr := fmt.Sprintf("%s:%d", svc.Host, port)

	hasOAuth := svc.OAuthClientID.Valid && svc.OAuthClientSecret.Valid && svc.OAuthRefreshToken.Valid
	db.logger.Info("Connecting to IMAP server", "service_id", svc.ServiceID, "host", svc.Host, "port", port, "tls", svc.UseTLS, "oauth", hasOAuth)

	var c *client.Client
	var err error

	if svc.UseTLS {
		// Connect with TLS
		tlsConfig := &tls.Config{
			ServerName: svc.Host,
			MinVersion: tls.VersionTLS12, // Gmail requires TLS 1.2+
		}
		c, err = client.DialTLS(addr, tlsConfig)
	} else {
		// Connect without TLS (will use STARTTLS if available)
		c, err = client.Dial(addr)
		if err == nil {
			tlsConfig := &tls.Config{
				ServerName: svc.Host,
				MinVersion: tls.VersionTLS12,
			}
			err = c.StartTLS(tlsConfig)
		}
	}

	if err != nil {
		db.logger.Error("Failed to dial IMAP server", "service_id", svc.ServiceID, "host", svc.Host, "port", port, "error", err)
		return nil, fmt.Errorf("dial IMAP server: %w", err)
	}

	// Authenticate - use OAuth if configured, otherwise use password (from global config)
	if hasOAuth {
		// Use OAuth 2.0 XOAUTH2 authentication
		accessToken, err := db.getOAuthAccessTokenForService(ctx, svc)
		if err != nil {
			_ = c.Logout()
			return nil, fmt.Errorf("get OAuth access token: %w", err)
		}

		saslClient := NewXOAUTH2Client(svc.Username, accessToken)
		err = c.Authenticate(saslClient)
		if err != nil {
			_ = c.Logout()
			return nil, fmt.Errorf("IMAP OAuth authentication: %w", err)
		}

		db.logger.Info("IMAP authenticated with OAuth 2.0", "service_id", svc.ServiceID)
	} else {
		// OAuth not configured for this service
		_ = c.Logout()
		return nil, fmt.Errorf("OAuth credentials required but not configured for service (configure via service IMAP settings)")
	}

	return c, nil
}

// processBatchForService processes a batch of message sequence numbers for a specific service.
func (db *DB) processBatchForService(ctx context.Context, c *client.Client, seqNums []uint32, svc IMAPServiceConfig) error {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	// Fetch envelope and body
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchRFC822}, messages)
	}()

	for msg := range messages {
		err := db.processMessageForService(ctx, c, msg, svc)
		if err != nil {
			db.logger.Error("IMAP process message failed", "service_id", svc.ServiceID, "error", err)
			// Continue with next message
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}

	return nil
}

// updateLastPolled updates the last_polled_at timestamp for a service.
func (db *DB) updateLastPolled(ctx context.Context, serviceID uuid.UUID) error {
	return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		queries := gadb.New(tx)
		return queries.IMAPUpdateLastPolled(ctx, serviceID)
	})
}
