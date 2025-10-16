package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	entities2 "github.com/Shelffy/shelffy/internal/entities"
	repositories2 "github.com/Shelffy/shelffy/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionIsExpired   = errors.New("session is expired")
)

type Auth interface {
	Login(ctx context.Context, email, password string) (entities2.User, entities2.Session, error)
	Logout(ctx context.Context, sessionID string) error
	GetSessionByID(ctx context.Context, id string) (entities2.Session, error)
	GetSessionByUserID(ctx context.Context, userID uuid.UUID) ([]entities2.Session, error)
	ValidateSession(ctx context.Context, id string) (entities2.Session, error)
	Deactivate(ctx context.Context, id string) error
	DeactivateByUserID(ctx context.Context, userID uuid.UUID, except ...string) error
}

type authService struct {
	userRepo        repositories2.Users
	sessionsRepo    repositories2.Session
	timeout         time.Duration
	sessionLifeTime time.Duration
	logger          *slog.Logger
	secret          string
}

func NewAuth(userRepo repositories2.Users, sessionRepo repositories2.Session, timeout time.Duration, sessionLifeTime time.Duration, logger *slog.Logger, secret string) Auth {
	return authService{
		userRepo:        userRepo,
		sessionsRepo:    sessionRepo,
		timeout:         timeout,
		sessionLifeTime: sessionLifeTime,
		logger:          logger,
		secret:          secret,
	}
}

func (s authService) generateSessionID() string {
	tokenBytes := [64]byte{}
	_, _ = rand.Read(tokenBytes[:])
	return hex.EncodeToString(tokenBytes[:])
}

func (s authService) ValidateSession(ctx context.Context, id string) (entities2.Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	session, err := s.sessionsRepo.GetByID(c, id)
	if err != nil {
		s.logger.Error("error while getting session", "error", err)
		return session, err
	}
	if !session.IsActive || session.ExpiresAt.Before(time.Now()) {
		return session, ErrSessionIsExpired
	}
	return session, nil
}

func (s authService) Login(ctx context.Context, email, password string) (entities2.User, entities2.Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	dbUser, err := s.userRepo.GetByEmail(c, email)
	if err != nil {
		s.logger.Error("cannot get user", "error", err)
		return entities2.User{}, entities2.Session{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return entities2.User{}, entities2.Session{}, ErrInvalidCredentials
	}
	c, cancel = context.WithTimeout(ctx, s.timeout)
	defer cancel()
	expiresAt := time.Now().Add(s.sessionLifeTime)
	session, err := s.sessionsRepo.Create(c, entities2.Session{
		UserID:    dbUser.ID,
		ID:        s.generateSessionID(),
		IsActive:  true,
		ExpiresAt: expiresAt,
	})
	return dbUser, session, err
}

func (s authService) Logout(ctx context.Context, id string) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := s.sessionsRepo.DeleteSession(c, id); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("error while deleting session", "error", err)
		return err
	}
	return nil
}

func (s authService) GetSessionByID(ctx context.Context, id string) (entities2.Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionsRepo.GetByID(c, id)
}

func (s authService) GetSessionByUserID(ctx context.Context, userID uuid.UUID) ([]entities2.Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionsRepo.GetByUserID(c, userID)
}

func (s authService) Deactivate(ctx context.Context, id string) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionsRepo.Deactivate(c, id)
}

func (s authService) DeactivateByUserID(ctx context.Context, userID uuid.UUID, except ...string) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.sessionsRepo.DeactivateByUserID(c, userID, except...)
}
