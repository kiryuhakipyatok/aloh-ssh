package repositories

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/storage"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, nickname string) error
	GetUser(ctx context.Context, nickname string) (*models.User, error)
	//ExistenceCheck(ctx context.Context, nickname, key string) (bool, error)
	SetPassword(ctx context.Context, nickname string, password []byte) error
	NewKeys(ctx context.Context, nickname string, key, fingerprint string) error
	GetPassword(ctx context.Context, nickname string) ([]byte, uuid.UUID, error)
	NewFriendRequest(ctx context.Context, userId uuid.UUID, nickname string, timeReq time.Time) (uuid.UUID, error)
	GetPersonalData(ctx context.Context, userId uuid.UUID) (*models.PersonalData, error)
	AcceptFriendship(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error)
	DenyFriendship(ctx context.Context, userId uuid.UUID, nickname string) error
	DeleteFromFriends(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error)
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

// func (s *userRepository) ExistenceCheck(ctx context.Context, nickname, key string) (bool, error) {
// 	op := "userRepository.ExistenceCheck"
// 	query := "SELECT 1 FROM users WHERE nickname = $1 and key = $2"
// 	var res int
// 	if err := s.storage.Pool.QueryRow(ctx, query, nickname, key).Scan(&res); err != nil {
// 		if errors.Is(err, storage.ErrNotFound()) {
// 			return false, errs.ErrNotFound(op)
// 		}
// 		return false, errs.NewAppError(op, err)
// 	}
// 	return res == 1, nil
// }

func (s *userRepository) GetUser(ctx context.Context, nickname string) (*models.User, error) {
	op := "userRepository.GetUser"
	query := "SELECT id, nickname, key, fingerprint, register_time FROM users WHERE nickname = $1"
	var user models.User
	if err := s.storage.Pool.QueryRow(ctx, query, nickname).Scan(
		&user.ID,
		&user.PersonalData.Nickname,
		&user.Key,
		&user.Fingerprint,
		&user.PersonalData.RegisterTime,
	); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return &user, nil
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

func (s *userRepository) NewKeys(ctx context.Context, nickname string, key, fingerprint string) error {
	op := "userRepository.NewKeys"
	query := "UPDATE users SET key=$1, fingerprint=$2 WHERE nickname=$3"
	res, err := s.storage.Pool.Exec(ctx, query, key, fingerprint, nickname)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (s *userRepository) GetPassword(ctx context.Context, nickname string) ([]byte, uuid.UUID, error) {
	op := "userRepository.GetPassword"
	query := "SELECT password FROM users WHERE nickname = $1"
	var res struct {
		pswrd []byte
		id    uuid.UUID
	}
	if err := s.storage.Pool.QueryRow(ctx, query, nickname).Scan(&res); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, uuid.UUID{}, errs.ErrNotFound(op)
		}
		return nil, uuid.UUID{}, errs.NewAppError(op, err)
	}
	return res.pswrd, res.id, nil
}

func (s *userRepository) GetPersonalData(ctx context.Context, userId uuid.UUID) (*models.PersonalData, error) {
	op := "userRepository.GetPersonalData"
	query := `SELECT u.nickname, u.register_time, 
    		  COALESCE(
              	json_agg(json_build_object(
                			'nickname', friend_u.nickname, 
                			'reqTime',  f.req_time
            			)) FILTER (WHERE f.user_id1 IS NOT NULL AND f.user_id2 = $1 AND f.status = 'pending'), '[]'
    		  ) AS friends_reqs,
			   COALESCE(
              	json_agg(friend_u.nickname) FILTER (WHERE f.user_id1 IS NOT NULL AND f.user_id2 IS NOT NULL AND f.status = 'active'), '[]'
    		  ) AS active_friends
			   FROM users u LEFT JOIN friends f ON (u.id = f.user_id1 OR u.id = f.user_id2)
			   LEFT JOIN users friend_u ON friend_u.id = (CASE WHEN f.user_id1 = u.id THEN f.user_id2 ELSE f.user_id1 END)
      		   WHERE u.id = $1 GROUP BY u.id, u.nickname, u.register_time`
	pd := models.PersonalData{}
	if err := s.storage.Pool.QueryRow(ctx, query, userId).Scan(
		&pd.Nickname,
		&pd.RegisterTime,
		&pd.FriendsReqs,
		&pd.Friends,
	); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return &pd, nil
}

// ON CONFLICT (LEAST(user_id1, user_id2), GREATEST(user_id1, user_id2))
//     			DO UPDATE SET status = 'active' WHERE friends.user_id2 = $1 AND friends.status = 'pending'

func (s *userRepository) NewFriendRequest(ctx context.Context, userId uuid.UUID, nickname string, reqTime time.Time) (uuid.UUID, error) {
	op := "userRepository.NewFriendRequest"
	query := `WITH found_user AS (
    			SELECT id FROM users WHERE nickname = $2 AND id != $1
			),
			inserted_friend AS (
    			INSERT INTO friends (user_id1, user_id2, req_time)
				SELECT $1, id, $3 FROM found_user
				RETURNING user_id2
			)
			SELECT user_id2 FROM inserted_friend`
	var id uuid.UUID
	err := s.storage.Pool.QueryRow(ctx, query, userId, nickname, reqTime).Scan(&id)
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

func (s *userRepository) AcceptFriendship(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "userRepository.AcceptFriendship"
	query := `UPDATE friends f SET status = 'active' FROM users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  f.user_id1 = u.id AND f.user_id2 = $1 RETURNING u.id`
	var id uuid.UUID
	err := s.storage.Pool.QueryRow(ctx, query, userId, nickname).Scan(&id)
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

func (s *userRepository) DenyFriendship(ctx context.Context, userId uuid.UUID, nickname string) error {
	op := "userRepository.DenyFriendship"
	query := `DELETE FROM friends f USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  f.user_id1 = u.id AND f.user_id2 = $1`
	res, err := s.storage.Pool.Exec(ctx, query, userId, nickname)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (s *userRepository) DeleteFromFriends(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "userRepository.DeleteFromFriends"
	query := `DELETE FROM friends f USING users u
			  WHERE u.nickname = $2 AND u.id != $1 AND
      		  (f.user_id1 = u.id AND f.user_id2 = $1) OR (f.user_id1 = $1 AND f.user_id2 = u.id) AND f.status = 'active'
			  RETURNING u.id`
	var id uuid.UUID
	err := s.storage.Pool.QueryRow(ctx, query, userId, nickname).Scan(&id)
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
