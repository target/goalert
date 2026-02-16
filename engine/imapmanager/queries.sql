-- name: IMAPFilterRulesForService :many
-- IMAPFilterRulesForService returns all enabled filter rules for a given service.
SELECT
    id,
    service_id,
    name,
    from_pattern,
    subject_pattern,
    to_pattern,
    match_mode,
    exclude_replies
FROM
    imap_filter_rules
WHERE
    service_id = $1
    AND enabled = TRUE;

-- name: IMAPMessageProcessed :one
-- IMAPMessageProcessed checks if a message with the given Message-ID has already been processed.
SELECT
    EXISTS (
        SELECT
            1
        FROM
            imap_processed_messages
        WHERE
            message_id = $1) AS processed;

-- name: IMAPMarkMessageProcessed :exec
-- IMAPMarkMessageProcessed records that a message has been processed (prevents duplicate alerts).
INSERT INTO imap_processed_messages (message_id, processed_at)
    VALUES ($1, now())
ON CONFLICT (message_id)
    DO NOTHING;

-- name: IMAPCleanupProcessedMessages :execrows
-- IMAPCleanupProcessedMessages removes processed message records older than 30 days.
DELETE FROM imap_processed_messages
WHERE message_id = ANY (
        SELECT
            message_id
        FROM
            imap_processed_messages
        WHERE
            processed_at < now() - '30 days'::interval
        ORDER BY
            processed_at
        LIMIT 1000
        FOR UPDATE
            SKIP LOCKED);

-- name: IMAPGetActiveServices :many
-- IMAPGetActiveServices returns all services with IMAP enabled and at least one enabled filter rule.
SELECT DISTINCT
    s.id,
    s.name,
    sic.enabled,
    sic.oauth_client_id,
    sic.oauth_client_secret,
    sic.oauth_refresh_token,
    sic.host,
    sic.port,
    sic.username,
    sic.use_tls,
    sic.mailbox,
    sic.poll_interval_minutes,
    sic.mark_as_read,
    sic.delete_after,
    sic.include_headers,
    sic.include_from,
    sic.include_to,
    sic.include_subject,
    sic.include_body,
    sic.last_polled_at
FROM
    services s
    JOIN service_imap_config sic ON sic.service_id = s.id
    JOIN imap_filter_rules ifr ON ifr.service_id = s.id
WHERE
    sic.enabled = TRUE
    AND ifr.enabled = TRUE;

-- name: IMAPFilterRulesAll :many
-- IMAPFilterRulesAll returns all filter rules for a given service (enabled and disabled).
SELECT
    id,
    service_id,
    name,
    enabled,
    from_pattern,
    subject_pattern,
    to_pattern,
    match_mode,
    exclude_replies,
    created_at,
    updated_at
FROM
    imap_filter_rules
WHERE
    service_id = $1
ORDER BY
    created_at DESC;

-- name: IMAPFilterRuleGet :one
-- IMAPFilterRuleGet returns a single filter rule by ID.
SELECT
    id,
    service_id,
    name,
    enabled,
    from_pattern,
    subject_pattern,
    to_pattern,
    match_mode,
    exclude_replies,
    created_at,
    updated_at
FROM
    imap_filter_rules
WHERE
    id = $1;

-- name: IMAPFilterRuleCreate :one
-- IMAPFilterRuleCreate creates a new filter rule.
INSERT INTO imap_filter_rules (
    service_id,
    name,
    enabled,
    from_pattern,
    subject_pattern,
    to_pattern,
    match_mode,
    exclude_replies
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, service_id, name, enabled, from_pattern, subject_pattern, to_pattern, match_mode, exclude_replies, created_at, updated_at;

-- name: IMAPFilterRuleUpdate :exec
-- IMAPFilterRuleUpdate updates an existing filter rule.
UPDATE imap_filter_rules
SET
    name = COALESCE(sqlc.narg('name'), name),
    enabled = COALESCE(sqlc.narg('enabled'), enabled),
    from_pattern = COALESCE(sqlc.narg('from_pattern'), from_pattern),
    subject_pattern = COALESCE(sqlc.narg('subject_pattern'), subject_pattern),
    to_pattern = COALESCE(sqlc.narg('to_pattern'), to_pattern),
    match_mode = COALESCE(sqlc.narg('match_mode'), match_mode),
    exclude_replies = COALESCE(sqlc.narg('exclude_replies'), exclude_replies),
    updated_at = now()
WHERE
    id = sqlc.arg('id');

-- name: IMAPFilterRuleDelete :exec
-- IMAPFilterRuleDelete deletes a filter rule by ID.
DELETE FROM imap_filter_rules
WHERE
    id = $1;

-- name: IMAPConfigGet :one
-- IMAPConfigGet returns the IMAP configuration for a given service.
SELECT
    service_id,
    enabled,
    oauth_client_id,
    oauth_client_secret,
    oauth_refresh_token,
    host,
    port,
    username,
    use_tls,
    mailbox,
    poll_interval_minutes,
    mark_as_read,
    delete_after,
    include_headers,
    include_from,
    include_to,
    include_subject,
    include_body,
    created_at,
    updated_at,
    last_polled_at
FROM
    service_imap_config
WHERE
    service_id = $1;

-- name: IMAPConfigCreate :one
-- IMAPConfigCreate creates a new IMAP configuration for a service.
INSERT INTO service_imap_config (
    service_id,
    enabled,
    oauth_client_id,
    oauth_client_secret,
    oauth_refresh_token,
    host,
    port,
    username,
    use_tls,
    mailbox,
    poll_interval_minutes,
    mark_as_read,
    delete_after,
    include_headers,
    include_from,
    include_to,
    include_subject,
    include_body
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
) RETURNING service_id, enabled, oauth_client_id, oauth_client_secret, oauth_refresh_token, host, port, username, use_tls, mailbox, poll_interval_minutes, mark_as_read, delete_after, include_headers, include_from, include_to, include_subject, include_body, created_at, updated_at, last_polled_at;

-- name: IMAPConfigUpdate :exec
-- IMAPConfigUpdate updates an existing IMAP configuration.
UPDATE service_imap_config
SET
    enabled = COALESCE(sqlc.narg('enabled'), enabled),
    oauth_client_id = COALESCE(sqlc.narg('oauth_client_id'), oauth_client_id),
    oauth_client_secret = COALESCE(sqlc.narg('oauth_client_secret'), oauth_client_secret),
    oauth_refresh_token = COALESCE(sqlc.narg('oauth_refresh_token'), oauth_refresh_token),
    host = COALESCE(sqlc.narg('host'), host),
    port = COALESCE(sqlc.narg('port'), port),
    username = COALESCE(sqlc.narg('username'), username),
    use_tls = COALESCE(sqlc.narg('use_tls'), use_tls),
    mailbox = COALESCE(sqlc.narg('mailbox'), mailbox),
    poll_interval_minutes = COALESCE(sqlc.narg('poll_interval_minutes'), poll_interval_minutes),
    mark_as_read = COALESCE(sqlc.narg('mark_as_read'), mark_as_read),
    delete_after = COALESCE(sqlc.narg('delete_after'), delete_after),
    include_headers = COALESCE(sqlc.narg('include_headers'), include_headers),
    include_from = COALESCE(sqlc.narg('include_from'), include_from),
    include_to = COALESCE(sqlc.narg('include_to'), include_to),
    include_subject = COALESCE(sqlc.narg('include_subject'), include_subject),
    include_body = COALESCE(sqlc.narg('include_body'), include_body),
    updated_at = now()
WHERE
    service_id = sqlc.arg('service_id');

-- name: IMAPConfigDelete :exec
-- IMAPConfigDelete deletes an IMAP configuration for a service.
DELETE FROM service_imap_config
WHERE
    service_id = $1;

-- name: IMAPUpdateLastPolled :exec
-- IMAPUpdateLastPolled updates the last_polled_at timestamp for a service.
UPDATE service_imap_config
SET
    last_polled_at = now()
WHERE
    service_id = $1;
