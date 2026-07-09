package app

import (
	"aloh-ssh/internal/config"
	"aloh-ssh/internal/domain/repositories"
	"aloh-ssh/internal/domain/services"
	"aloh-ssh/internal/server"
	"aloh-ssh/pkg/logger"
	"aloh-ssh/pkg/storage"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	path := os.Getenv("CONFIG_PATH")
	cfg := config.MustInitConfig(path)
	log := logger.NewLogger(cfg.App)
	log.Info("config initialized successfully")
	storage := storage.MustConnect(cfg.Storage)
	defer func() {
		storage.Close()
		log.Info("storage closed successfully")
	}()
	log.Info("storage connected successfully")
	userRepo := repositories.NewUserRepository(storage)
	log.Info("users repository created successfully")
	userService := services.NewUserService(userRepo, log)
	log.Info("users service created successfully")
	sessionRepo := repositories.NewSessionRepository()
	log.Info("sessions repository created successfully")
	sessionService := services.NewSessionService(sessionRepo, log)
	log.Info("sessions service created successfully")
	friendshipRepo := repositories.NewFriendshipRepository(storage)
	log.Info("friendship repository created successfully")
	friendshipService := services.NewFriendshipService(friendshipRepo, log)
	log.Info("friendship service created successfully")
	blockedRepo := repositories.NewBlockedRepository(storage)
	log.Info("blocked repository created successfully")
	blockedService := services.NewBlockedService(blockedRepo, log)
	log.Info("blocked service created successfully")
	server := server.NewServer(server.NewServerSetup{
		Cfg:               cfg.Server,
		SessionService:    sessionService,
		FriendshipService: friendshipService,
		UserService:       userService,
		BlockedService:    blockedService,
	})
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
		defer cancel()
		server.MustClose(ctx)
		log.Info("server closed successfully")
	}()
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		server.MustStart()
	}()
	log.Info("server sarted successfully")
	<-closeChan
	log.Info("app stopping...")
}
