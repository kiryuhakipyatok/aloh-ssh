-- +goose Up
ALTER TABLE IF EXISTS users
ADD COLUMN tagline VARCHAR(28) NOT NULL;

-- +goose Down
ALTER TABLE IF EXISTS users
DROP COLUMN tagline;
