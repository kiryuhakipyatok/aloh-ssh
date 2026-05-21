package server

import (
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
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
		log.Info("new login request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		friendNickname := string(req.Payload)
		if err := s.userService.NewFriend(appCtx, userID, friendNickname); err != nil {
			log.Error("failed to add new friend request", logger.Err(err))
			return false, castErr(err)
		}
		log.Info("new friend request added successfully", logUserNickname)
		return true, nil
	}
}

// func (s *Server) proccessSessionMessagesRequests() ssh.RequestHandler {
// 	op := "server.proccessSessionMessagesRequests"
// 	log := s.log.AddOp(op)
// 	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
// 		nickname := ctx.User()
// 		logUserNickname := logger.Attr("nickname", nickname)
// 		log.Info("proccess session messages request", logUserNickname)
// 		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
// 		defer cancel()
// 		userID, ok := ctx.Value("userID").(uuid.UUID)
// 		if !ok {
// 			log.Error("failed to get user id", logUserNickname)
// 			return false, castErr(errs.ErrInvalidType(op))
// 		}
// 		userSession, err := s.sessionService.GetSession(appCtx, userID)
// 		if err != nil {
// 			log.Error("failed to get user's session", logger.Err(err), logUserNickname)
// 			return false, castErr(err)
// 		}
// 		for {
// 			select {
// 			case msg := <-userSession.MessagesChan:
// 				switch msg.Type{
// 				case models.NEW_FRIEND:

// 				}
// 			}
// 		}
// 		friendNickname := string(req.Payload)
// 		if err := s.userService.NewFriend(appCtx, userID, friendNickname); err != nil {
// 			log.Error("failed to add new friend request", logger.Err(err))
// 			return false, castErr(err)
// 		}
// 		log.Info("new friend request added successfully", logUserNickname)
// 		return true, nil
// 	}
// }
