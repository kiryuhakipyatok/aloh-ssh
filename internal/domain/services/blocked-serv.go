package services

import (
	"aloh-ssh/internal/domain/repositories"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
	"time"

	"github.com/google/uuid"
)

type BlockedService interface {
	BlockUser(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error)
	UnblockUser(ctx context.Context, userId uuid.UUID, nickname string) error
}

type blockedService struct {
	blockedRepository repositories.BlockedRepository
	logger            *logger.Logger
}

func NewBlockedService(br repositories.BlockedRepository, l *logger.Logger) BlockedService {
	return &blockedService{
		blockedRepository: br,
		logger:            l,
	}
}

func (bs *blockedService) BlockUser(ctx context.Context, userId uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "blockedService.BlockUser"
	log := bs.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("blocking user", logUserNickname)
	t := time.Now().UTC()
	friendId, err := bs.blockedRepository.BlockUser(ctx, userId, nickname, t)
	if err != nil {
		log.Error("failed to block user", logUserNickname, logger.Err(err))
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	log.Info("user blocked successfully successfully", logUserNickname)

	return friendId, nil
}

func (bs *blockedService) UnblockUser(ctx context.Context, userId uuid.UUID, nickname string) error {
	op := "blockedService.UnblockUser"
	log := bs.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("unblocking user", logUserNickname)

	if err := bs.blockedRepository.UnblockUser(ctx, userId, nickname); err != nil {
		log.Error("failed to unblock user", logUserNickname, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user unblocked successfully", logUserNickname)

	return nil
}
