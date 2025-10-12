package resolvers

import (
	"log/slog"

	"github.com/Shelffy/shelffy/internal/config"
	"github.com/Shelffy/shelffy/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UsersService services.Users
	AuthService  services.Auth
	BooksService services.Books
	Logger       *slog.Logger
	AppEndpoints config.Endpoints
}
