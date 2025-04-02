package user

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Create(ctx context.Context, user User) (User, error)
	Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (User, error)
	Deactivate(ctx context.Context, id uuid.UUID) (User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

var (
	ErrUserNotFound = errors.New("user not found")
)

type ErrInvalidField struct {
	Field string
}

func (e ErrInvalidField) Error() string {
	return "invalid field: " + e.Field
}

type postgresRepository struct {
	conn *pgxpool.Pool
}

func NewPostgresRepository(conn *pgxpool.Pool) Repository {
	return postgresRepository{conn}
}

func (r postgresRepository) scanUserRow(row pgx.Row) (User, error) {
	user := User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.IsActive, &user.CreatedAt)
	return user, err
}

func (r postgresRepository) getByField(ctx context.Context, field string, value any) (User, error) {
	query := fmt.Sprintf(`SELECT id, email, password, is_active, created_at FROM users WHERE %s = $1`, field)
	user, err := r.scanUserRow(r.conn.QueryRow(ctx, query, value))
	if err != nil {
		fmt.Printf("err := %e\n", err)
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Printf("here\n")
			return User{}, ErrUserNotFound
		} else {
			return User{}, err
		}
	}
	return user, nil
}

func (r postgresRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	return r.getByField(ctx, "email", email)
}

func (r postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	return r.getByField(ctx, "id", id)
}

func (r postgresRepository) Create(ctx context.Context, user User) (User, error) {
	query := `
INSERT INTO users (id, email, password, is_active) 
VALUES ($1, $2, $3, $4)
RETURNING id, email, password, is_active, created_at`
	row := r.conn.QueryRow(ctx, query, user.ID, user.Email, user.Password, user.IsActive)
	return r.scanUserRow(row)
}

func (r postgresRepository) Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (User, error) {
	if len(fieldValue) == 0 {
		return r.GetByID(ctx, userID)
	}
	builder := sq.Update("users").
		SetMap(fieldValue).
		Where(sq.Eq{"id": userID}).
		Suffix("RETURNING id, email, password, is_active, created_at").
		PlaceholderFormat(sq.Dollar)
	query, args, err := builder.ToSql()
	if err != nil {
		return User{}, err
	}
	row := r.conn.QueryRow(ctx, query, args...)
	return r.scanUserRow(row)
}

func (r postgresRepository) Deactivate(ctx context.Context, id uuid.UUID) (User, error) {
	query := `
UPDATE users 
SET is_active = false 
WHERE id = $1 
RETURNING id, email, password, is_active, created_at`
	row := r.conn.QueryRow(ctx, query, id)
	return r.scanUserRow(row)
}

func (r postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
DELETE FROM users
WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	return err
}
