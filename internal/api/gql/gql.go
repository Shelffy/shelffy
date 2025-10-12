package gql

import (
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/Shelffy/shelffy/internal/api/gql/graph"
	"github.com/Shelffy/shelffy/internal/api/gql/resolvers"
	"github.com/Shelffy/shelffy/internal/api/middlewares"
	"github.com/Shelffy/shelffy/internal/config"
	"github.com/Shelffy/shelffy/internal/services"
)

type GQL http.Handler

type Args struct {
	UserService    services.Users
	AuthService    services.Auth
	BookService    services.Books
	Logger         *slog.Logger
	AuthMiddleware middlewares.Auth
	Endpoints      config.Endpoints
}

func New(args Args, introspection bool) GQL {
	cfg := graph.Config{
		Resolvers: &resolvers.Resolver{
			UsersService: args.UserService,
			AuthService:  args.AuthService,
			BooksService: args.BookService,
			AppEndpoints: args.Endpoints,
			Logger:       args.Logger,
		},
	}
	cfg.Directives.Auth = args.AuthMiddleware.GQLDirective
	srv := handler.New(graph.NewExecutableSchema(cfg))
	if introspection {
		srv.Use(extension.Introspection{})
	}
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.AddTransport(transport.MultipartMixed{})
	srv.AddTransport(transport.SSE{})
	return GQL(srv)
}
