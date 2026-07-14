package repositories

import (
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/storage"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type BlockedRepository interface {
	BlockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string, blockTime time.Time) (uuid.UUID, uuid.UUID, error)
	UnblockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string) (uuid.UUID, error)
}

type blockedRepository struct {
	storage *storage.Storage
}

func NewBlockedRepository(s *storage.Storage) BlockedRepository {
	return &blockedRepository{
		storage: s,
	}
}

func (br *blockedRepository) BlockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string, blockTime time.Time) (uuid.UUID, uuid.UUID, error) {
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
				INSERT INTO blocked_users bu (bu.blocker_id, bu.blocked_id, bu.block_time)
				USING found_user fu VALUES ($1, fu.id, $3)
				RETURNING fu.id
			)
			SELECT COALESCE(
			    (SELECT deleted_friend_id FROM deleted_friend),
			    '00000000-0000-0000-0000-000000000000'
				), SELECT blocked_id FROM blocked_user
			FROM blocked_user`
	var ids struct {
		fId uuid.UUID
		bId uuid.UUID
	}
	err := br.storage.Pool.QueryRow(ctx, query, userId, blockedUserNick, blockTime).Scan(&ids)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return uuid.Nil, uuid.Nil, errs.ErrAlreadyExists(op, err)
		} else if errors.Is(err, storage.ErrNotFound()) {
			return uuid.Nil, uuid.Nil, errs.ErrNotFound(op)
		}
		return uuid.Nil, uuid.Nil, errs.NewAppError(op, err)
	}

	return ids.bId, ids.fId, nil
}

func (br *blockedRepository) UnblockUser(ctx context.Context, userId uuid.UUID, unblockedUserNick string) (uuid.UUID, error) {
	op := "blockedRepository.UnblockUser"
	query := `DELETE FROM blocked_users bu USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  bu.blocker_id = $1 AND bu.blocked_id = u.id
			  RETURNING u.id
			`
	var id uuid.UUID
	err := br.storage.Pool.QueryRow(ctx, query, userId, unblockedUserNick).Scan(&id)
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
