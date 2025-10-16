package routers

import (
	"net/http"

	"github.com/Shelffy/shelffy/internal/api/http/handlers"
	"github.com/go-chi/chi/v5"
)

type BooksRouterArgs struct {
	Handler        handlers.BooksHandler
	AuthMiddleware func(http.Handler) http.Handler
}

func NewBooksRouter(args BooksRouterArgs) *chi.Mux {
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(args.AuthMiddleware)
		r.Get("/{id}", args.Handler.GetContentByID)
	})

	return router
}
