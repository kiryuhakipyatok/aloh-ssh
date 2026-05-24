package repositories

import (
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/storage"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type FriendshipRepository interface {
	NewFriendRequest(ctx context.Context, userId uuid.UUID, nickname string, timeReq time.Time) (uuid.UUID, error)
	AcceptFriendship(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error)
	DenyFriendship(ctx context.Context, userId uuid.UUID, nickname string) error
	DeleteFromFriends(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error)
}

type friendshipRepository struct {
	storage *storage.Storage
}

func NewFriendshipRepository(fr *storage.Storage) FriendshipRepository {
	return &friendshipRepository{
		storage: fr,
	}
}

func (fr *friendshipRepository) NewFriendRequest(ctx context.Context, userId uuid.UUID, nickname string, reqTime time.Time) (uuid.UUID, error) {
	op := "friendshipRepository.NewFriendRequest"
	query := `WITH found_user AS (
    			SELECT u.id FROM users u WHERE u.nickname = $2 AND u.id != $1 AND NOT EXISTS (
					SELECT 1 FROM blocked_users bu WHERE (bu.blocker_id = $1 AND bu.blocked_id = u.id) OR
					(bu.blocker_id = u.id AND bu.blocked_id = $1)
				)
			),
			inserted_friend AS (
    			INSERT INTO friends (user_id1, user_id2, req_time)
				SELECT $1, id, $3 FROM found_user
				RETURNING user_id2
			)
			SELECT user_id2 FROM inserted_friend`
	var id uuid.UUID
	err := fr.storage.Pool.QueryRow(ctx, query, userId, nickname, reqTime).Scan(&id)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return uuid.UUID{}, errs.ErrAlreadyExists(op, err)
		} else if errors.Is(err, storage.ErrNotFound()) {
			return uuid.UUID{}, errs.ErrNotFound(op)
		}
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	return id, nil
}

func (fr *friendshipRepository) AcceptFriendship(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "friendshipRepository.AcceptFriendship"
	query := `UPDATE friends f SET status = 'active' FROM users u
			  WHERE u.nickname = $2 AND u.id != $1 AND NOT EXISTS (
					SELECT 1 FROM blocked_users bu WHERE (bu.blocker_id = $1 AND bu.blocked_id = u.id) OR
					(bu.blocker_id = u.id AND bu.blocked_id = $1) AND
      		  f.user_id1 = u.id AND f.user_id2 = $1 RETURNING u.id`
	var id uuid.UUID
	err := fr.storage.Pool.QueryRow(ctx, query, userId, nickname).Scan(&id)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return uuid.UUID{}, errs.ErrAlreadyExists(op, err)
		} else if errors.Is(err, storage.ErrNotFound()) {
			return uuid.UUID{}, errs.ErrNotFound(op)
		}
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	return id, nil
}

func (fr *friendshipRepository) DenyFriendship(ctx context.Context, userId uuid.UUID, nickname string) error {
	op := "friendshipRepository.DenyFriendship"
	query := `DELETE FROM friends f USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  f.user_id1 = u.id AND f.user_id2 = $1`
	res, err := fr.storage.Pool.Exec(ctx, query, userId, nickname)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (fr *friendshipRepository) DeleteFromFriends(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "friendshipRepository.DeleteFromFriends"
	query := `DELETE FROM friends f USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  ((f.user_id1 = u.id AND f.user_id2 = $1) OR (f.user_id1 = $1 AND f.user_id2 = u.id)) AND f.status = 'active'
			  RETURNING u.id`
	var id uuid.UUID
	err := fr.storage.Pool.QueryRow(ctx, query, userId, nickname).Scan(&id)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return uuid.UUID{}, errs.ErrAlreadyExists(op, err)
		} else if errors.Is(err, storage.ErrNotFound()) {
			return uuid.UUID{}, errs.ErrNotFound(op)
		}
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	return id, nil
}
