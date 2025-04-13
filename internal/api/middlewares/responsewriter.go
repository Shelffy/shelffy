package middlewares

import (
	"context"
	"github.com/plinkplenk/booki/internal/api/apictx"
	"net/http"
)

func ResponseWriterAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), apictx.ResponseWriterAccess, w))
		next.ServeHTTP(w, r)
	})
}
