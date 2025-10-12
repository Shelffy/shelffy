package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	contextvalues "github.com/Shelffy/shelffy/internal/context_values"
	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/Shelffy/shelffy/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BooksHandler struct {
	books   services.Books
	storage services.FileStorage
	logger  *slog.Logger
}

func NewBooksHandler(booksService services.Books, storage services.FileStorage, logger *slog.Logger) BooksHandler {
	return BooksHandler{
		books:   booksService,
		storage: storage,
		logger:  logger,
	}
}

func (h BooksHandler) IsOwnerOrAdmin(ctx context.Context, book entities.Book) bool {
	user := contextvalues.GetUserOrPanic(ctx)
	return user.IsAdmin || book.UploadedBy == user.ID
}

func (h BooksHandler) GetContentByID(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "id")
	bookID, err := uuid.Parse(strID)
	if err != nil {
		h.logger.Info("failed to parse book id", "error", err, "id", strID)
		err = errorResponse("invalid book id", http.StatusBadRequest, w)
		logResponseWriteError(err, h.logger)
		return
	}
	book, err := h.books.GetByID(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, services.ErrBookNotFound) {
			err = errorResponse(err.Error(), http.StatusNotFound, w)
			logResponseWriteError(err, h.logger)
			return
		}
		h.logger.Error("failed to parse user id", "error", err)
		err = errorResponse("internal error", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	if !h.IsOwnerOrAdmin(r.Context(), book) {
		err = errorResponse("access denied", http.StatusForbidden, w)
		logResponseWriteError(err, h.logger)
		return
	}
	contentStream, err := h.storage.Get(r.Context(), book.StoragePath)
	if err != nil {
		h.logger.Error("failed to get book from storage", "error", err, "path", book.StoragePath)
		if errors.Is(err, services.ErrObjectNotFound) {
			err = errorResponse(err.Error()+"(book content)", http.StatusNotFound, w)
			logResponseWriteError(err, h.logger)
			return
		}
		err = errorResponse(err.Error(), http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+book.Title)
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err = io.Copy(w, contentStream); err != nil {
		h.logger.Error("failed to write book content to the http writer", "error", err)
		err = errorResponse("internal error", http.StatusInternalServerError, w)
		logResponseWriteError(err, h.logger)
		return
	}
	w.WriteHeader(http.StatusOK)
}
