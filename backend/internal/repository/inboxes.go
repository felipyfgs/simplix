package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type InboxRepo struct{ db *pgxpool.Pool }

func NewInboxRepo(db *pgxpool.Pool) *InboxRepo { return &InboxRepo{db: db} }

func (r *InboxRepo) List(ctx context.Context) ([]domain.Inbox, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, channel_type, settings, created_at FROM inboxes ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inboxes []domain.Inbox
	for rows.Next() {
		ix, err := scanInbox(rows)
		if err != nil {
			return nil, err
		}
		inboxes = append(inboxes, *ix)
	}
	return inboxes, nil
}

func (r *InboxRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Inbox, error) {
	return scanInbox(r.db.QueryRow(ctx,
		`SELECT id, name, channel_type, settings, created_at FROM inboxes WHERE id=$1`, id))
}

func (r *InboxRepo) Create(ctx context.Context, name string, channelType domain.ChannelType, settings any) (*domain.Inbox, error) {
	b, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	return scanInbox(r.db.QueryRow(ctx,
		`INSERT INTO inboxes (name, channel_type, settings) VALUES ($1, $2, $3)
		 RETURNING id, name, channel_type, settings, created_at`,
		name, channelType, b))
}

func (r *InboxRepo) Update(ctx context.Context, id uuid.UUID, name string, settings any) (*domain.Inbox, error) {
	b, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	return scanInbox(r.db.QueryRow(ctx,
		`UPDATE inboxes SET name=$2, settings=$3 WHERE id=$1
		 RETURNING id, name, channel_type, settings, created_at`,
		id, name, b))
}

func (r *InboxRepo) UpdateSettings(ctx context.Context, id uuid.UUID, settings any) error {
	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE inboxes SET settings=$2 WHERE id=$1`, id, b)
	return err
}

func (r *InboxRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM inboxes WHERE id=$1`, id)
	return err
}

func scanInbox(row interface{ Scan(...any) error }) (*domain.Inbox, error) {
	ix := &domain.Inbox{}
	var settingsRaw []byte
	if err := row.Scan(&ix.ID, &ix.Name, &ix.ChannelType, &settingsRaw, &ix.CreatedAt); err != nil {
		return nil, err
	}
	if len(settingsRaw) > 0 {
		switch ix.ChannelType {
		case domain.ChannelQuePasa:
			var s domain.QuePasaSettings
			if err := json.Unmarshal(settingsRaw, &s); err == nil {
				ix.Settings = s
			}
		default:
			var s domain.WhatsAppSettings
			if err := json.Unmarshal(settingsRaw, &s); err == nil {
				ix.Settings = s
			}
		}
	}
	return ix, nil
}
