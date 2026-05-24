package services

import (
	"aloh-ssh/internal/domain/repositories"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
	"time"

	"github.com/google/uuid"
)

type FriendshipService interface {
	NewFriend(ctx context.Context, userID uuid.UUID, nickname string) (uuid.UUID, error)
	AcceptFriendship(ctx context.Context, userID uuid.UUID, nickname string) (uuid.UUID, error)
	DenyFriendship(ctx context.Context, userID uuid.UUID, nickname string) error
	DeleteFromFriends(ctx context.Context, userID uuid.UUID, nickname string) (uuid.UUID, error)
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

func (fs *friendshipService) NewFriend(ctx context.Context, userID uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "friendshipService.NewFriend"
	log := fs.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("additing new friend request", logUserNickname)
	t := time.Now().UTC()
	friendId, err := fs.friendshipRepository.NewFriendRequest(ctx, userID, nickname, t)
	if err != nil {
		log.Error("failed to add new friend request", logUserNickname, logger.Err(err))
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	log.Info("new friend request added successfully", logUserNickname)

	return friendId, nil
}

func (fs *friendshipService) AcceptFriendship(ctx context.Context, userID uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "friendshipService.AcceptFriendship"
	log := fs.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("accpeting friendship", logUserNickname)

	id, err := fs.friendshipRepository.AcceptFriendship(ctx, userID, nickname)
	if err != nil {
		log.Error("failed to accpet friendship", logUserNickname, logger.Err(err))
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	log.Info("friendship accepted successfully", logUserNickname)

	return id, nil
}
func (fs *friendshipService) DenyFriendship(ctx context.Context, userID uuid.UUID, nickname string) error {
	op := "friendshipService.DenyFriendship"
	log := fs.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("denying friendship", logUserNickname)

	if err := fs.friendshipRepository.DenyFriendship(ctx, userID, nickname); err != nil {
		log.Error("failed to deny friendship", logUserNickname, logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("friendship denyed successfully", logUserNickname)

	return nil
}

func (fs *friendshipService) DeleteFromFriends(ctx context.Context, userID uuid.UUID, nickname string) (uuid.UUID, error) {
	op := "friendshipService.DeleteFromFriends"
	log := fs.logger.AddOp(op)
	logUserNickname := logger.Attr("nickname", nickname)
	log.Info("deleting from freinds", logUserNickname)

	id, err := fs.friendshipRepository.DeleteFromFriends(ctx, userID, nickname)
	if err != nil {
		log.Error("failed to delete from friends", logUserNickname, logger.Err(err))
		return uuid.UUID{}, errs.NewAppError(op, err)
	}

	log.Info("deleted from friends successfully", logUserNickname)

	return id, nil
}
