package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/Shelffy/shelffy/internal/api"
	"github.com/Shelffy/shelffy/internal/api/apictx"
	"github.com/Shelffy/shelffy/internal/auth"
	"github.com/Shelffy/shelffy/internal/user"
	"log/slog"
	"net/http"
)

const SessionLength = 128

type Auth struct {
	userService user.Service
	authService auth.Service
	logger      *slog.Logger
}

func NewAuthMiddleware(userService user.Service, authService auth.Service, logger *slog.Logger) Auth {
	return Auth{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

func (a Auth) getSession(r *http.Request) (auth.Session, error) {
	sessionCookie, err := r.Cookie(api.SessionIDCookieName)
	if err != nil {
		return auth.NilSession, err
	}
	session, err := a.authService.ValidateSession(r.Context(), sessionCookie.Value)
	if err != nil {
		return auth.NilSession, err
	}
	return session, nil
}

func (a Auth) setSessionToCtx(ctx context.Context, session auth.Session) context.Context {
	ctx = context.WithValue(ctx, apictx.UserIDCtxKey, session.UserID)
	ctx = context.WithValue(ctx, apictx.AuthSessionIDCtxKey, session.ID)
	return ctx
}

func (a Auth) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			session, err := a.getSession(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"}); err != nil {
					a.logger.Error("failed to write response", "error", err)
					return
				}
			}
			r = r.WithContext(a.setSessionToCtx(r.Context(), session))
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
			r = r.WithContext(a.setSessionToCtx(r.Context(), session))
			next.ServeHTTP(w, r)
		},
	)
}

func (a Auth) GQLDirective(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
	session, err := apictx.GetSessionIDFromContext(ctx)
	if err != nil {
		return nil, errors.New("unauthorized")
	}
	if session == "" || len(session) < SessionLength {
		return nil, errors.New("invalid session value")
	}
	return next(ctx)
}
