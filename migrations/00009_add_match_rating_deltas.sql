-- +goose Up
ALTER TABLE matches
ADD COLUMN IF NOT EXISTS white_rating_delta integer NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS black_rating_delta integer NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE matches
DROP COLUMN IF EXISTS black_rating_delta,
DROP COLUMN IF EXISTS white_rating_delta;
