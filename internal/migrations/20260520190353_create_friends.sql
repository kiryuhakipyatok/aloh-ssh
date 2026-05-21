-- +goose Up
CREATE TYPE friendship_status AS ENUM ('pending', 'accepted', 'blocked');
CREATE TABLE IF NOT EXISTS friends(
    user_id1 UUID UNIQUE NOT NULL,
    user_id2 UUID UNIQUE NOT NULL,
    status friendship_status NOT NULL DEFAULT 'pending',
    PRIMARY KEY (user_id1,user_id2),
    FOREIGN KEY (user_id1) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id2) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT not_self CHECK (user_id1 < user_id2)
);

-- +goose Down
DELETE TABLE IF EXISTS friends;
DELETE TYPE IF EXISTS friendship_status;