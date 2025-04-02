package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Service interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, user User) (User, error)
	Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (User, error)
	Deactivate(ctx context.Context, id uuid.UUID) (User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (User, error)
	ValidatePassword(password, passwordHash []byte) error
}

func NewService(repo Repository, timeout time.Duration, logger *slog.Logger) Service {
	return service{
		repository: repo,
		timeout:    timeout,
		logger:     logger,
	}
}

var ErrPasswordsNotMatch = errors.New("invalid password")

type service struct {
	repository Repository
	timeout    time.Duration
	logger     *slog.Logger
}

func (s service) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.GetByID(c, id)
}

func (s service) ValidatePassword(password, passwordHash []byte) error {
	return bcrypt.CompareHashAndPassword(passwordHash, password)
}

func (s service) GetByEmail(ctx context.Context, email string) (User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.GetByEmail(c, email)
}

func (s service) Create(ctx context.Context, user User) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	user.Password = string(hash)
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Create(c, user)
}

func (s service) Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Update(c, userID, fieldValue)
}

func (s service) Deactivate(ctx context.Context, id uuid.UUID) (User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Deactivate(c, id)
}

func (s service) Delete(ctx context.Context, id uuid.UUID) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Delete(c, id)
}

func (s service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	user, err := s.repository.GetByID(c, userID)
	defer cancel()
	if err != nil {
		return User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return User{}, ErrPasswordsNotMatch
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	c, cancel = context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Update(c, userID, map[string]any{"password": string(hash)})
}
