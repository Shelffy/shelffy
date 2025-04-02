package auth

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, session Session) (Session, error)
	GetByID(ctx context.Context, id string) (Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Session, error)
	Deactivate(ctx context.Context, id string) error
	DeactivateByUserID(ctx context.Context, userID uuid.UUID, exceptSessionIDs ...string) error
	DeleteSession(ctx context.Context, id string) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, except ...string) error
}

type sessionRepository struct {
	conn *pgxpool.Pool
}

func NewRepository(conn *pgxpool.Pool) Repository {
	return sessionRepository{
		conn: conn,
	}
}

func (r sessionRepository) scanSessionRow(row pgx.Row) (Session, error) {
	session := Session{}
	err := row.Scan(&session.ID, &session.UserID, &session.IsActive, &session.ExpiresAt)
	return session, err
}

func (r sessionRepository) Create(ctx context.Context, session Session) (Session, error) {
	query := `
INSERT INTO sessions (id, user_id, is_active, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, is_active, expires_at;`
	row := r.conn.QueryRow(ctx, query, session.ID, session.UserID, session.IsActive, session.ExpiresAt)
	return r.scanSessionRow(row)
}

func (r sessionRepository) GetByID(ctx context.Context, id string) (Session, error) {
	query := `
SELECT id, user_id , is_active, expires_at
FROM sessions
WHERE id = $1`
	row := r.conn.QueryRow(ctx, query, id)
	return r.scanSessionRow(row)
}

func (r sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	query := `
SELECT id, user_id, is_active, expires_at
FROM sessions
WHERE user_id = $1`
	rows, err := r.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		session, err := r.scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r sessionRepository) Update(ctx context.Context, id string, fieldValue map[string]any) (Session, error) {
	builder := sq.Update("sessions").
		SetMap(fieldValue).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, user_id, is_active, expires_at").
		PlaceholderFormat(sq.Dollar)
	query, args, err := builder.ToSql()
	if err != nil {
		return Session{}, err
	}
	row := r.conn.QueryRow(ctx, query, args...)
	return r.scanSessionRow(row)
}

func (r sessionRepository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE sessions SET is_active = false WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	return err
}

func (r sessionRepository) DeactivateByUserID(ctx context.Context, userID uuid.UUID, exceptSessionIDs ...string) error {
	builder := `
UPDATE sessions
SET is_active = false
WHERE user_id = $1 AND id NOT IN ($2)`
	_, err := r.conn.Exec(ctx, builder, userID, exceptSessionIDs)
	return err
}

func (r sessionRepository) DeleteSession(ctx context.Context, id string) error {
	query := `
DELETE FROM sessions
WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	return err
}

func (r sessionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, except ...string) error {
	query := `
DELETE FROM sessions
WHERE user_id = $1 AND id NOT IN ($2)`
	_, err := r.conn.Exec(ctx, query, userID, except)
	return err
}
