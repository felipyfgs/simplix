package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type LabelRepo struct{ db *pgxpool.Pool }

func NewLabelRepo(db *pgxpool.Pool) *LabelRepo { return &LabelRepo{db: db} }

func (r *LabelRepo) List(ctx context.Context) ([]domain.Label, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, color, description, created_at FROM labels ORDER BY name`)
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

func (r *LabelRepo) Create(ctx context.Context, name, color string, description *string) (*domain.Label, error) {
	l := &domain.Label{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO labels (name, color, description) VALUES ($1,$2,$3) RETURNING id, name, color, description, created_at`,
		name, color, description,
	).Scan(&l.ID, &l.Name, &l.Color, &l.Description, &l.CreatedAt)
	return l, err
}

func (r *LabelRepo) Update(ctx context.Context, id uuid.UUID, name, color string, description *string) (*domain.Label, error) {
	l := &domain.Label{}
	err := r.db.QueryRow(ctx,
		`UPDATE labels SET name=$2, color=$3, description=$4 WHERE id=$1 RETURNING id, name, color, description, created_at`,
		id, name, color, description,
	).Scan(&l.ID, &l.Name, &l.Color, &l.Description, &l.CreatedAt)
	return l, err
}

func (r *LabelRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM labels WHERE id=$1`, id)
	return err
}
