package gql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/plinkplenk/booki/internal/api/gql/graph"
	"github.com/plinkplenk/booki/internal/api/gql/resolvers"
	"github.com/plinkplenk/booki/internal/api/middlewares"
	"github.com/plinkplenk/booki/internal/auth"
	"github.com/plinkplenk/booki/internal/user"
	"log/slog"
	"net/http"
)

type GQL http.Handler

type Args struct {
	UserService    user.Service
	AuthService    auth.Service
	Logger         *slog.Logger
	AuthMiddleware middlewares.Auth
}

func New(args Args) GQL {
	cfg := graph.Config{
		Resolvers: &resolvers.Resolver{
			UserService: args.UserService,
			AuthService: args.AuthService,
			Logger:      args.Logger,
		},
	}
	cfg.Directives.IsAuthenticated = args.AuthMiddleware.GQLDirective
	srv := handler.New(graph.NewExecutableSchema(cfg))
	srv.Use(extension.Introspection{}) // TODO: remove in production
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.SSE{})
	return GQL(srv)
}
