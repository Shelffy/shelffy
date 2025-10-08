package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/Shelffy/shelffy/internal/query-builder/postgres/public/model"
	"github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBookNotFound = fmt.Errorf("user book not found")
)

func entityBookToModel(book entities.Book) model.Books {
	return model.Books{
		ID:         book.ID,
		Title:      book.Title,
		UploadedBy: book.UploadedBy,
		UploadedAt: &book.UploadedAt,
		Hash:       book.Hash[:],
		Path:       book.StoragePath,
	}
}

func bookModelToEntity(book model.Books) entities.Book {
	var hash entities.BookHash
	for i, b := range book.Hash {
		hash[i] = b
	}
	return entities.Book{
		ID:          book.ID,
		Title:       book.Title,
		StoragePath: book.Path,
		Hash:        hash,
		UploadedBy:  book.UploadedBy,
		UploadedAt:  *book.UploadedAt,
	}
}

func scanIntoBookModel(row scannable) (book model.Books, err error) {
	err = row.Scan(&book.ID, &book.Hash, &book.UploadedBy, &book.UploadedAt, &book.Path, &book.Title)
	return
}

type Books interface {
	Create(ctx context.Context, book entities.Book) (entities.Book, error)

	Delete(ctx context.Context, bookID uuid.UUID) error
	GetByID(ctx context.Context, bookID uuid.UUID) (entities.Book, error)
	GetByTitleAndUserID(ctx context.Context, title string, userID uuid.UUID) (entities.Book, error)
	GetByHash(ctx context.Context, hash entities.BookHash) ([]entities.Book, error)
}

type postgresBooksRepository struct {
	pool   *pgxpool.Pool
	getter *pgxv5.CtxGetter
}

func NewBooksPSQLRepository(pool *pgxpool.Pool) Books {
	return postgresBooksRepository{
		pool:   pool,
		getter: pgxv5.DefaultCtxGetter,
	}
}

func (r postgresBooksRepository) Create(ctx context.Context, bookToCreate entities.Book) (entities.Book, error) {
	return entities.Book{}, errors.New("")
	//	book := entityBookToModel(bookToCreate)
	//	sql := `INSERT INTO books(id, hash, uploaded_by, path, title)
	//VALUES ($1, $2, $3, $4, $5)
	//RETURNING id, hash, uploaded_by, uploaded_at, path, title`
	//
	//	b, err := scanIntoBookModel(r.pool.QueryRow(ctx, sql, book.ID, book.Hash, book.UploadedBy, book.Path, book.Title))
	//	if err != nil {
	//		return entities.Book{}, err
	//	}
	//	return bookModelToEntity(b), nil
}

func (r postgresBooksRepository) Delete(ctx context.Context, bookID uuid.UUID) error {
	sql := `DELETE FROM books WHERE id = $1`
	conn := r.getter.DefaultTrOrDB(ctx, r.pool)
	_, err := conn.Exec(ctx, sql, bookID)
	return err
}

func (r postgresBooksRepository) GetByID(ctx context.Context, bookID uuid.UUID) (entities.Book, error) {
	sql := `
SELECT id, hash, uploaded_by, uploaded_at, path, title
FROM books
WHERE id = $1`
	book, err := scanIntoBookModel(r.pool.QueryRow(ctx, sql, bookID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Book{}, ErrBookNotFound
		}
		return entities.Book{}, err
	}
	return bookModelToEntity(book), nil

}

func (r postgresBooksRepository) GetByTitleAndUserID(ctx context.Context, title string, userID uuid.UUID) (entities.Book, error) {
	sql := `
SELECT id, hash, uploaded_by, uploaded_at, path, title
FROM books
WHERE title = $1 AND uploaded_by = $2`
	book, err := scanIntoBookModel(r.pool.QueryRow(ctx, sql, title, userID))
	if err != nil {
		return entities.Book{}, err
	}
	return bookModelToEntity(book), nil
}

func (r postgresBooksRepository) GetByHash(ctx context.Context, hash entities.BookHash) ([]entities.Book, error) {
	sql := `
SELECT id, hash, uploaded_by, uploaded_at, path, title
FROM books
WHERE hash = $1`
	rows, err := r.pool.Query(ctx, sql, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	books := make([]entities.Book, 0, 1)
	for rows.Next() {
		book, err := scanIntoBookModel(rows)
		if err != nil {
			return nil, err
		}
		books = append(books, bookModelToEntity(book))
	}
	return books, nil
}
