package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type MessageRepo struct{ db *pgxpool.Pool }

func NewMessageRepo(db *pgxpool.Pool) *MessageRepo { return &MessageRepo{db: db} }

func (r *MessageRepo) List(ctx context.Context, conversationID uuid.UUID) ([]domain.Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, conversation_id, sender_type, sender_id, content, content_type, attachments, source_id, external_status, created_at
		FROM messages WHERE conversation_id=$1 ORDER BY created_at ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []domain.Message
	for rows.Next() {
		m := domain.Message{}
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.SenderID, &m.Content, &m.ContentType, &m.Attachments, &m.SourceID, &m.ExternalStatus, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *MessageRepo) Create(ctx context.Context, conversationID uuid.UUID, senderType domain.SenderType, senderID *uuid.UUID, content string, contentType domain.ContentType) (*domain.Message, error) {
	m := &domain.Message{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_type, sender_id, content, content_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, conversation_id, sender_type, sender_id, content, content_type, attachments, source_id, external_status, created_at`,
		conversationID, senderType, senderID, content, contentType,
	).Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.SenderID, &m.Content, &m.ContentType, &m.Attachments, &m.SourceID, &m.ExternalStatus, &m.CreatedAt)
	return m, err
}

func (r *MessageRepo) CreateWithSourceID(ctx context.Context, conversationID uuid.UUID, senderType domain.SenderType, senderID *uuid.UUID, content string, contentType domain.ContentType, sourceID string) (*domain.Message, error) {
	m := &domain.Message{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_type, sender_id, content, content_type, source_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, conversation_id, sender_type, sender_id, content, content_type, attachments, source_id, external_status, created_at`,
		conversationID, senderType, senderID, content, contentType, sourceID,
	).Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.SenderID, &m.Content, &m.ContentType, &m.Attachments, &m.SourceID, &m.ExternalStatus, &m.CreatedAt)
	return m, err
}

func (r *MessageRepo) UpdateSourceID(ctx context.Context, id uuid.UUID, sourceID string) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET source_id=$2 WHERE id=$1`, id, sourceID)
	return err
}

func (r *MessageRepo) FindBySourceID(ctx context.Context, sourceID string) (*domain.Message, error) {
	m := &domain.Message{}
	err := r.db.QueryRow(ctx, `
		SELECT id, conversation_id, sender_type, sender_id, content, content_type, attachments, source_id, external_status, created_at
		FROM messages WHERE source_id=$1`, sourceID,
	).Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.SenderID, &m.Content, &m.ContentType, &m.Attachments, &m.SourceID, &m.ExternalStatus, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MessageRepo) Delete(ctx context.Context, id, conversationID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM messages WHERE id=$1 AND conversation_id=$2`, id, conversationID)
	return err
}
