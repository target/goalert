-- +migrate Up
CREATE TABLE alerts_metadata(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  alert_id INT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
  metadata jsonb
);

-- +migrate Down
DROP TABLE alerts_metadata;
