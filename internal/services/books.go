package services

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/Shelffy/shelffy/internal/repositories"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
)

var (
	ErrBookNotFound = errors.New("book not found")
)

type Books interface {
	Upload(ctx context.Context, book entities.Book, contentLength int64, content io.Reader) (entities.Book, error)
	Delete(ctx context.Context, bookID uuid.UUID) error
	GetByID(ctx context.Context, bookID uuid.UUID) (entities.Book, error)
	GetManyByUserID(ctx context.Context, bookID uuid.UUID, limit, offset *uint64) ([]entities.Book, error)
	GetByTitleAndUserID(ctx context.Context, title string, userID uuid.UUID) (entities.Book, error)
	GetBookContentByID(ctx context.Context, bookID uuid.UUID) (io.Reader, error)
}

type booksService struct {
	booksRepository     repositories.Books
	storageService      FileStorage
	timeout             time.Duration
	logger              *slog.Logger
	booksEventPublisher BooksEventsPublisher
	txManager           *manager.Manager
}

func NewBookService(
	booksRepo repositories.Books,
	storage FileStorage,
	timeout time.Duration,
	booksEventPublisher BooksEventsPublisher,
	txManager *manager.Manager,
	logger *slog.Logger,
) Books {
	return booksService{
		booksRepository:     booksRepo,
		storageService:      storage,
		timeout:             timeout,
		logger:              logger,
		booksEventPublisher: booksEventPublisher,
		txManager:           txManager,
	}
}

func (s booksService) createStoragePath(ownerUsername string, title string) string {
	builder := strings.Builder{}
	builder.Grow(len(ownerUsername) + len(title) + 1)
	builder.WriteString(ownerUsername)
	builder.WriteByte('/')
	builder.WriteString(title)
	return builder.String()
}

func (s booksService) Upload(ctx context.Context, book entities.Book, contentLength int64, content io.Reader) (entities.Book, error) {
	l := s.logger.WithGroup("Upload")
	book.StoragePath = s.createStoragePath(book.UploadedBy.String(), uuid.New().String())
	hash := sha256.New()
	reader := io.TeeReader(content, hash)
	if err := s.storageService.Upload(ctx, book.StoragePath, contentLength, reader); err != nil {
		l.Error("could not upload file to storage", "error", err.Error(), "path", book.StoragePath)
		return entities.Book{}, ErrInternal
	}
	book.Hash = entities.BookHash(hash.Sum(nil))
	book.ID = uuid.New()
	// TODO: it is still not proper way to do this
	createdBook, err := s.booksRepository.Create(ctx, book)
	if err != nil {
		l.Error("cannot create book in book's repository", "error", err.Error())
		if err := s.booksEventPublisher.PublishDeleteBookEvent(ctx, book.StoragePath); err != nil {
			l.Error("cannot create publish delete book event", "error", err.Error(), "ack")
		}
		return entities.Book{}, ErrInternal
	}
	return createdBook, nil
}

func (s booksService) Delete(ctx context.Context, bookID uuid.UUID) error {
	l := s.logger.WithGroup("Delete")
	return s.txManager.Do(ctx, func(ctx context.Context) error {
		book, err := s.booksRepository.GetByID(ctx, bookID)
		if err != nil {
			if !errors.Is(err, repositories.ErrBookNotFound) {
				l.Error("cannot get book from book's repository", "error", err.Error(), "book", book)
				return ErrBookNotFound
			}
			return err
		}
		if err := s.booksRepository.Delete(ctx, bookID); err != nil {
			l.Error("cannot delete book from book's repository", "error", err.Error(), "book", book)
			return ErrInternal
		}
		err = s.booksEventPublisher.PublishDeleteBookEvent(ctx, book.StoragePath)
		if err != nil {
			l.Error("could not publish delete book event", "error", err.Error())
			return ErrInternal
		}
		return nil
	})
}

func (s booksService) GetByID(ctx context.Context, bookID uuid.UUID) (entities.Book, error) {
	l := s.logger.WithGroup("GetByID")
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	book, err := s.booksRepository.GetByID(c, bookID)
	if err != nil {
		if !errors.Is(err, repositories.ErrBookNotFound) {
			l.Error("cannot get book from book's repository", "error", err.Error())
			return entities.Book{}, ErrInternal
		}
		return entities.Book{}, ErrBookNotFound
	}
	return book, nil
}

func (s booksService) GetByTitleAndUserID(ctx context.Context, title string, userID uuid.UUID) (entities.Book, error) {
	l := s.logger.WithGroup("GetByTitleAndUserID")
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	book, err := s.booksRepository.GetByTitleAndUserID(c, title, userID)
	if err != nil {
		if !errors.Is(err, repositories.ErrBookNotFound) {
			l.Error("cannot get book from book's repository", "error", err.Error())
			return entities.Book{}, ErrInternal
		}
		return entities.Book{}, ErrBookNotFound
	}
	return book, nil
}
func (s booksService) GetBookContentByID(ctx context.Context, bookID uuid.UUID) (io.Reader, error) {
	l := s.logger.WithGroup("GetBookContentByID")
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	book, err := s.booksRepository.GetByID(c, bookID)
	if err != nil {
		if !errors.Is(err, repositories.ErrBookNotFound) {
			l.Error("cannot get book from book's repository", "error", err.Error())
			return nil, ErrInternal
		}
		return nil, ErrBookNotFound
	}
	content, err := s.storageService.Get(ctx, book.StoragePath)
	if err != nil {
		l.Error("cannot get book from storage service", "error", err.Error(), "path", book.StoragePath)
		return nil, err
	}
	return content, nil
}

func (s booksService) GetManyByUserID(ctx context.Context, userID uuid.UUID, limit, offset *uint64) ([]entities.Book, error) {
	l := s.logger.WithGroup("GetManyByUserID")
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	books, err := s.booksRepository.GetManyByUserID(c, userID, limit, offset)
	if err != nil {
		if !errors.Is(err, repositories.ErrBookNotFound) {
			l.Error("cannot get books from book's repository", "error", err.Error())
			return nil, ErrInternal
		}
		return nil, ErrBookNotFound
	}
	return books, nil
}
