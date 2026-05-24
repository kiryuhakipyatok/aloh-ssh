-- +goose Up
CREATE TABLE IF NOT EXISTS blocked_users(
    blocker_id UUID NOT NULL,
    blocked_id UUID NOT NULL,
    block_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (blocker_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (blocker_id, blocked_id),
    CONSTRAINT not_self CHECK (blocker_id <> blocked_id)
);

-- +goose Down
DROP TABLE IF EXISTS blocked_users;
