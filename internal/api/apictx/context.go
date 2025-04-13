package apictx

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

const (
	UserIDCtxKey         = "user-id"
	AuthSessionIDCtxKey  = "auth-session-id"
	ResponseWriterAccess = "response-writer-access"
)

var (
	ErrNoUserIDInContext    = errors.New("no user id in context")
	ErrNoSessionIDInContext = errors.New("no session id in context")
	ErrNoResponseWriter     = errors.New("no response writer in context")
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	value, ok := ctx.Value(UserIDCtxKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrNoUserIDInContext
	}
	return value, nil
}

func GetSessionIDFromContext(ctx context.Context) (string, error) {
	value, ok := ctx.Value(AuthSessionIDCtxKey).(string)
	if !ok {
		return "", ErrNoSessionIDInContext
	}
	return value, nil
}

func GetResponseWriter(ctx context.Context) (http.ResponseWriter, error) {
	w, ok := ctx.Value(ResponseWriterAccess).(http.ResponseWriter)
	if !ok {
		return nil, ErrNoResponseWriter
	}
	return w, nil
}
