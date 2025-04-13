package resolvers

import (
	"github.com/Shelffy/shelffy/internal/auth"
	"github.com/Shelffy/shelffy/internal/user"
	"log/slog"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UserService user.Service
	AuthService auth.Service
	Logger      *slog.Logger
}
