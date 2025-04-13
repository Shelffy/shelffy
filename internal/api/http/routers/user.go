package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/plinkplenk/booki/internal/api/http/handlers"
	"net/http"
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
