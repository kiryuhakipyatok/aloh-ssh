package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/models"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/storage"

	"github.com/google/uuid"
)

type BlockedRepository interface {
	BlockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string, blockTime time.Time) (uuid.UUID, error)
	UnblockUser(ctx context.Context, userId uuid.UUID, blockedUserNick string) (uuid.UUID, error)
	FetchBlockedUsersById(ctx context.Context, userId uuid.UUID) ([]models.Identity, error)
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
				SELECT id FROM users WHERE nickname = $2 AND id != $1 FOR KEY SHARE
			),
			deleted_friend AS (
				DELETE FROM friends f USING found_user fu
				WHERE (f.user_id1 = fu.id AND f.user_id2 = $1) 
       			OR (f.user_id1 = $1 AND f.user_id2 = fu.id) 
			),
			blocked_user AS (
				INSERT INTO blocked_users (blocker_id, blocked_id, block_time)
				SELECT $1, id, $3 FROM found_user
				RETURNING blocked_id
			)
			SELECT blocked_id FROM blocked_user`
	var bId uuid.UUID

	err := br.storage.Pool.QueryRow(ctx, query, userId, blockedUserNick, blockTime).Scan(&bId)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return uuid.Nil, errs.ErrAlreadyExists(op, err)
		} else if errors.Is(err, storage.ErrNotFound()) {
			return uuid.Nil, errs.ErrNotFound(op)
		}
		return uuid.Nil, errs.NewAppError(op, err)
	}

	return bId, nil
}

func (br *blockedRepository) UnblockUser(ctx context.Context, userId uuid.UUID, unblockedUserNick string) (uuid.UUID, error) {
	op := "blockedRepository.UnblockUser"
	query := `DELETE FROM blocked_users bu USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  bu.blocker_id = $1 AND bu.blocked_id = u.id
			  RETURNING bu.blocked_id
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

func (br *blockedRepository) FetchBlockedUsersById(ctx context.Context, userId uuid.UUID) ([]models.Identity, error) {
	op := "blockedRepository.GetBlockFetchBlockedUsersByIdedUsersById"

	query := `SELECT COALESCE(
    			json_agg(json_build_object(
            		'id', bu.blocker_id,
            		'nickname', u.nickname
        		)), '[]') FROM blocked_users bu
			  JOIN users u ON bu.blocker_id = u.id
			  WHERE bu.blocked_id = $1
			`
	var blockers []models.Identity
	err := br.storage.Pool.QueryRow(ctx, query, userId).Scan(&blockers)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}

	return blockers, nil
}
