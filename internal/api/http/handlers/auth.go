package handlers

import (
	"errors"
	"github.com/google/uuid"
	"github.com/plinkplenk/booki/internal/api"
	"github.com/plinkplenk/booki/internal/api/http/schema"
	"github.com/plinkplenk/booki/internal/auth"
	"github.com/plinkplenk/booki/internal/user"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	authService auth.Service
	userService user.Service
	logger      *slog.Logger
}

func NewAuthHandler(authService auth.Service, userService user.Service, logger *slog.Logger) AuthHandler {
	return AuthHandler{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	registerData, err := getRequestData[schema.Register](r)
	if err != nil {
		h.logger.Error("failed to get request data body data", "error", err)
		err = errorResponse("invalid data provided", http.StatusBadRequest, w)
		logResponseWriteError(err, h.logger)
		return
	}
	_, err = h.userService.GetByEmail(r.Context(), registerData.Email)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		h.logger.Error("failed to get user", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	} else if err == nil {
		err = errorResponse("user with this email already exists", http.StatusConflict, w)
		logResponseWriteError(err, h.logger)
		return
	}
	dbUser, err := h.userService.Create(r.Context(), user.User{
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
	err = successResponse(dbUser, http.StatusCreated, w)
	logResponseWriteError(err, h.logger)
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginData, err := getRequestData[schema.Login](r)
	if err != nil {
		h.logger.Error("failed to get request data body data", "error", err)
		err = errorResponse("invalid data provided", http.StatusBadRequest, w)
		logResponseWriteError(err, h.logger)
		return
	}
	dbUser, session, err := h.authService.Login(r.Context(), loginData.Email, loginData.Password)
	if err != nil {
		h.logger.Error("failed to create session", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	cookie := http.Cookie{
		Name:     api.SessionIDCookieName,
		Value:    session.ID,
		Path:     "/",
		Domain:   r.Host,
		Expires:  session.ExpiresAt,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	err = successResponse(R{"user": dbUser}, http.StatusOK, w)
	logResponseWriteError(err, h.logger)
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie(api.SessionIDCookieName)
	if err != nil || sessionCookie.Value == "" {
		err = errorResponse("unauthorized", http.StatusUnauthorized, w)
		logResponseWriteError(err, h.logger)
		return
	}
	if err := h.authService.Deactivate(r.Context(), sessionCookie.Value); err != nil {
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
	err = successResponse(nil, http.StatusOK, w)
	logResponseWriteError(err, h.logger)
}
