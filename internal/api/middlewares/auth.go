package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/Shelffy/shelffy/internal/api"
	"github.com/Shelffy/shelffy/internal/context_values"
	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/Shelffy/shelffy/internal/services"
)

const SessionLength = 128

type Auth struct {
	userService services.Users
	authService services.Auth
	logger      *slog.Logger
}

func NewAuthMiddleware(userService services.Users, authService services.Auth, logger *slog.Logger) Auth {
	return Auth{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

func (a Auth) getSession(r *http.Request) (entities.Session, error) {
	sessionCookie, err := r.Cookie(api.SessionIDCookieName)
	if err != nil {
		return entities.NilSession, err
	}
	session, err := a.authService.ValidateSession(r.Context(), sessionCookie.Value)
	if err != nil {
		return entities.NilSession, err
	}
	return session, nil
}

func (a Auth) setSessionToCtx(ctx context.Context, sessionID string) context.Context {
	ctx = context.WithValue(ctx, contextvalues.AuthSessionIDCtxKey, sessionID)
	return ctx
}

func (a Auth) setUserToCtx(ctx context.Context, user entities.User) context.Context {
	return context.WithValue(ctx, contextvalues.UserCtxKey, user)
}

func (a Auth) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			session, err := a.getSession(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"}); err != nil {
					a.logger.Error("failed to write response", "error", err.Error())
					return
				}
			}
			user, err := a.userService.GetByID(r.Context(), session.UserID)
			if err != nil {
				a.logger.Error("error while trying to get user in auth middleware", "error", err.Error())
				if err := json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"}); err != nil {
					a.logger.Error("failed to write response", "error", err.Error())
					return
				}
			}
			r = r.WithContext(a.setSessionToCtx(r.Context(), session.ID))
			r = r.WithContext(a.setUserToCtx(r.Context(), user))
			next.ServeHTTP(w, r)
		},
	)
}

func (a Auth) GQLHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			session, err := a.getSession(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			user, err := a.userService.GetByID(r.Context(), session.UserID)
			if err != nil {
				a.logger.Error("error while trying to get user in auth middleware", "error", err.Error())
				if err := json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"}); err != nil {
					a.logger.Error("failed to write response", "error", err.Error())
					return
				}
			}
			r = r.WithContext(a.setSessionToCtx(r.Context(), session.ID))
			r = r.WithContext(a.setUserToCtx(r.Context(), user))
			next.ServeHTTP(w, r)
		},
	)
}

func (a Auth) GQLDirective(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
	session := contextvalues.GetSessionIDOrPanic(ctx)
	if session == "" || len(session) < SessionLength {
		return nil, errors.New("invalid session value")
	}
	return next(ctx)
}
