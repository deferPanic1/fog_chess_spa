-- +goose Up
ALTER TABLE matches
    ADD COLUMN board_fen    TEXT NOT NULL DEFAULT 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
    ADD COLUMN current_turn VARCHAR(5) NOT NULL DEFAULT 'white';

-- +goose Down
ALTER TABLE matches
    DROP COLUMN board_fen,
    DROP COLUMN current_turn;
