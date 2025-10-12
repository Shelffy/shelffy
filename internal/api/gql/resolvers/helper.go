package resolvers

import (
	"context"
	"net/url"

	"github.com/Shelffy/shelffy/internal/config"
	contextvalues "github.com/Shelffy/shelffy/internal/context_values"
	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/google/uuid"
)

func IsBookOwnerOrAdmin(ctx context.Context, book entities.Book) bool {
	user := contextvalues.GetUserOrPanic(ctx)
	return user.IsAdmin || user.ID == book.UploadedBy
}

func BuildBookContentURL(baseURL string, endpoints config.Endpoints, bookID uuid.UUID) (string, error) {
	bookURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	api := endpoints.API
	bookURL.Path, err = url.JoinPath(bookURL.Path, api.Base, api.V1.Base, api.V1.Books, bookID.String())
	if err != nil {
		return "", err
	}
	return bookURL.String(), nil
}
