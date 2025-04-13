package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plinkplenk/booki/internal/user"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionIsExpired   = errors.New("session is expired")
)

type Service interface {
	Login(ctx context.Context, email, password string) (user.User, Session, error)
	Logout(ctx context.Context, sessionID string) error
	GetSessionByID(ctx context.Context, id string) (Session, error)
	GetSessionByUserID(ctx context.Context, userID uuid.UUID) ([]Session, error)
	ValidateSession(ctx context.Context, id string) (Session, error)
	Deactivate(ctx context.Context, id string) error
	DeactivateByUserID(ctx context.Context, userID uuid.UUID, except ...string) error
}

type service struct {
	userRepo        user.Repository
	sessionRepo     Repository
	timeout         time.Duration
	sessionLifeTime time.Duration
	logger          *slog.Logger
	secret          string
}

func NewService(userRepo user.Repository, sessionRepo Repository, timeout time.Duration, sessionLifeTime time.Duration, logger *slog.Logger, secret string) Service {
	return service{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		timeout:         timeout,
		sessionLifeTime: sessionLifeTime,
		logger:          logger,
		secret:          secret,
	}
}

func (s service) generateSessionID() string {
	tokenBytes := [64]byte{}
	rand.Read(tokenBytes[:])
	return hex.EncodeToString(tokenBytes[:])
}

func (s service) ValidateSession(ctx context.Context, id string) (Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	session, err := s.sessionRepo.GetByID(c, id)
	if err != nil {
		s.logger.Error("error while getting session", "error", err)
		return session, err
	}
	if !session.IsActive || session.ExpiresAt.Before(time.Now()) {
		return session, ErrSessionIsExpired
	}
	return session, nil
}

func (s service) Login(ctx context.Context, email, password string) (user.User, Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	dbUser, err := s.userRepo.GetByEmail(c, email)
	if err != nil {
		s.logger.Error("cannot get user", "error", err)
		return user.User{}, Session{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return user.User{}, Session{}, ErrInvalidCredentials
	}
	c, cancel = context.WithTimeout(ctx, s.timeout)
	defer cancel()
	expiresAt := time.Now().Add(s.sessionLifeTime)
	session, err := s.sessionRepo.Create(c, Session{
		UserID:    dbUser.ID,
		ID:        s.generateSessionID(),
		IsActive:  true,
		ExpiresAt: expiresAt,
	})
	return dbUser, session, err
}

func (s service) Logout(ctx context.Context, id string) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := s.sessionRepo.DeleteSession(c, id); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("error while deleting session", "error", err)
		return err
	}
	return nil
}

func (s service) GetSessionByID(ctx context.Context, id string) (Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionRepo.GetByID(c, id)
}

func (s service) GetSessionByUserID(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionRepo.GetByUserID(c, userID)
}

func (s service) Deactivate(ctx context.Context, id string) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionRepo.Deactivate(c, id)
}

func (s service) DeactivateByUserID(ctx context.Context, userID uuid.UUID, except ...string) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionRepo.DeactivateByUserID(c, userID, except...)
}
