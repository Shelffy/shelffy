package contextvalues

import (
	"context"
	"net/http"

	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	UserCtxKey           = "c-user"
	AuthSessionIDCtxKey  = "c-auth-session-id"
	ResponseWriterAccess = "c-response-writer-access"
	BaseURLCtxKey        = "c-base-url"
	IsAdminCtxKey        = "c-is-admin"
)

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
		panic("session value expected in context")
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

func GetBaseURL(ctx context.Context) string {
	url := ctx.Value(BaseURLCtxKey)
	if url, ok := url.(string); ok {
		return url
	} else {
		panic("must be unreachable")
	}
}

func GetRequestID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}

func GetIsAdmin(ctx context.Context) bool {
	return ctx.Value(IsAdminCtxKey).(bool)
}
