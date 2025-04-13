package handlers

import (
	"github.com/Shelffy/shelffy/internal/api/apictx"
	"github.com/Shelffy/shelffy/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type UserHandler struct {
	service user.Service
	logger  *slog.Logger
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func toUserResponse(user user.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.String(),
	}
}

func NewUserHandler(service user.Service, logger *slog.Logger) UserHandler {
	return UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "id")
	userID, err := uuid.Parse(strID)
	if err != nil {
		h.logger.Error("failed to parse user id", "error", err)
		err = errorResponse("invalid user id", http.StatusBadRequest, w)
		logResponseWriteError(err, h.logger)
		return
	}
	dbUser, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	err = response(R{"user": toUserResponse(dbUser)}, http.StatusOK, w)
	logResponseWriteError(err, h.logger)
}

func (h UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	id, err := apictx.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("failed to get user id from context", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	dbUser, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", "error", err)
		err = errorResponse("something went wrong", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	err = response(R{"user": toUserResponse(dbUser)}, http.StatusOK, w)
	logResponseWriteError(err, h.logger)
}
