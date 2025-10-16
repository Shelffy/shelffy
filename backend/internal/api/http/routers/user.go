package routers

import (
	"net/http"

	"github.com/Shelffy/shelffy/internal/api/http/handlers"
	"github.com/go-chi/chi/v5"
)

type UserRouterArgs struct {
	Handler        handlers.UserHandler
	AuthMiddleware func(http.Handler) http.Handler
}

func NewUserRouter(args UserRouterArgs) *chi.Mux {
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(args.AuthMiddleware)
		r.Get("/{id}", args.Handler.GetByID)
		r.Get("/me", args.Handler.Me)
	})

	return router
}
