package services

import (
	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/repositories"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/logger"
	"context"
	"time"

	"github.com/google/uuid"
)

type FriendshipService interface {
	NewFriend(ctx context.Context, userID, friendID uuid.UUID) error
	AcceptFriendship(ctx context.Context, userID, friendID uuid.UUID) error
	DenyFriendship(ctx context.Context, userID, friendID uuid.UUID) error
	DeleteFromFriends(ctx context.Context, userID, friendID uuid.UUID) error
}

type friendshipService struct {
	friendshipRepository repositories.FriendshipRepository
	logger               *logger.Logger
}

func NewFriendshipService(fr repositories.FriendshipRepository, l *logger.Logger) FriendshipService {
	return &friendshipService{
		friendshipRepository: fr,
		logger:               l,
	}
}

func (fs *friendshipService) NewFriend(ctx context.Context, userID, friendID uuid.UUID) error {
	op := "friendshipService.NewFriend"
	log := fs.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("additing new friend request", logUserId)
	t := time.Now().UTC()
	if err := fs.friendshipRepository.NewFriendRequest(ctx, userID, friendID, t); err != nil {
		log.Error("failed to add new friend request", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("new friend request added successfully", logUserId)

	return nil
}

func (fs *friendshipService) AcceptFriendship(ctx context.Context, userID, friendID uuid.UUID) error {
	op := "friendshipService.AcceptFriendship"
	log := fs.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("accpeting friendship", logUserId)

	if err := fs.friendshipRepository.AcceptFriendship(ctx, userID, friendID); err != nil {
		log.Error("failed to accpet friendship", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("friendship accepted successfully", logUserId)

	return nil
}
func (fs *friendshipService) DenyFriendship(ctx context.Context, userID, friendID uuid.UUID) error {
	op := "friendshipService.DenyFriendship"
	log := fs.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("denying friendship", logUserId)

	if err := fs.friendshipRepository.DenyFriendship(ctx, userID, friendID); err != nil {
		log.Error("failed to deny friendship", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("friendship denyed successfully", logUserId)

	return nil
}

func (fs *friendshipService) DeleteFromFriends(ctx context.Context, userID, friendID uuid.UUID) error {
	op := "friendshipService.DeleteFromFriends"
	log := fs.logger.AddOp(op)
	logUserId := logger.Attr("id", userID)
	log.Info("deleting from freinds", logUserId)

	if err := fs.friendshipRepository.DeleteFromFriends(ctx, userID, friendID); err != nil {
		log.Error("failed to delete from friends", logUserId, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("deleted from friends successfully", logUserId)

	return nil
}
