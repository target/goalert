-- +migrate Up notransaction

-- Add IMAP to engine processing type enum
ALTER TYPE engine_processing_type ADD VALUE IF NOT EXISTS 'imap';

-- Initialize IMAP module version
INSERT INTO engine_processing_versions (type_id) VALUES ('imap');

-- Add IMAP integration key type
ALTER TYPE enum_integration_keys_type ADD VALUE IF NOT EXISTS 'imap';

-- Create table for per-service IMAP configuration
CREATE TABLE service_imap_config (
    service_id UUID PRIMARY KEY REFERENCES services(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,

    -- OAuth credentials (per-service)
    oauth_client_id TEXT,
    oauth_client_secret TEXT,
    oauth_refresh_token TEXT,

    -- IMAP connection settings
    host TEXT NOT NULL DEFAULT 'imap.gmail.com',
    port INT NOT NULL DEFAULT 993,
    username TEXT NOT NULL,
    use_tls BOOLEAN NOT NULL DEFAULT TRUE,
    mailbox TEXT NOT NULL DEFAULT 'INBOX',

    -- Polling configuration
    poll_interval_minutes INT NOT NULL DEFAULT 5 CHECK (poll_interval_minutes > 0 AND poll_interval_minutes <= 1440),
    mark_as_read BOOLEAN NOT NULL DEFAULT FALSE,
    delete_after BOOLEAN NOT NULL DEFAULT FALSE,
    last_polled_at TIMESTAMPTZ,

    -- Alert content customization
    include_headers BOOLEAN NOT NULL DEFAULT FALSE,
    include_from BOOLEAN NOT NULL DEFAULT TRUE,
    include_to BOOLEAN NOT NULL DEFAULT TRUE,
    include_subject BOOLEAN NOT NULL DEFAULT TRUE,
    include_body BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_service_imap_config_enabled ON service_imap_config(enabled) WHERE enabled = TRUE;

-- Create table for IMAP filter rules
CREATE TABLE imap_filter_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    -- Email filtering criteria (all conditions are ANDed within a rule)
    from_pattern TEXT,
    subject_pattern TEXT,
    to_pattern TEXT,

    -- Pattern matching mode: 'exact', 'contains', 'regex'
    match_mode TEXT NOT NULL DEFAULT 'contains' CHECK (match_mode IN ('exact', 'contains', 'regex')),

    -- Exclude reply emails (Re:, Fwd:, In-Reply-To, References headers)
    exclude_replies BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure at least one filter criterion is specified
    CHECK (from_pattern IS NOT NULL OR subject_pattern IS NOT NULL OR to_pattern IS NOT NULL)
);

CREATE INDEX idx_imap_filter_rules_service_id ON imap_filter_rules(service_id);
CREATE INDEX idx_imap_filter_rules_enabled ON imap_filter_rules(enabled) WHERE enabled = TRUE;

-- Create table to track processed IMAP messages (for deduplication)
CREATE TABLE imap_processed_messages (
    message_id TEXT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_imap_processed_messages_processed_at ON imap_processed_messages(processed_at);

-- +migrate Down

DROP TABLE IF EXISTS imap_processed_messages;
DROP TABLE IF EXISTS imap_filter_rules;
DROP TABLE IF EXISTS service_imap_config;

DELETE FROM engine_processing_versions WHERE type_id = 'imap';

-- Note: Cannot remove enum values in PostgreSQL without recreating the entire enum
-- Manual intervention required if this migration needs to be rolled back
