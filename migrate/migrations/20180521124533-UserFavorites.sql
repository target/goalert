-- +migrate Up
CREATE TABLE user_favorites (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
    tgt_service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    UNIQUE (user_id, tgt_service_id)
);
-- +migrate Down
DROP TABLE user_favorites;

