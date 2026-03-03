package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type NoteRepo struct{ db *pgxpool.Pool }

func NewNoteRepo(db *pgxpool.Pool) *NoteRepo { return &NoteRepo{db: db} }

func (r *NoteRepo) List(ctx context.Context, contactID uuid.UUID) ([]domain.Note, error) {
	rows, err := r.db.Query(ctx, `
		SELECT n.id, n.contact_id, n.user_id, n.content, n.created_at, n.updated_at,
		       u.id, u.name, u.email, u.role, u.availability, u.avatar_url, u.created_at, u.updated_at
		FROM notes n
		LEFT JOIN users u ON u.id = n.user_id
		WHERE n.contact_id = $1
		ORDER BY n.created_at DESC`, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []domain.Note
	for rows.Next() {
		n := domain.Note{}
		var u domain.User
		var uID *uuid.UUID
		err := rows.Scan(
			&n.ID, &n.ContactID, &n.UserID, &n.Content, &n.CreatedAt, &n.UpdatedAt,
			&uID, &u.Name, &u.Email, &u.Role, &u.Availability, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if uID != nil {
			u.ID = *uID
			n.Author = &u
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *NoteRepo) Create(ctx context.Context, contactID uuid.UUID, userID *uuid.UUID, content string) (*domain.Note, error) {
	n := &domain.Note{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO notes (contact_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, contact_id, user_id, content, created_at, updated_at`,
		contactID, userID, content,
	).Scan(&n.ID, &n.ContactID, &n.UserID, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func (r *NoteRepo) Delete(ctx context.Context, id, contactID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM notes WHERE id=$1 AND contact_id=$2`, id, contactID)
	return err
}
