package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type WebhookRepo struct{ db *pgxpool.Pool }

func NewWebhookRepo(db *pgxpool.Pool) *WebhookRepo { return &WebhookRepo{db: db} }

func (r *WebhookRepo) List(ctx context.Context) ([]domain.Webhook, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, url, subscriptions, secret, enabled, created_at FROM webhooks ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var whs []domain.Webhook
	for rows.Next() {
		w := domain.Webhook{}
		if err := rows.Scan(&w.ID, &w.URL, &w.Subscriptions, &w.Secret, &w.Enabled, &w.CreatedAt); err != nil {
			return nil, err
		}
		whs = append(whs, w)
	}
	return whs, nil
}

func (r *WebhookRepo) Create(ctx context.Context, url string, subs []string, secret *string) (*domain.Webhook, error) {
	w := &domain.Webhook{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO webhooks (url, subscriptions, secret) VALUES ($1, $2, $3)
		 RETURNING id, url, subscriptions, secret, enabled, created_at`,
		url, subs, secret,
	).Scan(&w.ID, &w.URL, &w.Subscriptions, &w.Secret, &w.Enabled, &w.CreatedAt)
	return w, err
}

func (r *WebhookRepo) Update(ctx context.Context, id uuid.UUID, url string, subs []string, enabled bool) (*domain.Webhook, error) {
	w := &domain.Webhook{}
	err := r.db.QueryRow(ctx,
		`UPDATE webhooks SET url=$2, subscriptions=$3, enabled=$4 WHERE id=$1
		 RETURNING id, url, subscriptions, secret, enabled, created_at`,
		id, url, subs, enabled,
	).Scan(&w.ID, &w.URL, &w.Subscriptions, &w.Secret, &w.Enabled, &w.CreatedAt)
	return w, err
}

func (r *WebhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM webhooks WHERE id=$1`, id)
	return err
}

func (r *WebhookRepo) ListEnabled(ctx context.Context) ([]domain.Webhook, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, url, subscriptions, secret, enabled, created_at FROM webhooks WHERE enabled=TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var whs []domain.Webhook
	for rows.Next() {
		w := domain.Webhook{}
		if err := rows.Scan(&w.ID, &w.URL, &w.Subscriptions, &w.Secret, &w.Enabled, &w.CreatedAt); err != nil {
			return nil, err
		}
		whs = append(whs, w)
	}
	return whs, nil
}
