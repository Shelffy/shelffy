package routers

import (
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Shelffy/shelffy/internal/api/http/handlers"
	"github.com/Shelffy/shelffy/internal/api/middlewares"
	"github.com/Shelffy/shelffy/internal/services"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type RouterArgs struct {
	UserService    services.Users
	AuthService    services.Auth
	BooksService   services.Books
	StorageService services.FileStorage
	GQLHandler     http.Handler
	Logger         *slog.Logger
}

func NewRouter(args RouterArgs) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		chimiddleware.Logger,
		chimiddleware.Recoverer,

		middlewares.BaseURLMiddleware,
	)
	authMiddleware := middlewares.NewAuthMiddleware(args.UserService, args.AuthService, args.Logger)
	router.Route("/api", func(r chi.Router) {
		if args.GQLHandler != nil {
			r.Route("/gql", func(r chi.Router) {
				r.Use(authMiddleware.GQLHandler, middlewares.ResponseWriterAccess)
				r.Handle("/", playground.Handler("GraphQL Playground", "/api/gql/query"))
				r.Handle("/query", args.GQLHandler)
			})
		}
		r.Route("/v1", func(r chi.Router) {
			r.Mount(
				"/auth",
				NewAuthRouter(AuthRouterArgs{
					Handler:        handlers.NewAuthHandler(args.AuthService, args.UserService, args.Logger),
					AuthMiddleware: authMiddleware.HTTPHandler,
				}),
			)
			r.Mount(
				"/users",
				NewUserRouter(UserRouterArgs{
					Handler:        handlers.NewUserHandler(args.UserService, args.Logger),
					AuthMiddleware: authMiddleware.HTTPHandler,
				}),
			)
			r.Mount(
				"/books",
				NewBooksRouter(BooksRouterArgs{
					Handler:        handlers.NewBooksHandler(args.BooksService, args.StorageService, args.Logger),
					AuthMiddleware: authMiddleware.HTTPHandler,
				}),
			)
		})
	})
	return router
}
