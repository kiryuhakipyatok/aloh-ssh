package repositories

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/storage"
	"context"
	"errors"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetKey(ctx context.Context, nickname string) ([]byte, error)
	ExistenceCheck(ctx context.Context, nickname, key string) (bool, error)
}

type userRepository struct {
	storage *storage.Storage
}

func NewUserRepository(s *storage.Storage) UserRepository {
	return &userRepository{
		storage: s,
	}
}

func (s *userRepository) Create(ctx context.Context, user *models.User) error {
	op := "userRepository.Create"
	query := "INSERT INTO users (id, nickname, key, fingerprint, register_time) VALUES ($1, $2, $3, $4, $5)"
	res, err := s.storage.Pool.Exec(ctx, query, user.ID, user.Nickname, user.Key, user.Fingerprint, user.RegisterTime)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return errs.ErrAlreadyExists(op, err)
		}
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (s *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	op := "userRepository.Delete"
	query := "DELETE FROM users WHERE id = $1"
	res, err := s.storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (s *userRepository) ExistenceCheck(ctx context.Context, nickname, key string) (bool, error) {
	op := "userRepository.ExistenceCheck"
	query := "SELECT 1 FROM users WHERE nickname = $1 and key = $2"
	var res int
	if err := s.storage.Pool.QueryRow(ctx, query, nickname, key).Scan(&res); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return false, errs.ErrNotFound(op)
		}
		return false, errs.NewAppError(op, err)
	}
	return res == 1, nil
}

func (s *userRepository) GetKey(ctx context.Context, nickname string) ([]byte, error) {
	op := "userRepository.GetKey"
	query := "SELECT key FROM users WHERE nickname = $1"
	var key string
	if err := s.storage.Pool.QueryRow(ctx, query, nickname).Scan(&key); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return []byte(key), nil
}
