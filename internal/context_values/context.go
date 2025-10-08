package contextvalues

import (
	"context"
	"net/http"

	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

const (
	UserIDCtxKey         = "c-user-id"
	UserCtxKey           = "c-user"
	AuthSessionIDCtxKey  = "c-auth-session-id"
	ResponseWriterAccess = "c-response-writer-access"
)

func GetUserID(ctx context.Context) uuid.UUID {
	value, ok := ctx.Value(UserIDCtxKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return value
}

func GetUserIDOrPanic(ctx context.Context) uuid.UUID {
	id := GetUserID(ctx)
	if id == uuid.Nil {
		panic("user id shouldn't be nil")
	}
	return id
}

func GetUser(ctx context.Context) entities.User {
	value, ok := ctx.Value(UserCtxKey).(entities.User)
	if !ok {
		return entities.User{}
	}
	return value
}

func GetUserOrPanic(ctx context.Context) entities.User {
	value, ok := ctx.Value(UserCtxKey).(entities.User)
	if !ok {
		panic("user value expected in ctx")
	}
	return value

}

func GetSessionID(ctx context.Context) string {
	value, ok := ctx.Value(AuthSessionIDCtxKey).(string)
	if !ok {
		return ""
	}
	return value
}
func GetSessionIDOrPanic(ctx context.Context) string {
	session := GetSessionID(ctx)
	if session == "" {
		panic("session shouldn't be empty")
	}
	return session
}

func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	w, ok := ctx.Value(ResponseWriterAccess).(http.ResponseWriter)
	if !ok {
		return nil
	}
	return w
}

func GetResponseWriterOrPanic(ctx context.Context) http.ResponseWriter {
	w := GetResponseWriter(ctx)
	if w == nil {
		panic("response writer is null")
	}
	return w
}

func GetRequestID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}
