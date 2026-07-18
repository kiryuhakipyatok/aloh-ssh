package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/kiryuhakipyatok/aloh-ssh/internal/config"
	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/services"

	"github.com/charmbracelet/ssh"
)

type Server struct {
	serv             *ssh.Server
	userService      services.UserService
	sessionService   services.SessionService
	friendshipSerive services.FriendshipService
	blockedService   services.BlockedService
	cfg              config.Server
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
}

func NewServer(nss NewServerSetup) *Server {
	addr := fmt.Sprintf("%s:%s", nss.Cfg.Host, nss.Cfg.Port)
	s := &Server{
		userService:      nss.UserService,
		sessionService:   nss.SessionService,
		friendshipSerive: nss.FriendshipService,
		blockedService:   nss.BlockedService,
		cfg:              nss.Cfg,
	}
	server := &ssh.Server{
		Addr:             addr,
		PublicKeyHandler: s.publicKeyHandler(),
		RequestHandlers: map[string]ssh.RequestHandler{
			"pswrd":         s.passwordRequest(),
			"key":           s.setNewKeyRequest(),
			"new-friend":    s.newFriendRequest(),
			"accept-friend": s.acceptFriendshipRequest(),
			"deny-friend":   s.denyFriendshipRequest(),
			"personal-data": s.fetchPersonalRequest(),
			"delete-friend": s.deleteFromFriendsRequest(),
			"block-user":    s.blockUserRequest(),
			"unblock-user":  s.unblockUserRequest(),
			"conns-update":  s.updateCurOnlineRequest(),
			"set-tagline":   s.setTaglineRequest(),
			"new-nickname":  s.newNicknameRequest(),
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"event-channel": s.proccessEventChannel,
		},
		PasswordHandler: s.passwordHandler(),
		IdleTimeout:     s.cfg.IdleTimeout,
	}

	s.serv = server
	return s
}

func (s *Server) passwordHandler() ssh.PasswordHandler {
	return func(ctx ssh.Context, password string) bool {
		if ctx.ClientVersion() != LOGIN {
			return false
		}
		nickname := ctx.User()
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		id, err := s.userService.CheckPassword(appCtx, nickname, []byte(password))
		if err != nil {
			return false
		}
		ctx.SetValue("userID", id)

		return true
	}
}

func (s *Server) publicKeyHandler() ssh.PublicKeyHandler {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		nickname := ctx.User()

		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		switch ctx.ClientVersion() {
		case REGISTER:
			id, err := s.userService.NewUser(appCtx, nickname, key)
			if err != nil {
				return false
			}
			ctx.SetValue("userID", id)
			return true
		default:
			user, err := s.userService.GetUserByNickname(ctx, nickname)
			if err != nil {
				return false
			}
			userKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(user.Key))
			if err != nil {
				return false
			}
			equal := ssh.KeysEqual(userKey, key)
			if equal {
				ctx.SetValue("userID", user.PersonalData.Identity.ID)
			}
			return equal
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
