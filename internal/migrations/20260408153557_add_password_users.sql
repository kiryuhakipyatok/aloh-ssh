-- +goose Up
ALTER TABLE IF EXISTS users
ADD COLUMN password BYTEA UNIQUE;


-- +goose Down
ALTER TABLE IF EXISTS users
DROP COLUMN password;
