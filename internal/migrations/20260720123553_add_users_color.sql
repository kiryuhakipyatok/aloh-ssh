-- +goose Up
ALTER TABLE IF EXISTS users
ADD COLUMN color VARCHAR(7) NOT NULL DEFAULT '#random';

-- +goose Down
ALTER TABLE IF EXISTS users
DROP COLUMN color;