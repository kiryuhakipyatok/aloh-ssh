package repositories

import (
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/storage"
	"context"
	"time"

	"github.com/google/uuid"
)

type FriendshipRepository interface {
	NewFriendRequest(ctx context.Context, userId, friendId uuid.UUID, timeReq time.Time) error
	AcceptFriendship(ctx context.Context, userId, friendId uuid.UUID) error
	DenyFriendship(ctx context.Context, userId, friendId uuid.UUID) error
	DeleteFromFriends(ctx context.Context, userId, friendId uuid.UUID) error
}

type friendshipRepository struct {
	storage *storage.Storage
}

func NewFriendshipRepository(fr *storage.Storage) FriendshipRepository {
	return &friendshipRepository{
		storage: fr,
	}
}

func (fr *friendshipRepository) NewFriendRequest(ctx context.Context, userId, friendId uuid.UUID, timeReq time.Time) error {
	op := "friendshipRepository.NewFriendRequest"
	query := `INSERT INTO friends (user_id1, user_id2, req_time) VALUES ($1, $2, $3)`
	res, err := fr.storage.Pool.Exec(ctx, query, userId, friendId, timeReq)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (fr *friendshipRepository) AcceptFriendship(ctx context.Context, userId, friendId uuid.UUID) error {
	op := "friendshipRepository.AcceptFriendship"
	query := `UPDATE friends f SET status = 'active' WHERE user_id1 = $1 AND user_id2 = $2`
	res, err := fr.storage.Pool.Exec(ctx, query, friendId, userId)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (fr *friendshipRepository) DenyFriendship(ctx context.Context, userId, friendId uuid.UUID) error {
	op := "friendshipRepository.DenyFriendship"
	query := `DELETE FROM friends user_id1 = $1 AND user_id2 = $2`
	res, err := fr.storage.Pool.Exec(ctx, query, friendId, userId)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (fr *friendshipRepository) DeleteFromFriends(ctx context.Context, userId, friendId uuid.UUID) error {
	op := "friendshipRepository.DeleteFromFriends"
	query := `DELETE FROM friends ((user_id1 = $2 AND user_id2 = $1) OR (user_id1 = $1 AND user_id2 = $2)) AND status = 'active'`
	res, err := fr.storage.Pool.Exec(ctx, query, friendId, userId)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}