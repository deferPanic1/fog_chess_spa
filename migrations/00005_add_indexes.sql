-- +goose Up
CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

CREATE INDEX idx_matches_white_player ON matches(white_player);
CREATE INDEX idx_matches_black_player ON matches(black_player);
CREATE INDEX idx_matches_status       ON matches(status);

-- +goose Down
DROP INDEX idx_matches_status;
DROP INDEX idx_matches_black_player;
DROP INDEX idx_matches_white_player;
DROP INDEX idx_refresh_tokens_expires_at;
DROP INDEX idx_refresh_tokens_user_id;
