package repositories

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Users interface {
	GetByEmail(ctx context.Context, email string) (entities.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (entities.User, error)
	Create(ctx context.Context, user entities.User) (entities.User, error)
	Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (entities.User, error)
	Deactivate(ctx context.Context, id uuid.UUID) (entities.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

var (
	ErrUserNotFound = errors.New("user not found")
)

type postgresUsersRepository struct {
	conn *pgxpool.Pool
}

func NewUsersPSQLRepository(conn *pgxpool.Pool) Users {
	return postgresUsersRepository{conn}
}

func (r postgresUsersRepository) scanUserRow(row pgx.Row) (entities.User, error) {
	user := entities.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.IsActive, &user.CreatedAt)
	return user, err
}

func (r postgresUsersRepository) getByField(ctx context.Context, field string, value any) (entities.User, error) {
	query := fmt.Sprintf(`SELECT id, email, password, is_active, created_at FROM users WHERE %s = $1`, field)
	user, err := r.scanUserRow(r.conn.QueryRow(ctx, query, value))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, ErrUserNotFound
		} else {
			return entities.User{}, err
		}
	}
	return user, nil
}

func (r postgresUsersRepository) GetByEmail(ctx context.Context, email string) (entities.User, error) {
	return r.getByField(ctx, "email", email)
}

func (r postgresUsersRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	return r.getByField(ctx, "id", id)
}

func (r postgresUsersRepository) Create(ctx context.Context, user entities.User) (entities.User, error) {
	query := `
INSERT INTO users (id, email, password, is_active, username) 
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, password, is_active, created_at`
	row := r.conn.QueryRow(ctx, query, user.ID, user.Email, user.Password, user.IsActive, user.Username)
	return r.scanUserRow(row)
}

func (r postgresUsersRepository) Update(ctx context.Context, userID uuid.UUID, fieldValue map[string]any) (entities.User, error) {
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
		return entities.User{}, err
	}
	row := r.conn.QueryRow(ctx, query, args...)
	return r.scanUserRow(row)
}

func (r postgresUsersRepository) Deactivate(ctx context.Context, id uuid.UUID) (entities.User, error) {
	query := `
UPDATE users 
SET is_active = false 
WHERE id = $1 
RETURNING id, email, password, is_active, created_at`
	row := r.conn.QueryRow(ctx, query, id)
	return r.scanUserRow(row)
}

func (r postgresUsersRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
DELETE FROM users
WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	return err
}
