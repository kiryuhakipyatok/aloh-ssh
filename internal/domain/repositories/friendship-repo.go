package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/storage"

	"github.com/google/uuid"
)

type FriendshipRepository interface {
	NewFriendRequest(ctx context.Context, userId uuid.UUID, friendNickname string, timeReq time.Time) (uuid.UUID, error)
	AcceptFriendship(ctx context.Context, userId, friendId uuid.UUID) error
	DenyFriendship(ctx context.Context, userId, friendId uuid.UUID) error
	DeleteFromFriends(ctx context.Context, userId, friendId uuid.UUID) error
	GetFriendsRequestForId(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
}

type friendshipRepository struct {
	storage *storage.Storage
}

func NewFriendshipRepository(fr *storage.Storage) FriendshipRepository {
	return &friendshipRepository{
		storage: fr,
	}
}

func (fr *friendshipRepository) NewFriendRequest(ctx context.Context, userId uuid.UUID, friendNickname string, timeReq time.Time) (uuid.UUID, error) {
	op := "friendshipRepository.NewFriendRequest"
	query := `INSERT INTO friends (user_id1, user_id2, req_time)
			  SELECT $1, id, $3 FROM users WHERE nickname = $2 AND id != $1
              RETURNING user_id2`
	var id uuid.UUID
	err := fr.storage.Pool.QueryRow(ctx, query, userId, friendNickname, timeReq).Scan(&id)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return uuid.Nil, errs.ErrAlreadyExists(op, err)
		} else if errors.Is(err, storage.ErrNotFound()) {
			return uuid.Nil, errs.ErrNotFound(op)
		}
		return uuid.Nil, errs.NewAppError(op, err)
	}

	return id, nil
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
	query := `DELETE FROM friends WHERE user_id1 = $1 AND user_id2 = $2`
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
	query := `DELETE FROM friends WHERE ((user_id1 = $2 AND user_id2 = $1) 
			  OR (user_id1 = $1 AND user_id2 = $2)) AND status = 'active'`
	res, err := fr.storage.Pool.Exec(ctx, query, friendId, userId)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (fr *friendshipRepository) GetFriendsRequestForId(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error) {
	op := "friendshipRepository.GetFriendsRequestForId"

	query := `SELECT user1_id FROM friends WHERE user_id2 = $1`
	var friendReqs []uuid.UUID
	rows, err := fr.storage.Pool.Query(ctx, query, userId)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		friendReqs = append(friendReqs, uid)
	}

	if err := rows.Err(); err != nil {
		return nil, errs.NewAppError(op, err)
	}

	return friendReqs, nil
}
