package server

import (
	"aloh-ssh/internal/config"
	"aloh-ssh/internal/domain/services"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/ssh"
)

type Server struct {
	serv *ssh.Server
}

type handlerFunc = func(ctx ssh.Context, key ssh.PublicKey) bool

func NewServer(cfg config.Server, us services.UserService, l *logger.Logger) *Server {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	server := &ssh.Server{
		Addr:             addr,
		PublicKeyHandler: handler(cfg.Timeout, us, l),
	}
	return &Server{
		serv: server,
	}
}

func handler(timeout time.Duration, us services.UserService, l *logger.Logger) handlerFunc {
	op := "server.Handler"
	log := l.AddOp(op)
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new connect", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		userKeyByte, err := us.GetUserKey(ctx, nickname)
		if err != nil {
			if errors.Is(err, errs.ErrNotFoundBase) {
				log.Info("new user", logUserNickname)
				if err := us.NewUser(appCtx, nickname, key); err != nil {
					log.Error("failed to create new user", logger.Err(err))
					return false
				}
				return true
			}
			log.Error("failed to get user's key", logger.Err(err))
			return false
		}

		userKey, _, _, _, err := ssh.ParseAuthorizedKey(userKeyByte)
		if err != nil {
			log.Error("failed to parse user's key", logger.Err(err))
			return false
		}

		return ssh.KeysEqual(userKey, key)
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
