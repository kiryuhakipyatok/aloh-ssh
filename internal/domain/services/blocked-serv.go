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
	BlockUser(ctx context.Context, userId uuid.UUID, blockedUserId string) (uuid.UUID, error)
	UnblockUser(ctx context.Context, userId uuid.UUID, unblockedUserNick string) error
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

func (bs *blockedService) BlockUser(ctx context.Context, userId uuid.UUID, blockedUserId string) (uuid.UUID, error) {
	op := "blockedService.BlockUser"
	log := bs.logger.AddOp(op)
	logUserId := logger.Attr("id", blockedUserId)
	log.Info("blocking user", logUserId)
	t := time.Now().UTC()
	friendId, err := bs.blockedRepository.BlockUser(ctx, userId, blockedUserId, t)
	if err != nil {
		log.Error("failed to block user", logUserId, logger.Err(err))
		return uuid.Nil, errs.NewAppError(op, err)
	}

	log.Info("user blocked successfully successfully", logUserId)

	return friendId, nil
}

func (bs *blockedService) UnblockUser(ctx context.Context, userId uuid.UUID, unblockedUserNick string) error {
	op := "blockedService.UnblockUser"
	log := bs.logger.AddOp(op)
	logUserId := logger.Attr("id", unblockedUserNick)
	log.Info("unblocking user", logUserId)

	if err := bs.blockedRepository.UnblockUser(ctx, userId, unblockedUserNick); err != nil {
		log.Error("failed to unblock user", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("user unblocked successfully", logUserId)

	return nil
}
