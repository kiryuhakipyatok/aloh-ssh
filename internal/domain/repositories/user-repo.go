package repositories

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/storage"
	"context"
	"errors"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, nickname string) error
	GetKey(ctx context.Context, nickname string) ([]byte, error)
	ExistenceCheck(ctx context.Context, nickname, key string) (bool, error)
	SetPassword(ctx context.Context, nickname string, password []byte) error
	NewKeys(ctx context.Context, nickname string, key, fingerprint string) (string, error)
	GetPassword(ctx context.Context, nickname string) ([]byte, error)
	GetPersonalData(ctx context.Context, nickname string) (*models.PersonalData, error)
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
	query := "INSERT INTO users (id, nickname, password, key, fingerprint, register_time) VALUES ($1, $2, $3, $4, $5, $6)"
	res, err := s.storage.Pool.Exec(ctx, query, user.ID, user.PersonalData.Nickname, user.Password, user.Key, user.Fingerprint, user.PersonalData.RegisterTime)
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

func (s *userRepository) Delete(ctx context.Context, nickname string) error {
	op := "userRepository.Delete"
	query := "DELETE FROM users WHERE nickname = $1 AND register_time = $2"
	res, err := s.storage.Pool.Exec(ctx, query, nickname)
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

func (s *userRepository) SetPassword(ctx context.Context, nickname string, password []byte) error {
	op := "userRepository.Create"
	query := "UPDATE users SET password = $1 WHERE nickname = $2"
	res, err := s.storage.Pool.Exec(ctx, query, password, nickname)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (s *userRepository) NewKeys(ctx context.Context, nickname string, key, fingerprint string) (string, error) {
	op := "userRepository.NewKeys"
	query := "UPDATE users SET key=$1, fingerprint=$2 WHERE nickname=$3 RETURNING register_time"
	var regTime time.Time
	if err := s.storage.Pool.QueryRow(ctx, query, key, fingerprint, nickname).Scan(&regTime); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return "", errs.ErrNotFound(op)
		}
		return "", errs.NewAppError(op, err)
	}

	regTimeString := regTime.Format("2006-01-02")
	return regTimeString, nil
}

func (s *userRepository) GetPassword(ctx context.Context, nickname string) ([]byte, error) {
	op := "userRepository.GetPassword"
	query := "SELECT password FROM users WHERE nickname = $1"
	var res []byte
	if err := s.storage.Pool.QueryRow(ctx, query, nickname).Scan(&res); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return res, nil
}

func (s *userRepository) GetPersonalData(ctx context.Context, nickname string) (*models.PersonalData, error) {
	op := "userRepository.GetPersonalData"
	query := "SELECT nickname, register_time FROM users WHERE nickname = $1"
	pd := &models.PersonalData{}
	if err := s.storage.Pool.QueryRow(ctx, query, nickname).Scan(&pd.Nickname, &pd.RegisterTime); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return pd, nil
}
