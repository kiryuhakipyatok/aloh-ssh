-- +goose Up
CREATE UNIQUE INDEX unique_friendship_idx ON friends (LEAST(user_id1, user_id2), GREATEST(user_id1, user_id2));

-- +goose Down
DROP INDEX IF EXISTS unique_friendship_idx;