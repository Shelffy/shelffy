package repositories

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/Shelffy/shelffy/internal/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session interface {
	Create(ctx context.Context, session entities.Session) (entities.Session, error)
	GetByID(ctx context.Context, id string) (entities.Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Session, error)
	Deactivate(ctx context.Context, id string) error
	DeactivateByUserID(ctx context.Context, userID uuid.UUID, exceptSessionIDs ...string) error
	DeleteSession(ctx context.Context, id string) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, except ...string) error
}

type postgresSessionRepository struct {
	conn *pgxpool.Pool
}

func NewAuthPSQLRepository(conn *pgxpool.Pool) Session {
	return postgresSessionRepository{
		conn: conn,
	}
}

func (r postgresSessionRepository) scanSessionRow(row scannable) (entities.Session, error) {
	session := entities.Session{}
	err := row.Scan(&session.ID, &session.UserID, &session.IsActive, &session.ExpiresAt)
	return session, err
}

func (r postgresSessionRepository) Create(ctx context.Context, session entities.Session) (entities.Session, error) {
	query := `
INSERT INTO sessions (id, user_id, is_active, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, is_active, expires_at;`
	row := r.conn.QueryRow(ctx, query, session.ID, session.UserID, session.IsActive, session.ExpiresAt)
	return r.scanSessionRow(row)
}

func (r postgresSessionRepository) GetByID(ctx context.Context, id string) (entities.Session, error) {
	query := `
SELECT id, user_id , is_active, expires_at
FROM sessions
WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return r.scanSessionRow(row)
}

func (r postgresSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Session, error) {
	query := `
SELECT id, user_id, is_active, expires_at
FROM sessions
WHERE user_id = $1`
	rows, err := r.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []entities.Session
	for rows.Next() {
		session, err := r.scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r postgresSessionRepository) Update(ctx context.Context, id string, fieldValue map[string]any) (entities.Session, error) {
	builder := sq.Update("sessions").
		SetMap(fieldValue).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, user_id, is_active, expires_at").
		PlaceholderFormat(sq.Dollar)
	query, args, err := builder.ToSql()
	if err != nil {
		return entities.Session{}, err
	}
	row := r.conn.QueryRow(ctx, query, args...)
	return r.scanSessionRow(row)
}

func (r postgresSessionRepository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE sessions SET is_active = false WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	return err
}

func (r postgresSessionRepository) DeactivateByUserID(ctx context.Context, userID uuid.UUID, exceptSessionIDs ...string) error {
	builder := `
UPDATE sessions
SET is_active = false
WHERE user_id = $1 AND id NOT IN ($2)`
	_, err := r.conn.Exec(ctx, builder, userID, exceptSessionIDs)
	return err
}

func (r postgresSessionRepository) DeleteSession(ctx context.Context, id string) error {
	query := `
DELETE FROM sessions
WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	return err
}

func (r postgresSessionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, except ...string) error {
	query := `
DELETE FROM sessions
WHERE user_id = $1 AND id NOT IN ($2)`
	_, err := r.conn.Exec(ctx, query, userID, except)
	return err
}
