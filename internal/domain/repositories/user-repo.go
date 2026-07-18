package repositories

import (
	"context"
	"errors"

	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/models"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/storage"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetUser(ctx context.Context, nickname string) (*models.User, error)
	SetPassword(ctx context.Context, id uuid.UUID, password []byte) error
	NewKeys(ctx context.Context, id uuid.UUID, key, fingerprint string) error
	GetPassword(ctx context.Context, nickname string) ([]byte, uuid.UUID, error)
	GetPersonalData(ctx context.Context, id uuid.UUID) (*models.PersonalData, error)
	GetUsersFriends(ctx context.Context, id uuid.UUID) ([]models.Friend, error)
	SetTagline(ctx context.Context, id uuid.UUID, tagline string) error
	EditNickname(ctx context.Context, id uuid.UUID, nickname string) error
}

type userRepository struct {
	storage *storage.Storage
}

func NewUserRepository(ur *storage.Storage) UserRepository {
	return &userRepository{
		storage: ur,
	}
}

func (ur *userRepository) Create(ctx context.Context, user *models.User) error {
	op := "userRepository.Create"
	query := "INSERT INTO users (id, nickname, password, tagline, key, fingerprint, register_time) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	res, err := ur.storage.Pool.Exec(ctx, query,
		user.PersonalData.Identity.ID, user.PersonalData.Identity.Nickname, user.Password, user.PersonalData.Tagline,
		user.Key, user.Fingerprint, user.PersonalData.RegisterTime)
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

func (ur *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	op := "userRepository.Delete"
	query := "DELETE FROM users WHERE id = $1"
	res, err := ur.storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

// func (ur *userRepository) ExistenceCheck(ctx context.Context, nickname, key string) (bool, error) {
// 	op := "userRepository.ExistenceCheck"
// 	query := "SELECT 1 FROM users WHERE nickname = $1 and key = $2"
// 	var res int
// 	if err := ur.storage.Pool.QueryRow(ctx, query, nickname, key).Scan(&res); err != nil {
// 		if errors.Is(err, storage.ErrNotFound()) {
// 			return false, errs.ErrNotFound(op)
// 		}
// 		return false, errs.NewAppError(op, err)
// 	}
// 	return res == 1, nil
// }

func (ur *userRepository) GetUser(ctx context.Context, nickname string) (*models.User, error) {
	op := "userRepository.GetUser"
	query := `SELECT id, nickname, key, fingerprint, register_time FROM users WHERE nickname = $1`
	var user models.User
	if err := ur.storage.Pool.QueryRow(ctx, query, nickname).Scan(
		&user.PersonalData.Identity.ID,
		&user.PersonalData.Identity.Nickname,
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

func (ur *userRepository) GetUsersFriends(ctx context.Context, id uuid.UUID) ([]models.Friend, error) {
	op := "userRepository.GetActiveFriends"
	query := `
		SELECT u.id, u.nickname FROM friends f 
		JOIN users u ON u.id = (CASE WHEN f.user_id1 = $1 THEN f.user_id2 ELSE f.user_id1 END)
		WHERE (f.user_id1 = $1 OR f.user_id2 = $1) AND f.status = 'active'
	`
	rows, err := ur.storage.Pool.Query(ctx, query, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	defer rows.Close()

	var friends []models.Friend
	for rows.Next() {
		var f models.Friend
		if err := rows.Scan(&f.ID, &f.Nickname); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		friends = append(friends, f)
	}

	if err := rows.Err(); err != nil {
		return nil, errs.NewAppError(op, err)
	}

	return friends, nil
}

func (ur *userRepository) GetPersonalData(ctx context.Context, id uuid.UUID) (*models.PersonalData, error) {
	op := "userRepository.GetPersonalData"
	query := `SELECT u.id, u.nickname, u.register_time, u.tagline,
    		  COALESCE((
              	SELECT json_agg(json_build_object(
							'identity', json_build_object(
            					'id', sender.id,
            					'nickname', sender.nickname
        					),
                			'reqTime',  f.req_time
            			)) FROM friends f JOIN users sender ON sender.id = f.user_id1 WHERE f.user_id2 = u.id
						AND f.status = 'pending'), '[]'
    		  ) AS friends_reqs,
			   COALESCE((
				SELECT json_agg(json_build_object(
        					'identity', json_build_object(
            					'id', friend_resolv.id,
            					'nickname', friend_resolv.nickname
        					),
        					'appereance', json_build_object(
            					'tagline', friend_resolv.tagline
        					)
    					))
				FROM friends f JOIN users friend_resolv ON 
				friend_resolv.id = (CASE WHEN f.user_id1 = u.id THEN f.user_id2 ELSE f.user_id1 END)
                WHERE (f.user_id1 = u.id OR f.user_id2 = u.id) AND f.status = 'active'), '[]'
    		  ) AS active_friends,
			   COALESCE((SELECT json_agg(json_build_object(
            					'id', bu.blocked_id,
            					'nickname', bn.nickname
        					)) FROM blocked_users bu JOIN users bn 
							ON bn.id = bu.blocked_id
							WHERE blocker_id = u.id), '[]') 
			   AS blocked_users
			   FROM users u WHERE u.id = $1`
	pd := models.PersonalData{}
	if err := ur.storage.Pool.QueryRow(ctx, query, id).Scan(
		&pd.Identity.ID,
		&pd.Identity.Nickname,
		&pd.RegisterTime,
		&pd.Tagline,
		&pd.FriendsReqs,
		&pd.Friends,
		&pd.BlockedUsers,
	); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return &pd, nil
}

func (ur *userRepository) SetPassword(ctx context.Context, id uuid.UUID, password []byte) error {
	op := "userRepository.Create"
	query := "UPDATE users SET password = $1 WHERE id = $2"
	res, err := ur.storage.Pool.Exec(ctx, query, password, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (ur *userRepository) NewKeys(ctx context.Context, id uuid.UUID, key, fingerprint string) error {
	op := "userRepository.NewKeys"
	query := "UPDATE users SET key=$1, fingerprint=$2 WHERE id=$3"
	res, err := ur.storage.Pool.Exec(ctx, query, key, fingerprint, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (ur *userRepository) GetPassword(ctx context.Context, nickname string) ([]byte, uuid.UUID, error) {
	op := "userRepository.GetPassword"
	query := "SELECT id, password FROM users WHERE nickname = $1"
	var res struct {
		id    uuid.UUID
		pswrd []byte
	}
	if err := ur.storage.Pool.QueryRow(ctx, query, nickname).Scan(
		&res.id,
		&res.pswrd,
	); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, uuid.UUID{}, errs.ErrNotFound(op)
		}
		return nil, uuid.UUID{}, errs.NewAppError(op, err)
	}
	return res.pswrd, res.id, nil
}

func (ur *userRepository) SetTagline(ctx context.Context, id uuid.UUID, tagline string) error {
	op := "userRepository.SetTalgile"
	query := "UPDATE users SET tagline=$1 WHERE id=$2"
	res, err := ur.storage.Pool.Exec(ctx, query, tagline, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (ur *userRepository) EditNickname(ctx context.Context, id uuid.UUID, nickname string) error {
	op := "userRepository.SetTalgile"
	query := "UPDATE users SET nickname=$1 WHERE id=$2"
	res, err := ur.storage.Pool.Exec(ctx, query, nickname, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}
