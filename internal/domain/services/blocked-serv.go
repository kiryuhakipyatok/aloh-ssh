package services

import (
	"context"
	"time"

	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/models"
	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/repositories"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/logger"

	"github.com/google/uuid"
)

type BlockedService interface {
	BlockUser(ctx context.Context, userId uuid.UUID, blockedUserId string) (uuid.UUID, error)
	UnblockUser(ctx context.Context, userId uuid.UUID, unblockedUserNick string) (uuid.UUID, error)
	FetchUsersBlockers(ctx context.Context, userId uuid.UUID) ([]models.Identity, error)
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
	logUserId := logger.Attr("id", userId)
	log.Info("blocking user", logUserId)
	t := time.Now().UTC()
	blockedId, err := bs.blockedRepository.BlockUser(ctx, userId, blockedUserId, t)
	if err != nil {
		log.Error("failed to block user", logUserId, logger.Err(err))
		return uuid.Nil, errs.NewAppError(op, err)
	}

	log.Info("user blocked successfully successfully", logUserId)

	return blockedId, nil
}

func (bs *blockedService) UnblockUser(ctx context.Context, userId uuid.UUID, unblockedUserNick string) (uuid.UUID, error) {
	op := "blockedService.UnblockUser"
	log := bs.logger.AddOp(op)
	logUserId := logger.Attr("id", userId)
	log.Info("unblocking user", logUserId)

	id, err := bs.blockedRepository.UnblockUser(ctx, userId, unblockedUserNick)
	if err != nil {
		log.Error("failed to unblock user", logUserId, logger.Err(err))
		return uuid.Nil, errs.NewAppError(op, err)
	}

	log.Info("user unblocked successfully", logUserId)

	return id, nil
}

func (bs *blockedService) FetchUsersBlockers(ctx context.Context, userId uuid.UUID) ([]models.Identity, error) {
	op := "blockedService.FetchUsersBlockers"
	log := bs.logger.AddOp(op)
	logUserId := logger.Attr("id", userId)
	log.Info("fetching users blockers", logUserId)

	blockers, err := bs.blockedRepository.FetchBlockedUsersById(ctx, userId)
	if err != nil {
		log.Error("failed to fetch users blockers", logUserId, logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	log.Info("users blockers fetched successfully", logUserId)

	return blockers, nil
}
