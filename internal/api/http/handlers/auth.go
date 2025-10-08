package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Shelffy/shelffy/internal/api"
	"github.com/Shelffy/shelffy/internal/context_values"
	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/Shelffy/shelffy/internal/repositories"
	"github.com/Shelffy/shelffy/internal/services"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService services.Auth
	userService services.Users
	logger      *slog.Logger
}

func NewAuthHandler(authService services.Auth, userService services.Users, logger *slog.Logger) AuthHandler {
	return AuthHandler{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	registerData, err := getRequestData[RegisterRequest](r)
	if err != nil {
		h.logger.Error("failed to get request data body data", "error", err)
		err = errorResponse("invalid data provided", http.StatusBadRequest, w)
		logResponseWriteError(err, h.logger)
		return
	}
	_, err = h.userService.GetByEmail(r.Context(), registerData.Email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		h.logger.Error("failed to get user", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	} else if err == nil {
		err = errorResponse("user with this email already exists", http.StatusBadRequest, w)
		logResponseWriteError(err, h.logger)
		return
	}
	_, err = h.userService.Create(r.Context(), entities.User{
		ID:       uuid.New(),
		Email:    registerData.Email,
		Password: registerData.Password,
		IsActive: true,
	})
	if err != nil {
		h.logger.Error("failed to create user", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	err = successResponse("user created", http.StatusCreated, w)
	logResponseWriteError(err, h.logger)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginData, err := getRequestData[LoginRequest](r)
	if err != nil {
		h.logger.Error("failed to get request data body data", "error", err)
		err = errorResponse("invalid data provided", http.StatusBadRequest, w)
		logResponseWriteError(err, h.logger)
		return
	}
	dbUser, session, err := h.authService.Login(r.Context(), loginData.Email, loginData.Password)
	if err != nil && !errors.Is(err, services.ErrInvalidCredentials) {
		h.logger.Error("failed to create session", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	} else if errors.Is(err, services.ErrInvalidCredentials) {
		err = errorResponse("invalid credentials", http.StatusUnauthorized, w)
		logResponseWriteError(err, h.logger)
		return
	}
	cookie := http.Cookie{
		Name:     api.SessionIDCookieName,
		Value:    session.ID,
		Path:     "/",
		Domain:   r.Host,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	err = response(R{"user": toUserResponse(dbUser)}, http.StatusOK, w)
	logResponseWriteError(err, h.logger)
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := contextvalues.GetSessionIDOrPanic(ctx)
	if err := h.authService.Deactivate(ctx, sessionID); err != nil {
		h.logger.Error("failed to deactivate session", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	newCookie := http.Cookie{
		Name:        api.SessionIDCookieName,
		Value:       "",
		Path:        "/",
		Domain:      r.Host,
		MaxAge:      -1,
		Secure:      false,
		HttpOnly:    false,
		SameSite:    0,
		Partitioned: false,
	}
	http.SetCookie(w, &newCookie)
	err := response(nil, http.StatusOK, w)
	logResponseWriteError(err, h.logger)
}
