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
	serv           *ssh.Server
	userService    services.UserService
	sessionService services.SessionService
	log            *logger.Logger
}

const (
	LOGIN    = "SSH-2.0-aloh-login"
	REGISTER = "SSH-2.0-aloh-register"
	DEFAULT  = "SSH-2.0-aloh-default"
)

func NewServer(cfg config.Server, ss services.SessionService, us services.UserService, l *logger.Logger) *Server {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	s := &Server{
		userService:    us,
		sessionService: ss,
		log:            l,
	}
	server := &ssh.Server{
		Addr: addr,
		//Handler:          s.sessionHandler,
		PublicKeyHandler: s.publicKeyHandler(cfg.Timeout),
		RequestHandlers: map[string]ssh.RequestHandler{
			"pswrd":      s.passwordRequest(cfg.Timeout),
			"key":        s.setNewKeyRequest(cfg.Timeout),
			"new-friend": s.newFriendRequest(cfg.Timeout),
			// personal-data": s.fetchPersonalRequest(cfg.Timeout),
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"event-channel": s.proccessEventChannel,
		},
		PasswordHandler: s.passwordHandler(cfg.Timeout),
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
				log.Error("failed to get user's key", logger.Err(err), logUserNickname)
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
				if err := s.sessionService.NewSession(appCtx, user.ID); err != nil {
					log.Error("failed to create session", logger.Err(err), logUserNickname)
					return false
				}
			}
			return equal
		}

	}
}

// func (s *Server) sessionHandler(session ssh.Session) {
// 	go func() {
// 		nicknameLog := logger.Attr("user", session.User())

// 		s.log.Info("session closed", nicknameLog)
// 		userId, ok := session.Context().Value("userID").(uuid.UUID)
// 		if !ok {
// 			s.log.Error("failed to get user id", nicknameLog)
// 			return
// 		}
// 		defer func() {
// 			if err := s.sessionService.DeleteSession(context.Background(), userId); err != nil {
// 				s.log.Error("failed to delete session", logger.Err(err), nicknameLog)
// 			}
// 			s.log.Info("session deleted successfully", nicknameLog)
// 		}()
// 		<-session.Context().Done()
// 	}()

// }

// func (s *Server) fetchPersonalRequest(timeout time.Duration) ssh.RequestHandler {
// 	op := "server.fetchPersonalRequest"
// 	log := s.log.AddOp(op)
// 	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
// 		nickname := ctx.User()
// 		logUserNickname := logger.Attr("nickname", nickname)
// 		log.Info("new fetch personal data request", logUserNickname)
// 		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
// 		defer cancel()
// 		data, err := s.userService.GetPersonalData(appCtx, nickname)
// 		if err != nil {
// 			log.Error("failed to fetch personal data", logger.Err(err), logUserNickname)
// 			return false, castErr(err)
// 		}
// 		log.Info("user's key updated successfully", logUserNickname)
// 		return true, data
// 	}
// }

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
