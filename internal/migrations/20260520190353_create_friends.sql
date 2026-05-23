-- +goose Up
CREATE TYPE friendship_status AS ENUM ('pending', 'active');
CREATE TABLE IF NOT EXISTS friends(
    user_id1 UUID NOT NULL,
    user_id2 UUID NOT NULL,
    req_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status friendship_status NOT NULL DEFAULT 'pending',
    PRIMARY KEY (user_id1,user_id2),
    FOREIGN KEY (user_id1) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id2) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT not_self CHECK (user_id1 <> user_id2)
);

-- +goose Down
DROP TABLE IF EXISTS friends;
DROP TYPE IF EXISTS friendship_status;