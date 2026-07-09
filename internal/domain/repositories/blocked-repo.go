package repositories

import (
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/storage"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type BlockedRepository interface {
	BlockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string, blockTime time.Time) (uuid.UUID, error)
	UnblockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string) error
}

type blockedRepository struct {
	storage *storage.Storage
}

func NewBlockedRepository(s *storage.Storage) BlockedRepository {
	return &blockedRepository{
		storage: s,
	}
}

func (br *blockedRepository) BlockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string, blockTime time.Time) (uuid.UUID, error) {
	op := "blockedRepository.BlockUser"
	query := `WITH found_user AS (
				SELECT id FROM user WHERE nickname = $2 AND id != $1 FOR KEY SHARE
			),
			deleted_friend AS (
				DELETE FROM friends f USING found_user fu
				WHERE f.user_id1 = fu.id AND f.user_id2 = $1) 
       			OR (f.user_id1 = $1 AND f.user_id2 = fu.id) 
				RETURNING fu.id AS deleted_friend_id
			),
			blocked_user AS (
				INSERT INTO blocked_users (blocker_id, blocked_id, block_time)
				VALUES ($1, $2, $3)
			)
			SELECT COALESCE(
			    (SELECT deleted_friend_id FROM deleted_friend),
			    '00000000-0000-0000-0000-000000000000'
			) FROM blocked_user`
	var id uuid.UUID
	err := br.storage.Pool.QueryRow(ctx, query, userId, blockedUserNick, blockTime).Scan(&id)
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

func (br *blockedRepository) UnblockUser(ctx context.Context, userId uuid.UUID, unblockedUserNick string) error {
	op := "blockedRepository.UnblockUser"
	query := `DELETE FROM blocked_users bu USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  bu.blocker_id = $1 AND bu.blocked_id = u.id`
	res, err := br.storage.Pool.Exec(ctx, query, userId, unblockedUserNick)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}
