-- +migrate Up

-- Free-form comments left on an alert by users, for triage context that does
-- not belong in the system-generated alert log.
--
-- alert_id CASCADEs: alert cleanup is a plain `DELETE FROM alerts`, so comments
-- are removed with their alert and cannot accumulate. This is deliberately
-- unlike alert_logs, which has no foreign key and therefore needs its own
-- separate cleanup job to avoid orphans.
--
-- user_id is SET NULL rather than CASCADE: deleting a user must not erase the
-- comment history on an alert. The text and timestamp survive, and only the
-- attribution is lost.
CREATE TABLE alert_comments (
    id BIGSERIAL PRIMARY KEY,
    alert_id BIGINT NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    CONSTRAINT alert_comments_body_not_empty CHECK (length(trim(body)) > 0),
    CONSTRAINT alert_comments_body_max_length CHECK (length(body) <= 4096)
);

-- Comments are always read for a given alert, oldest first.
CREATE INDEX idx_alert_comments_alert ON alert_comments (alert_id, created_at, id);

-- +migrate Down

DROP TABLE alert_comments;
