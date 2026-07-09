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
	BlockUser(ctx context.Context, userId, blockedUserId uuid.UUID, blockTime time.Time) (bool, error)
	UnblockUser(ctx context.Context, userId, unblockedUserId uuid.UUID) error
}

type blockedRepository struct {
	storage *storage.Storage
}

func NewBlockedRepository(s *storage.Storage) BlockedRepository {
	return &blockedRepository{
		storage: s,
	}
}

func (br *blockedRepository) BlockUser(ctx context.Context, userId, blockedUserId uuid.UUID, blockTime time.Time) (bool, error) {
	op := "blockedRepository.BlockUser"
	query := `WITH deleted_friend AS (
				DELETE FROM friends WHERE user_id1 = $2 AND user_id2 = $1) 
       			OR (user_id1 = $1 AND user_id2 = $2) 
				RETURNING $2 AS deleted_friend_id
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
	err := br.storage.Pool.QueryRow(ctx, query, userId, blockedUserId, blockTime).Scan(&id)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return false, errs.ErrAlreadyExists(op, err)
		} else if errors.Is(err, storage.ErrNotFound()) {
			return false, errs.ErrNotFound(op)
		}
		return false, errs.NewAppError(op, err)
	}

	if id == uuid.Nil {
		return false, nil
	}

	return true, nil
}

func (br *blockedRepository) UnblockUser(ctx context.Context, userId, unblockedUserId uuid.UUID) error {
	op := "blockedRepository.UnblockUser"
	query := `DELETE FROM blocked_users blocker_id = $1 AND blocked_id = $2`
	res, err := br.storage.Pool.Exec(ctx, query, userId, unblockedUserId)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}
