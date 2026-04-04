package services

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/internal/domain/repositories"
	"aloh-ssh/internal/utils"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type UserService interface {
	NewUser(ctx context.Context, nickname string, key ssh.PublicKey) error
	IsNew(ctx context.Context, nickname string, key ssh.PublicKey) (bool, error)
	GetUserKey(ctx context.Context, nickname string) ([]byte, error)
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
		ID:           id,
		Nickname:     nickname,
		Key:          keyString,
		Fingerprint:  fingerprint,
		RegisterTime: time.Now(),
	})

	if err := us.userRepository.Create(ctx, user); err != nil {
		log.Error("failed to create user", logUserNickname, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user registered successfully", logUserNickname)

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
