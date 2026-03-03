package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type ConversationRepo struct{ db *pgxpool.Pool }

func NewConversationRepo(db *pgxpool.Pool) *ConversationRepo { return &ConversationRepo{db: db} }

type ConversationFilter struct {
	Status     string
	InboxID    string
	AssignedTo string
	ContactID  string
	Label      string
	Page       int
	Limit      int
}

func (r *ConversationRepo) List(ctx context.Context, f ConversationFilter) ([]domain.Conversation, int, error) {
	args := []any{}
	conds := []string{}
	i := 1

	if f.Status != "" {
		conds = append(conds, fmt.Sprintf("cv.status = $%d", i)); args = append(args, f.Status); i++
	}
	if f.InboxID != "" {
		conds = append(conds, fmt.Sprintf("cv.inbox_id = $%d", i)); args = append(args, f.InboxID); i++
	}
	if f.AssignedTo != "" {
		conds = append(conds, fmt.Sprintf("cv.assigned_to = $%d", i)); args = append(args, f.AssignedTo); i++
	}
	if f.ContactID != "" {
		conds = append(conds, fmt.Sprintf("cv.contact_id = $%d", i)); args = append(args, f.ContactID); i++
	}
	if f.Label != "" {
		conds = append(conds, fmt.Sprintf("EXISTS (SELECT 1 FROM conversation_labels cl JOIN labels l ON l.id=cl.label_id WHERE cl.conversation_id=cv.id AND l.name=$%d)", i))
		args = append(args, f.Label); i++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM conversations cv %s", where), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	q := fmt.Sprintf(`
		SELECT cv.id, cv.contact_id, cv.inbox_id, cv.assigned_to, cv.status, cv.priority,
		       cv.subject, cv.meta, cv.custom_attributes, cv.created_at, cv.updated_at, cv.last_activity_at,
		       c.id, c.name, c.avatar_url,
		       lm.id, lm.sender_type, lm.content, lm.content_type, lm.created_at
		FROM conversations cv
		LEFT JOIN contacts c ON c.id = cv.contact_id
		LEFT JOIN LATERAL (
			SELECT id, sender_type, content, content_type, created_at
			FROM messages
			WHERE conversation_id = cv.id
			ORDER BY created_at DESC LIMIT 1
		) lm ON true
		%s
		ORDER BY cv.last_activity_at DESC LIMIT $%d OFFSET $%d`, where, i, i+1)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var convs []domain.Conversation
	for rows.Next() {
		cv, err := scanConversationRich(rows)
		if err != nil {
			return nil, 0, err
		}
		convs = append(convs, *cv)
	}
	return convs, total, nil
}

func scanConversation(row interface{ Scan(...any) error }) (*domain.Conversation, error) {
	cv := &domain.Conversation{}
	var caRaw []byte
	err := row.Scan(&cv.ID, &cv.ContactID, &cv.InboxID, &cv.AssignedTo, &cv.Status, &cv.Priority,
		&cv.Subject, &cv.Meta, &caRaw, &cv.CreatedAt, &cv.UpdatedAt, &cv.LastActivityAt)
	if err != nil {
		return nil, err
	}
	cv.CustomAttributes = make(map[string]any)
	if len(caRaw) > 0 {
		_ = json.Unmarshal(caRaw, &cv.CustomAttributes)
	}
	return cv, nil
}

func scanConversationRich(row interface{ Scan(...any) error }) (*domain.Conversation, error) {
	cv := &domain.Conversation{}
	var caRaw []byte
	var contactID *string
	var contactName *string
	var contactAvatar *string
	var msgID *string
	var msgSenderType *string
	var msgContent *string
	var msgContentType *string
	var msgCreatedAt *string

	err := row.Scan(
		&cv.ID, &cv.ContactID, &cv.InboxID, &cv.AssignedTo, &cv.Status, &cv.Priority,
		&cv.Subject, &cv.Meta, &caRaw, &cv.CreatedAt, &cv.UpdatedAt, &cv.LastActivityAt,
		&contactID, &contactName, &contactAvatar,
		&msgID, &msgSenderType, &msgContent, &msgContentType, &msgCreatedAt,
	)
	if err != nil {
		return nil, err
	}
	cv.CustomAttributes = make(map[string]any)
	if len(caRaw) > 0 {
		_ = json.Unmarshal(caRaw, &cv.CustomAttributes)
	}
	if contactName != nil {
		cv.Contact = &domain.Contact{
			ID:        cv.ContactID,
			Name:      *contactName,
			AvatarURL: contactAvatar,
		}
	}
	if msgID != nil && msgContent != nil {
		cv.LastMessage = &domain.Message{
			SenderType:  domain.SenderType(*msgSenderType),
			Content:     *msgContent,
			ContentType: domain.ContentType(*msgContentType),
		}
	}
	return cv, nil
}

func (r *ConversationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	cv, err := scanConversation(r.db.QueryRow(ctx, `
		SELECT id, contact_id, inbox_id, assigned_to, status, priority,
		       subject, meta, custom_attributes, created_at, updated_at, last_activity_at
		FROM conversations WHERE id=$1`, id))
	if err != nil {
		return nil, fmt.Errorf("find conversation: %w", err)
	}
	return cv, nil
}

func (r *ConversationRepo) Create(ctx context.Context, contactID, inboxID uuid.UUID, subject *string) (*domain.Conversation, error) {
	return scanConversation(r.db.QueryRow(ctx, `
		INSERT INTO conversations (contact_id, inbox_id, subject)
		VALUES ($1, $2, $3)
		RETURNING id, contact_id, inbox_id, assigned_to, status, priority,
		          subject, meta, custom_attributes, created_at, updated_at, last_activity_at`,
		contactID, inboxID, subject))
}

func (r *ConversationRepo) UpdateCustomAttributes(ctx context.Context, id uuid.UUID, attrs map[string]any) error {
	b, _ := json.Marshal(attrs)
	_, err := r.db.Exec(ctx, `UPDATE conversations SET custom_attributes=$2, updated_at=NOW() WHERE id=$1`, id, b)
	return err
}

func (r *ConversationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ConversationStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE conversations SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

func (r *ConversationRepo) Assign(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE conversations SET assigned_to=$2, updated_at=NOW() WHERE id=$1`, id, userID)
	return err
}

func (r *ConversationRepo) TouchActivity(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE conversations SET last_activity_at=NOW(), updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *ConversationRepo) GetLabels(ctx context.Context, conversationID uuid.UUID) ([]domain.Label, error) {
	rows, err := r.db.Query(ctx,
		`SELECT l.id, l.name, l.color, l.description, l.created_at
		 FROM labels l JOIN conversation_labels cl ON cl.label_id=l.id
		 WHERE cl.conversation_id=$1`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var labels []domain.Label
	for rows.Next() {
		l := domain.Label{}
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &l.Description, &l.CreatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, nil
}

func (r *ConversationRepo) SetLabels(ctx context.Context, conversationID uuid.UUID, labelIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM conversation_labels WHERE conversation_id=$1`, conversationID); err != nil {
		return err
	}
	for _, lid := range labelIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO conversation_labels VALUES ($1,$2) ON CONFLICT DO NOTHING`, conversationID, lid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ConversationRepo) GetFirstInboxID(ctx context.Context) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM inboxes ORDER BY created_at LIMIT 1`).Scan(&id)
	return id, err
}
