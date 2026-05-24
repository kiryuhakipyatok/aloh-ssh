package repositories

import (
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/storage"
	"context"
	"time"

	"github.com/google/uuid"
)

type BlockedRepository interface {
	BlockUser(ctx context.Context, userId uuid.UUID, nickname string, blockTime time.Time) error
	UnblockUser(ctx context.Context, userId uuid.UUID, nickname string) error
}

type blockedRepository struct {
	storage *storage.Storage
}

func NewBlockedRepository(s *storage.Storage) BlockedRepository {
	return &blockedRepository{
		storage: s,
	}
}

func (br *blockedRepository) BlockUser(ctx context.Context, userId uuid.UUID, nickname string, blockTime time.Time) error {
	op := "blockedRepository.BlockUser"
	query := `WITH found_user AS (
    			SELECT id FROM users WHERE nickname = $2 AND id != $1 FOR KEY SHARE
			),
			deleted_friend AS (
				DELETE FROM friends f
    			USING found_user fu
    			WHERE (f.user_id1 = fu.id AND f.user_id2 = $1) 
       			OR (f.user_id1 = $1 AND f.user_id2 = fu.id)
			)
			INSERT INTO blocked_users (blocker_id, blocked_id, block_time)
			SELECT $1, fu.id, $3 FROM found_user fu`
	res, err := br.storage.Pool.Exec(ctx, query, userId, nickname, blockTime)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (br *blockedRepository) UnblockUser(ctx context.Context, userId uuid.UUID, nickname string) error {
	op := "blockedRepository.UnblockUser"
	query := `DELETE FROM blocked_users bu USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  bu.blocker_id = $1 AND bu.blocked_id = u.id`
	res, err := br.storage.Pool.Exec(ctx, query, userId, nickname)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}
