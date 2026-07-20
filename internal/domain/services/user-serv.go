package services

import (
	"context"
	"strings"
	"time"

	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/models"
	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/repositories"
	"github.com/kiryuhakipyatok/aloh-ssh/internal/utils"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type UserService interface {
	NewUser(ctx context.Context, nickname string, key ssh.PublicKey) (uuid.UUID, error)
	GetUsersFriends(ctx context.Context, id uuid.UUID) ([]models.Friend, error)
	GetUserByNickname(ctx context.Context, nickname string) (*models.User, error)
	SetPassword(ctx context.Context, id uuid.UUID, password []byte) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	SetNewKey(ctx context.Context, id uuid.UUID, key []byte) error
	CheckPassword(ctx context.Context, nickname string, password []byte) (uuid.UUID, error)
	GetPersonalData(ctx context.Context, id uuid.UUID) (*models.PersonalData, error)
	SetTagline(ctx context.Context, id uuid.UUID, tagline string) error
	SetColor(ctx context.Context, userID uuid.UUID, color string) error
	NewNickname(ctx context.Context, userID uuid.UUID, nickname string, password []byte) error
	NewPassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword []byte) error
}

type userService struct {
	userRepository repositories.UserRepository
	logger         *logger.Logger
}

func NewUserService(ur repositories.UserRepository, l *logger.Logger) UserService {
	return &userService{
		userRepository: ur,
		logger:         l,
	}
}

func (us *userService) NewUser(ctx context.Context, nickname string, key ssh.PublicKey) (uuid.UUID, error) {
	op := "userService.NewUser"

	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("new user registering", logUserNickname)

	keyBytes := ssh.MarshalAuthorizedKey(key)
	keyString := strings.TrimSpace(string(keyBytes))

	id := uuid.New()
	fingerprint, err := utils.GenerateFingerprint(keyBytes)
	if err != nil {
		log.Error("failed to generate finger print for user's key", logUserNickname, logger.Err(err))
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	user := new(models.User{
		PersonalData: models.PersonalData{
			Identity: models.Identity{
				ID:       id,
				Nickname: nickname,
			},
			RegisterTime: time.Now().UTC(),
			Tagline:      "",
		},
		Key:         keyString,
		Fingerprint: fingerprint,
	})

	if err := us.userRepository.Create(ctx, user); err != nil {
		log.Error("failed to create user", logUserNickname, logger.Err(err))
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	log.Info("user registered successfully", logUserNickname)

	return id, nil
}

func (us *userService) SetPassword(ctx context.Context, id uuid.UUID, password []byte) error {
	op := "userService.SetPassword"
	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", id)
	log.Info("adding password to user", logUserId)
	passwordHash, err := bcrypt.GenerateFromPassword(password, 12)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err), logUserId)
		return errs.NewAppError(op, err)
	}
	if err := us.userRepository.SetPassword(ctx, id, passwordHash); err != nil {
		log.Error("failed to set user's password", logger.Err(err), logUserId)
		return errs.NewAppError(op, err)
	}
	log.Info("user's password setted successfully", logUserId)
	return nil
}

func (us *userService) GetUserByNickname(ctx context.Context, nickname string) (*models.User, error) {
	op := "userService.GetUserByNickname"

	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("receiving user", logUserNickname)

	user, err := us.userRepository.GetUser(ctx, nickname)
	if err != nil {
		log.Error("failed to receive user", logUserNickname, logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	log.Info("user received successfully", logUserNickname)

	return user, nil
}

func (us *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	op := "userService.DeleteUser"

	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", id)
	log.Info("deleting user", logUserId)

	if err := us.userRepository.Delete(ctx, id); err != nil {
		log.Error("failed to delete user", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user deleted successfully", logUserId)

	return nil
}

func (us *userService) SetNewKey(ctx context.Context, id uuid.UUID, key []byte) error {
	op := "userService.SetNewKey"
	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", id)
	log.Info("setting new user's key", logUserId)

	keyString := strings.TrimSpace(string(key))
	fingerprint, err := utils.GenerateFingerprint([]byte(key))
	if err != nil {
		log.Error("failed to generatge fingerprint", logger.Err(err), logUserId)
		return errs.NewAppError(op, err)
	}
	if err := us.userRepository.NewKeys(ctx, id, keyString, fingerprint); err != nil {
		log.Error("failed to set new key to user", logger.Err(err), logUserId)
		return errs.NewAppError(op, err)
	}

	log.Info("new key setted successfylly", logUserId)

	return nil
}

func (us *userService) CheckPassword(ctx context.Context, nickname string, password []byte) (uuid.UUID, error) {
	op := "userService.CheckPassword"
	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("checking user's password", logUserNickname)
	userPassword, id, err := us.userRepository.GetPasswordByNickname(ctx, nickname)
	if err != nil {
		log.Error("failed to get user's password", logger.Err(err), logUserNickname)
		return uuid.Nil, errs.NewAppError(op, err)
	}
	if err := bcrypt.CompareHashAndPassword(userPassword, password); err != nil {
		log.Error("passwords are not equal", logger.Err(err), logUserNickname)
		return uuid.Nil, errs.NewAppError(op, err)
	}
	return id, nil
}

func (us *userService) GetPersonalData(ctx context.Context, userID uuid.UUID) (*models.PersonalData, error) {
	op := "userService.GetPersonalData"

	log := us.logger.AddOp(op)
	logUserId := logger.Attr("nickname", userID)
	log.Info("getting user's personal data", logUserId)

	personalData, err := us.userRepository.GetPersonalData(ctx, userID)
	if err != nil {
		log.Error("failed to get user's personal data", logUserId, logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	log.Info("user's personal data got successfully", logUserId)

	return personalData, nil
}

func (us *userService) GetUsersFriends(ctx context.Context, userID uuid.UUID) ([]models.Friend, error) {
	op := "userService.GetUsersFriends"

	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("getting user's friends", logUserId)

	friends, err := us.userRepository.GetUsersFriends(ctx, userID)
	if err != nil {
		log.Error("failed to get user's friends", logUserId, logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	log.Info("user's friends got successfully", logUserId)

	return friends, nil
}

func (us *userService) SetTagline(ctx context.Context, userID uuid.UUID, tagline string) error {
	op := "userService.SetTagline"

	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("setting user's tagline", logUserId)
	if err := us.userRepository.SetTagline(ctx, userID, tagline); err != nil {
		log.Error("failed to set tagline", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user's tagline setted successfully", logUserId)

	return nil
}

func (us *userService) SetColor(ctx context.Context, userID uuid.UUID, color string) error {
	op := "userService.SetColor"

	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("setting user's color", logUserId)
	if err := us.userRepository.SetColor(ctx, userID, color); err != nil {
		log.Error("failed to set color", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user's color setted successfully", logUserId)

	return nil
}

func (us *userService) NewNickname(ctx context.Context, userID uuid.UUID, nickname string, password []byte) error {
	op := "userService.NewNickname"

	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("setting new user's nickname", logUserId)
	userPassword, err := us.userRepository.GetPasswordById(ctx, userID)
	if err != nil {
		log.Error("failed to get user's password", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}
	if err := bcrypt.CompareHashAndPassword(userPassword, password); err != nil {
		log.Error("failed to compare passwords", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}
	if err := us.userRepository.EditNickname(ctx, userID, nickname); err != nil {
		log.Error("failed to edit nickname", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user's new nickname setted successfully", logUserId)

	return nil
}

func (us *userService) NewPassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword []byte) error {
	op := "userService.NewPassword"

	log := us.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("setting new user's password", logUserId)
	userPassword, err := us.userRepository.GetPasswordById(ctx, userID)
	if err != nil {
		log.Error("failed to get user's password", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}
	if err := bcrypt.CompareHashAndPassword(userPassword, oldPassword); err != nil {
		log.Error("failed to compare passwords", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}
	passwordHash, err := bcrypt.GenerateFromPassword(newPassword, 12)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err), logUserId)
		return errs.NewAppError(op, err)
	}
	if err := us.userRepository.SetPassword(ctx, userID, passwordHash); err != nil {
		log.Error("failed to set password", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user's new password setted successfully", logUserId)

	return nil
}
