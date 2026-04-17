package services

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/internal/domain/repositories"
	"aloh-ssh/internal/utils"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type UserService interface {
	NewUser(ctx context.Context, nickname string, key ssh.PublicKey) error
	IsNew(ctx context.Context, nickname string, key ssh.PublicKey) (bool, error)
	GetUserKey(ctx context.Context, nickname string) ([]byte, error)
	AddPassword(ctx context.Context, nickname string, password []byte) error
	DeleteUser(ctx context.Context, nickname string) error
	SetNewKey(ctx context.Context, nickname string, key []byte) (string, error)
	CheckPassword(ctx context.Context, nickname string, password []byte) error
	GetPersonalData(ctx context.Context, nickname string) ([]byte, error)
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

func (us *userService) NewUser(ctx context.Context, nickname string, key ssh.PublicKey) error {
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
		return errs.NewAppError(op, err)
	}

	user := new(models.User{
		ID: id,
		PersonalData: models.PersonalData{
			Nickname:     nickname,
			RegisterTime: time.Now(),
		},
		Key:         keyString,
		Fingerprint: fingerprint,
	})

	if err := us.userRepository.Create(ctx, user); err != nil {
		log.Error("failed to create user", logUserNickname, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user registered successfully", logUserNickname)

	return nil
}

func (us *userService) AddPassword(ctx context.Context, nickname string, password []byte) error {
	op := "userService.AddPassword"
	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("adding password to user", logUserNickname)
	passwordHash, err := bcrypt.GenerateFromPassword(password, 12)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err), logUserNickname)
		return errs.NewAppError(op, err)
	}
	if err := us.userRepository.SetPassword(ctx, nickname, passwordHash); err != nil {
		log.Error("failed to set user's password", logger.Err(err), logUserNickname)
		return errs.NewAppError(op, err)
	}
	log.Info("user's password setted successfully", logUserNickname)
	return nil
}

func (us *userService) IsNew(ctx context.Context, nickname string, key ssh.PublicKey) (bool, error) {
	op := "userService.IsNew"

	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("user is new checking", logUserNickname)

	keyBytes := ssh.MarshalAuthorizedKey(key)
	keyString := strings.TrimSpace(string(keyBytes))

	res, err := us.userRepository.ExistenceCheck(ctx, nickname, keyString)
	if err != nil {
		log.Error("failed to check user's existance", logUserNickname, logger.Err(err))
		return false, errs.NewAppError(op, err)
	}

	log.Info("user's existance checked successfully", logUserNickname)

	return res, nil
}

func (us *userService) GetUserKey(ctx context.Context, nickname string) ([]byte, error) {
	op := "userService.IsNew"

	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("getting user's key", logUserNickname)

	key, err := us.userRepository.GetKey(ctx, nickname)
	if err != nil {
		log.Error("failed to get user's key", logUserNickname, logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	log.Info("user's key got successfully", logUserNickname)

	return key, nil
}

func (us *userService) DeleteUser(ctx context.Context, nickname string) error {
	op := "userService.DeleteUser"

	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("deleting user", logUserNickname)

	if err := us.userRepository.Delete(ctx, nickname); err != nil {
		log.Error("failed to delete user", logUserNickname, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user deleted successfully", logUserNickname)

	return nil
}

func (us *userService) SetNewKey(ctx context.Context, nickname string, key []byte) (string, error) {
	op := "userService.SetNewKey"
	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("setting new user's key", logUserNickname)

	keyString := strings.TrimSpace(string(key))
	fingerprint, err := utils.GenerateFingerprint([]byte(key))
	if err != nil {
		log.Error("failed to generatge fingerprint", logger.Err(err), logUserNickname)
	}
	regTime, err := us.userRepository.NewKeys(ctx, nickname, keyString, fingerprint)
	if err != nil {
		log.Error("failed to set new key to user", logger.Err(err), logUserNickname)
		return "", errs.NewAppError(op, err)
	}

	log.Info("new key setted successfylly", logUserNickname)

	return regTime, nil
}

func (us *userService) CheckPassword(ctx context.Context, nickname string, password []byte) error {
	op := "userService.CheckPassword"
	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("checking user's password", logUserNickname)
	userPassword, err := us.userRepository.GetPassword(ctx, nickname)
	if err != nil {
		log.Error("failed to get user's password", logger.Err(err), logUserNickname)
		return errs.NewAppError(op, err)
	}
	if err := bcrypt.CompareHashAndPassword(userPassword, password); err != nil {
		log.Error("passwords are not equal", logger.Err(err), logUserNickname)
		return errs.NewAppError(op, err)
	}
	return nil
}

func (us *userService) GetPersonalData(ctx context.Context, nickname string) ([]byte, error) {
	op := "userService.GetPersonalData"

	log := us.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("getting user's personal data", logUserNickname)

	personalData, err := us.userRepository.GetPersonalData(ctx, nickname)
	if err != nil {
		log.Error("failed to get user's personal data", logUserNickname, logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	personalDataBytes, err := json.Marshal(personalData)
	if err != nil {
		log.Error("failed to marshal user's personal data", logUserNickname, logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	log.Info("user's personal data got successfully", logUserNickname)

	return personalDataBytes, nil
}
