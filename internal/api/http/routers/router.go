package routers

import (
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Shelffy/shelffy/internal/api/http/handlers"
	"github.com/Shelffy/shelffy/internal/api/middlewares"
	"github.com/Shelffy/shelffy/internal/config"
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

func NewRouter(args RouterArgs, endpoints config.Endpoints) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		chimiddleware.Logger,
		chimiddleware.Recoverer,

		middlewares.BaseURLMiddleware,
	)
	authMiddleware := middlewares.NewAuthMiddleware(args.UserService, args.AuthService, args.Logger)
	router.Route(endpoints.API.Base, func(r chi.Router) {
		api := endpoints.API
		if args.GQLHandler != nil {
			r.Route(api.GQL.Base, func(r chi.Router) {
				r.Use(authMiddleware.GQLHandler, middlewares.ResponseWriterAccess)
				r.Handle("/", playground.Handler("GraphQL Playground", api.Base+api.GQL.Base+api.GQL.Query))
				r.Handle(endpoints.API.GQL.Query, args.GQLHandler)
			})
		}
		r.Route(api.V1.Base, func(r chi.Router) {
			v1 := api.V1
			r.Mount(
				v1.Auth,
				NewAuthRouter(AuthRouterArgs{
					Handler:        handlers.NewAuthHandler(args.AuthService, args.UserService, args.Logger),
					AuthMiddleware: authMiddleware.HTTPHandler,
				}),
			)
			r.Mount(
				v1.Users,
				NewUserRouter(UserRouterArgs{
					Handler:        handlers.NewUserHandler(args.UserService, args.Logger),
					AuthMiddleware: authMiddleware.HTTPHandler,
				}),
			)
			r.Mount(
				v1.Books,
				NewBooksRouter(BooksRouterArgs{
					Handler:        handlers.NewBooksHandler(args.BooksService, args.StorageService, args.Logger),
					AuthMiddleware: authMiddleware.HTTPHandler,
				}),
			)
		})
	})
	return router
}
