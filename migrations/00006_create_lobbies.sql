-- +goose Up
CREATE TABLE lobbies (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(6) NOT NULL UNIQUE,

    host_id      BIGINT NOT NULL REFERENCES users(id),
    guest_id     BIGINT REFERENCES users(id),

    host_ready BOOLEAN NOT NULL DEFAULT FALSE,
    guest_ready BOOLEAN NOT NULL DEFAULT FALSE,

    status       VARCHAR(20) NOT NULL DEFAULT 'waiting',
    time_control INTEGER NOT NULL DEFAULT 600,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lobbies_code ON lobbies(code);
CREATE INDEX idx_lobbies_status ON lobbies(status);
CREATE INDEX idx_lobbies_host_id ON lobbies(host_id);

-- +goose Down
DROP TABLE lobbies;
