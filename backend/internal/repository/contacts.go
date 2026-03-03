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

type ContactRepo struct{ db *pgxpool.Pool }

func NewContactRepo(db *pgxpool.Pool) *ContactRepo { return &ContactRepo{db: db} }

const contactCols = `c.id, c.name, c.email, c.phone, c.company, c.avatar_url,
	c.status, c.score, c.assigned_to, c.custom_attributes, c.created_at, c.updated_at, c.company_id`

func scanContact(row interface{ Scan(...any) error }) (*domain.Contact, error) {
	c := &domain.Contact{}
	var caRaw []byte
	err := row.Scan(
		&c.ID, &c.Name, &c.Email, &c.Phone, &c.Company, &c.AvatarURL,
		&c.Status, &c.Score, &c.AssignedTo, &caRaw, &c.CreatedAt, &c.UpdatedAt, &c.CompanyID,
	)
	if err != nil {
		return nil, err
	}
	c.CustomAttributes = make(map[string]any)
	if len(caRaw) > 0 {
		_ = json.Unmarshal(caRaw, &c.CustomAttributes)
	}
	return c, nil
}

type FilterCondition struct {
	Field string // name, email, phone, company, status
	Op    string // contains, equals, starts_with, present, not_present
	Value string
}

type ContactFilter struct {
	Query      string
	Status     string
	Label      string
	Conditions []FilterCondition
	Page       int
	Limit      int
}

func (r *ContactRepo) List(ctx context.Context, f ContactFilter) ([]domain.Contact, int, error) {
	args := []any{}
	conds := []string{}
	i := 1

	if f.Query != "" {
		conds = append(conds, fmt.Sprintf(
			"(c.name ILIKE $%d OR c.email ILIKE $%d OR c.company ILIKE $%d OR c.phone ILIKE $%d)",
			i, i+1, i+2, i+3))
		q := "%" + f.Query + "%"
		args = append(args, q, q, q, q)
		i += 4
	}
	if f.Status != "" {
		conds = append(conds, fmt.Sprintf("c.status = $%d", i))
		args = append(args, f.Status)
		i++
	}
	if f.Label != "" {
		conds = append(conds, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM contact_labels cl JOIN labels l ON l.id=cl.label_id WHERE cl.contact_id=c.id AND l.name=$%d)", i))
		args = append(args, f.Label)
		i++
	}
	allowedFields := map[string]string{
		"name": "c.name", "email": "c.email",
		"phone": "c.phone", "company": "c.company", "status": "c.status",
	}
	for _, cond := range f.Conditions {
		col, ok := allowedFields[cond.Field]
		if !ok {
			continue
		}
		switch cond.Op {
		case "contains":
			conds = append(conds, fmt.Sprintf("%s ILIKE $%d", col, i))
			args = append(args, "%"+cond.Value+"%")
			i++
		case "equals":
			conds = append(conds, fmt.Sprintf("%s ILIKE $%d", col, i))
			args = append(args, cond.Value)
			i++
		case "starts_with":
			conds = append(conds, fmt.Sprintf("%s ILIKE $%d", col, i))
			args = append(args, cond.Value+"%")
			i++
		case "present":
			conds = append(conds, fmt.Sprintf("(%s IS NOT NULL AND %s != '')", col, col))
		case "not_present":
			conds = append(conds, fmt.Sprintf("(%s IS NULL OR %s = '')", col, col))
		}
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM contacts c %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	q := fmt.Sprintf(`SELECT %s FROM contacts c %s ORDER BY c.updated_at DESC LIMIT $%d OFFSET $%d`,
		contactCols, where, i, i+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contacts []domain.Contact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, *c)
	}
	return contacts, total, nil
}

func (r *ContactRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	q := fmt.Sprintf(`SELECT %s FROM contacts c WHERE c.id = $1`, contactCols)
	c, err := scanContact(r.db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("find contact: %w", err)
	}
	return c, nil
}

func (r *ContactRepo) Create(ctx context.Context, name, email, phone, company *string, customAttrs map[string]any) (*domain.Contact, error) {
	cols := strings.ReplaceAll(contactCols, "c.", "")
	var caJSON []byte
	if len(customAttrs) > 0 {
		caJSON, _ = json.Marshal(customAttrs)
	} else {
		caJSON = []byte("{}")
	}
	q := fmt.Sprintf(`INSERT INTO contacts (name, email, phone, company, status, custom_attributes)
		VALUES (COALESCE($1,''), $2, $3, $4, 'novo', $5) RETURNING %s`, cols)
	return scanContact(r.db.QueryRow(ctx, q, name, email, phone, company, caJSON))
}

func (r *ContactRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]any) (*domain.Contact, error) {
	sets := []string{}
	args := []any{id}
	i := 2
	allowed := map[string]bool{
		"name": true, "email": true, "phone": true, "company": true,
		"avatar_url": true, "status": true, "score": true, "assigned_to": true,
		"company_id": true,
	}
	for k, v := range fields {
		if allowed[k] {
			sets = append(sets, fmt.Sprintf("%s=$%d", k, i))
			args = append(args, v)
			i++
		}
	}
	if len(sets) == 0 {
		return r.FindByID(ctx, id)
	}
	sets = append(sets, "updated_at=NOW()")
	q := fmt.Sprintf("UPDATE contacts SET %s WHERE id=$1", strings.Join(sets, ", "))
	if _, err := r.db.Exec(ctx, q, args...); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *ContactRepo) UpdateCustomAttributes(ctx context.Context, id uuid.UUID, attrs map[string]any) error {
	b, _ := json.Marshal(attrs)
	_, err := r.db.Exec(ctx, `UPDATE contacts SET custom_attributes=$2, updated_at=NOW() WHERE id=$1`, id, b)
	return err
}

func (r *ContactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM contacts WHERE id = $1`, id)
	return err
}

func (r *ContactRepo) GetLabels(ctx context.Context, contactID uuid.UUID) ([]domain.Label, error) {
	rows, err := r.db.Query(ctx,
		`SELECT l.id, l.name, l.color, l.description, l.created_at
		 FROM labels l JOIN contact_labels cl ON cl.label_id=l.id
		 WHERE cl.contact_id=$1 ORDER BY l.name`, contactID)
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

func (r *ContactRepo) SetLabels(ctx context.Context, contactID uuid.UUID, labelIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM contact_labels WHERE contact_id=$1`, contactID); err != nil {
		return err
	}
	for _, lid := range labelIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO contact_labels VALUES ($1,$2) ON CONFLICT DO NOTHING`, contactID, lid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
