package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/plinkplenk/booki/internal/api/http/handlers"
	"net/http"
)

type AuthRouterArgs struct {
	Handler        handlers.AuthHandler
	AuthMiddleware func(http.Handler) http.Handler
}

func NewAuthRouter(args AuthRouterArgs) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/register", args.Handler.Register)
	router.Post("/login", args.Handler.Login)
	router.Group(func(r chi.Router) {
		r.Use(args.AuthMiddleware)
		r.Post("/logout", args.Handler.Logout)
	})

	return router
}
