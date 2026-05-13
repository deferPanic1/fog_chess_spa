-- +goose Up
CREATE TABLE matches (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    white_player   BIGINT NOT NULL REFERENCES users(id),
    black_player   BIGINT NOT NULL REFERENCES users(id),
    status         VARCHAR(20) NOT NULL DEFAULT 'active',
    result         VARCHAR(20),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at    TIMESTAMPTZ
);

-- +goose Down
DROP TABLE matches;
