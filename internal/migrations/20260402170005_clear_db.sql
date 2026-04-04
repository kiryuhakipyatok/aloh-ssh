-- Active: 1774903565768@@127.0.0.1@4444
-- +goose Up
DELETE FROM users;

-- +goose Down
SELECT 'down SQL query';
