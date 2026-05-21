package services

// import (
// 	"aloh-ssh/internal/domain/models"
// 	"aloh-ssh/internal/domain/repositories"
// 	"aloh-ssh/pkg/errs"
// 	"aloh-ssh/pkg/logger"
// 	"context"
// 	"time"

// 	"github.com/google/uuid"
// )

// type SessionService interface {
// 	NewSession(ctx context.Context, userID uuid.UUID) error
// 	DeleteSession(ctx context.Context, userID uuid.UUID) error
// 	GetSession(ctx context.Context, userID uuid.UUID) (*models.Session, error)
// }

// type sessionService struct {
// 	sessionRepo repositories.SessionRepository
// 	logger      *logger.Logger
// }

// func NewSessionService(sr repositories.SessionRepository, l *logger.Logger) SessionService {
// 	return &sessionService{
// 		sessionRepo: sr,
// 		logger:      l,
// 	}
// }

// func (ss *sessionService) NewSession(ctx context.Context, userID uuid.UUID) error {
// 	op := "sessionService.NewSession"
// 	log := ss.logger.AddOp(op)
// 	userIDLog := logger.Attr("userID", userID)
// 	log.Info("creating new session", userIDLog)
// 	session := &models.Session{
// 		UserID:      userID,
// 		CreatedTime: time.Now(),
// 	}
// 	if err := ss.sessionRepo.NewSession(ctx, session); err != nil {
// 		log.Error("failed to create new session", userIDLog, logger.Err(err))
// 		return errs.NewAppError(op, err)
// 	}
// 	log.Info("new session created successfully", userIDLog)
// 	return nil
// }

// func (ss *sessionService) DeleteSession(ctx context.Context, userID uuid.UUID) error {
// 	op := "sessionService.DeleteSession"
// 	log := ss.logger.AddOp(op)
// 	userIDLog := logger.Attr("userID", userID)
// 	log.Info("deleting session", userIDLog)
// 	if err := ss.sessionRepo.DeleteSession(ctx, userID); err != nil {
// 		log.Error("failed to delete session", userIDLog, logger.Err(err))
// 		return errs.NewAppError(op, err)
// 	}
// 	log.Info("session deleted successfully", userIDLog)
// 	return nil
// }

// func (ss *sessionService) GetSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
// 	op := "sessionService.GetSession"
// 	log := ss.logger.AddOp(op)
// 	userIDLog := logger.Attr("userID", userID)
// 	log.Info("receiving session", userIDLog)
// 	session, err := ss.sessionRepo.GetSession(ctx, userID)
// 	if err != nil {
// 		log.Error("failed to receive session", userIDLog, logger.Err(err))
// 		return nil, errs.NewAppError(op, err)
// 	}
// 	log.Info("session received successfully", userIDLog)
// 	return session, nil
// }
