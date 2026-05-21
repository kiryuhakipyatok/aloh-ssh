package repositories

// import (
// 	"aloh-ssh/internal/domain/models"
// 	"aloh-ssh/pkg/errs"
// 	"context"
// 	"sync"

// 	"github.com/google/uuid"
// )

// type SessionRepository interface {
// 	NewSession(ctx context.Context, session *models.Session) error
// 	DeleteSession(ctx context.Context, sessionId uuid.UUID) error
// 	GetSession(ctx context.Context, sessionId uuid.UUID) (*models.Session, error)
// }

// type sessionRepository struct {
// 	sync.Map
// }

// func NewSessionRepository() SessionRepository {
// 	return &sessionRepository{}
// }

// func (sr *sessionRepository) NewSession(ctx context.Context, session *models.Session) error {
// 	op := "sessionRepository.NewSession"
// 	select {
// 	case <-ctx.Done():
// 		return errs.ErrRequestTimeout(op)
// 	default:
// 		if _, ok := sr.LoadOrStore(session.UserID, session); ok {
// 			return errs.ErrAlreadyExists(op, nil)
// 		}
// 		return nil
// 	}
// }

// func (sr *sessionRepository) DeleteSession(ctx context.Context, sessionId uuid.UUID) error {
// 	op := "sessionRepository.NewSession"
// 	select {
// 	case <-ctx.Done():
// 		return errs.ErrRequestTimeout(op)
// 	default:
// 		if _, ok := sr.LoadAndDelete(sessionId); !ok {
// 			return errs.ErrNotFound(op)
// 		}
// 		return nil
// 	}
// }

// func (sr *sessionRepository) GetSession(ctx context.Context, sessionId uuid.UUID) (*models.Session, error) {
// 	op := "sessionRepository.NewSession"
// 	select {
// 	case <-ctx.Done():
// 		return nil, errs.ErrRequestTimeout(op)
// 	default:
// 		val, ok := sr.Load(sessionId)
// 		if !ok {
// 			return nil, errs.ErrNotFound(op)
// 		}
// 		session, ok := val.(*models.Session)
// 		if !ok {
// 			return nil, errs.ErrInvalidType(op)
// 		}
// 		return session, nil
// 	}
// }
