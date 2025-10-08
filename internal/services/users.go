package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/Shelffy/shelffy/internal/repositories"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Users interface {
	GetByID(ctx context.Context, id uuid.UUID) (entities.User, error)
	GetByEmail(ctx context.Context, email string) (entities.User, error)
	Create(ctx context.Context, user entities.User) (entities.User, error)
	Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (entities.User, error)
	Deactivate(ctx context.Context, id uuid.UUID) (entities.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (entities.User, error)
	ValidatePassword(password, passwordHash []byte) error
}

func NewUsers(repo repositories.Users, timeout time.Duration, logger *slog.Logger) Users {
	return usersService{
		repository: repo,
		timeout:    timeout,
		logger:     logger,
	}
}

var ErrPasswordsNotMatch = errors.New("invalid password")

type usersService struct {
	repository repositories.Users
	timeout    time.Duration
	logger     *slog.Logger
}

func (s usersService) GetByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.GetByID(c, id)
}

func (s usersService) ValidatePassword(password, passwordHash []byte) error {
	return bcrypt.CompareHashAndPassword(passwordHash, password)
}

func (s usersService) GetByEmail(ctx context.Context, email string) (entities.User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.GetByEmail(c, email)
}

func (s usersService) Create(ctx context.Context, user entities.User) (entities.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return entities.User{}, err
	}
	user.Password = string(hash)
	user.ID = uuid.New()
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	newUser, err := s.repository.Create(c, user)
	if err != nil {
		s.logger.Error("failed to create user", "error", err)
		s.logger.Debug("", "user", user)
		return entities.User{}, err
	}
	return newUser, err
}

func (s usersService) Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (entities.User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Update(c, userID, fieldValue)
}

func (s usersService) Deactivate(ctx context.Context, id uuid.UUID) (entities.User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Deactivate(c, id)
}

func (s usersService) Delete(ctx context.Context, id uuid.UUID) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Delete(c, id)
}

func (s usersService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (entities.User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	user, err := s.repository.GetByID(c, userID)
	defer cancel()
	if err != nil {
		return entities.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return entities.User{}, ErrPasswordsNotMatch
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return entities.User{}, err
	}
	c, cancel = context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.repository.Update(c, userID, map[string]any{"password": string(hash)})
}
