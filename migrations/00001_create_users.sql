-- +goose Up
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    username   TEXT NOT NULL UNIQUE,
    pass_hash  BYTEA NOT NULL,
    rating     INTEGER NOT NULL DEFAULT 1200,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;
