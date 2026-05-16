-- +goose Up
CREATE INDEX IF NOT EXISTS nickname_index ON users (nickname);

-- +goose Down
DROP INDEX IF EXISTS nickname_index;