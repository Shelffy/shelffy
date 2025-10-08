package middlewares

import (
	"context"

	"github.com/Shelffy/shelffy/internal/context_values"

	"net/http"
)

func ResponseWriterAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), contextvalues.ResponseWriterAccess, w))
		next.ServeHTTP(w, r)
	})
}
