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
	gossh "golang.org/x/crypto/ssh"
)

type Server struct {
	serv        *ssh.Server
	userService services.UserService
	log         *logger.Logger
}

const (
	LOGIN    = "SSH-2.0-aloh-login"
	REGISTER = "SSH-2.0-aloh-register"
	DEFAULT  = "SSH-2.0-aloh-default"
)

func NewServer(cfg config.Server, us services.UserService, l *logger.Logger) *Server {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	s := &Server{
		userService: us,
		log:         l,
	}
	server := &ssh.Server{
		Addr:             addr,
		PublicKeyHandler: s.publicKeyHandler(cfg.Timeout),
		RequestHandlers: map[string]ssh.RequestHandler{
			"pswrd":         s.requestPasswordHandler(cfg.Timeout),
			"key":           s.setNewKeyRequest(cfg.Timeout),
			// personal-data": s.fetchPersonalRequest(cfg.Timeout),
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
		fmt.Println(password)
		if err := s.userService.CheckPassword(appCtx, nickname, []byte(password)); err != nil {
			log.Error("failed to check password", logger.Err(err), logUserNickname)
			return false
		}
		return true
	}
}

func (s *Server) requestPasswordHandler(timeout time.Duration) ssh.RequestHandler {
	op := "server.requestPasswordHandler"
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
			if err := s.userService.NewUser(appCtx, nickname, key); err != nil {
				log.Error("failed to create new user", logger.Err(err), logUserNickname)
				return false
			}
			return true
		default:
			userKeyByte, err := s.userService.GetUserKey(ctx, nickname)
			if err != nil {
				log.Error("failed to get user's key", logger.Err(err), logUserNickname)
				return false
			}
			userKey, _, _, _, err := ssh.ParseAuthorizedKey(userKeyByte)
			if err != nil {
				log.Error("failed to parse user's key", logger.Err(err), logUserNickname)
				return false
			}
			return ssh.KeysEqual(userKey, key)
		}

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
