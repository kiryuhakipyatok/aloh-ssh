package server

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
	"encoding/json"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

func (s *Server) passwordRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.passwordRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new password request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := s.userService.AddPassword(appCtx, nickname, req.Payload); err != nil {
			log.Error("failed to add user's password", logger.Err(err), logUserNickname)
			if err := s.userService.DeleteUser(ctx, nickname); err != nil {
				log.Error("failed to delete user", logUserNickname, logger.Err(err))
			}
			return false, castErr(err)
		}
		log.Info("user's password added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) setNewKeyRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.setNewKeyRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new login request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		regTime, err := s.userService.SetNewKey(appCtx, nickname, req.Payload)
		if err != nil {
			log.Error("failed to set user's new keys", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		log.Info("user's key updated successfully", logUserNickname)
		return true, []byte(regTime)
	}
}

func (s *Server) newFriendRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.newFriendRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new friend request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		friendNickname := string(req.Payload)
		friendId, err := s.userService.NewFriend(appCtx, userID, friendNickname)
		if err != nil {
			log.Error("failed to add new friend request", logger.Err(err))
			return false, castErr(err)
		}
		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			log.Error("failed to get friend's session", logger.Err(err))
			return false, castErr(err)
		}
		newFriendEvent := models.NewFriendEvent(friendNickname)
		select {
		case friendSession.EventsChan <- newFriendEvent:
		default:
		}
		log.Info("new friend request added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) proccessEventChannel(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
	op := "server.proccessEventChannel"
	log := s.log.AddOp(op)
	channel, requests, err := newChan.Accept()
	if err != nil {
		log.Error("failed to accept channel")
		return
	}
	defer channel.Close()

	go gossh.DiscardRequests(requests)
	nickname := ctx.User()
	logUserNickname := logger.Attr("nickname", nickname)

	log.Info("event channel accepted successfully", logUserNickname)

	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		log.Error("failed to get user id", logUserNickname)
		return
	}

	defer func() {
		channel.Close()
		if err := s.sessionService.DeleteSession(context.Background(), userID); err != nil {
			s.log.Error("failed to delete session", logger.Err(err), logUserNickname)
		}
		s.log.Info("session deleted successfully", logUserNickname)
	}()

	userSession, err := s.sessionService.GetSession(ctx, userID)
	if err != nil {
		log.Error("failed to get user's session", logger.Err(err), logUserNickname)
		return
	}

	encoder := json.NewEncoder(channel)

	for {
		select {
		case <-ctx.Done():
			log.Info("context done", logUserNickname)
			return
		case event := <-userSession.EventsChan:
			if err := encoder.Encode(event); err != nil {
				log.Error("failed to encode evennt", logger.Err(err), logUserNickname)
				return
			}
		}
	}

}
