package server

import (
	"aloh-ssh/internal/config"
	"aloh-ssh/internal/domain/services"
	"aloh-ssh/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/ssh"
)

type Server struct {
	serv             *ssh.Server
	userService      services.UserService
	sessionService   services.SessionService
	friendshipSerive services.FriendshipService
	blockedService   services.BlockedService
	cfg              config.Server
	log              *logger.Logger
}

const (
	LOGIN    = "SSH-2.0-aloh-login"
	REGISTER = "SSH-2.0-aloh-register"
	DEFAULT  = "SSH-2.0-aloh-default"
)

type NewServerSetup struct {
	Cfg               config.Server
	SessionService    services.SessionService
	UserService       services.UserService
	FriendshipService services.FriendshipService
	BlockedService    services.BlockedService
	Log               *logger.Logger
}

func NewServer(nss NewServerSetup) *Server {
	addr := fmt.Sprintf("%s:%s", nss.Cfg.Host, nss.Cfg.Port)
	s := &Server{
		userService:      nss.UserService,
		sessionService:   nss.SessionService,
		friendshipSerive: nss.FriendshipService,
		blockedService:   nss.BlockedService,
		cfg:              nss.Cfg,
		log:              nss.Log,
	}
	server := &ssh.Server{
		Addr:             addr,
		PublicKeyHandler: s.publicKeyHandler(s.cfg.Timeout),
		RequestHandlers: map[string]ssh.RequestHandler{
			"pswrd":         s.passwordRequest(s.cfg.Timeout),
			"key":           s.setNewKeyRequest(s.cfg.Timeout),
			"new-friend":    s.newFriendRequest(s.cfg.Timeout),
			"accept-friend": s.acceptFriendshipRequest(s.cfg.Timeout),
			"deny-friend":   s.denyFriendshipRequest(s.cfg.Timeout),
			"personal-data": s.fetchPersonalRequest(s.cfg.Timeout),
			"delete-friend": s.deleteFromFriendsRequest(s.cfg.Timeout),
			"block-user":    s.blockUserRequest(s.cfg.Timeout),
			"unblock-user":  s.unblockUserRequest(s.cfg.Timeout),
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"event-channel": s.proccessEventChannel,
		},
		PasswordHandler: s.passwordHandler(s.cfg.Timeout),
		IdleTimeout:     s.cfg.IdleTimeout,
	}

	s.serv = server
	return s
}

func (s *Server) passwordHandler(timeout time.Duration) ssh.PasswordHandler {
	op := "server.passwordHandler"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, password string) bool {
		if ctx.ClientVersion() != LOGIN {
			return false
		}
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		id, err := s.userService.CheckPassword(appCtx, nickname, []byte(password))
		if err != nil {
			log.Error("failed to check password", logger.Err(err), logUserNickname)
			return false
		}
		ctx.SetValue("userID", id)
		// user, err := s.userService.GetUser(ctx, nickname)
		// if err != nil {
		// 	log.Error("failed to get user", logger.Err(err), logUserNickname)
		// 	return false
		// }
		// for _, friend := range user.PersonalData.Friends {
		// 	go func() {
		// 		friendSession, err := s.sessionService.GetSession(appCtx, friend.ID)
		// 		if err != nil {
		// 			if errors.Is(err, errs.ErrNotFoundBase) {
		// 				log.Info("friend is offline", logUserNickname)
		// 			} else {
		// 				log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
		// 			}
		// 			return
		// 		}
		// 		friendOnlineEvent := models.FriendOnlineEvent(nickname)
		// 		select {
		// 		case friendSession.EventsChan <- friendOnlineEvent:
		// 		default:
		// 		}

		// 	}()
		// }
		return true
	}
}

func (s *Server) publicKeyHandler(timeout time.Duration) ssh.PublicKeyHandler {
	op := "server.publicKeyHandler"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new connect", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		switch ctx.ClientVersion() {
		case REGISTER:
			id, err := s.userService.NewUser(appCtx, nickname, key)
			if err != nil {
				log.Error("failed to create new user", logger.Err(err), logUserNickname)
				return false
			}
			ctx.SetValue("userID", id)
			if err := s.sessionService.NewSession(appCtx, id); err != nil {
				log.Error("failed to create session", logger.Err(err), logUserNickname)
				return false
			}
			return true
		default:
			user, err := s.userService.GetUser(ctx, nickname)
			if err != nil {
				log.Error("failed to get user", logger.Err(err), logUserNickname)
				return false
			}
			userKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(user.Key))
			if err != nil {
				log.Error("failed to parse user's key", logger.Err(err), logUserNickname)
				return false
			}
			equal := ssh.KeysEqual(userKey, key)
			if equal {
				ctx.SetValue("userID", user.ID)
			

				// for _, friend := range user.PersonalData.Friends {
				// 	go func() {
				// 		friendSession, err := s.sessionService.GetSession(appCtx, friend.ID)
				// 		if err != nil {
				// 			if errors.Is(err, errs.ErrNotFoundBase) {
				// 				log.Info("friend is offline", logUserNickname)
				// 			} else {
				// 				log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
				// 			}
				// 			return
				// 		}
				// 		friendOnlineEvent := models.FriendOnlineEvent(nickname)
				// 		select {
				// 		case friendSession.EventsChan <- friendOnlineEvent:
				// 			log.Info("online event sended successfully", logUserNickname)
				// 		default:
				// 		}

				// 	}()
				// }
			}
			return ssh.KeysEqual(userKey, key)
		}

	}
}

func (s *Server) MustStart() {
	err := s.serv.ListenAndServe()
	if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		panic(fmt.Errorf("failed to listen and serve: %w", err))
	}
}

func (s *Server) MustClose(ctx context.Context) {
	if err := s.serv.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("failed to close server: %w", err))
	}
}
