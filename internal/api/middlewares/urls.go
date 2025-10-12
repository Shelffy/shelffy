package middlewares

import (
	"context"
	"fmt"
	"net/http"

	contextvalues "github.com/Shelffy/shelffy/internal/context_values"
)

func BaseURLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)
		ctx := context.WithValue(r.Context(), contextvalues.BaseURLCtxKey, baseURL)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
