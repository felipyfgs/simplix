package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type CustomAttributeRepo struct{ db *pgxpool.Pool }

func NewCustomAttributeRepo(db *pgxpool.Pool) *CustomAttributeRepo {
	return &CustomAttributeRepo{db: db}
}

func scanAttr(row interface{ Scan(...any) error }) (*domain.CustomAttributeDefinition, error) {
	a := &domain.CustomAttributeDefinition{}
	var optionsRaw []byte
	err := row.Scan(&a.ID, &a.EntityType, &a.AttributeKey, &a.DisplayName, &a.AttributeType, &optionsRaw, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(optionsRaw, &a.Options)
	if a.Options == nil {
		a.Options = []string{}
	}
	return a, nil
}

func (r *CustomAttributeRepo) List(ctx context.Context, entityType string) ([]domain.CustomAttributeDefinition, error) {
	q := `SELECT id, entity_type, attribute_key, display_name, attribute_type, options, created_at
	      FROM custom_attribute_definitions`
	args := []any{}
	if entityType != "" {
		q += " WHERE entity_type=$1"
		args = append(args, entityType)
	}
	q += " ORDER BY created_at ASC"
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var attrs []domain.CustomAttributeDefinition
	for rows.Next() {
		a, err := scanAttr(rows)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, *a)
	}
	return attrs, nil
}

func (r *CustomAttributeRepo) Create(ctx context.Context, entityType, key, displayName, attrType string, options []string) (*domain.CustomAttributeDefinition, error) {
	optJSON, _ := json.Marshal(options)
	row := r.db.QueryRow(ctx, `
		INSERT INTO custom_attribute_definitions (entity_type, attribute_key, display_name, attribute_type, options)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, entity_type, attribute_key, display_name, attribute_type, options, created_at`,
		entityType, key, displayName, attrType, optJSON,
	)
	return scanAttr(row)
}

func (r *CustomAttributeRepo) Update(ctx context.Context, id uuid.UUID, displayName string, options []string) (*domain.CustomAttributeDefinition, error) {
	optJSON, _ := json.Marshal(options)
	row := r.db.QueryRow(ctx, `
		UPDATE custom_attribute_definitions SET display_name=$2, options=$3
		WHERE id=$1
		RETURNING id, entity_type, attribute_key, display_name, attribute_type, options, created_at`,
		id, displayName, optJSON,
	)
	return scanAttr(row)
}

func (r *CustomAttributeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM custom_attribute_definitions WHERE id=$1`, id)
	return err
}
