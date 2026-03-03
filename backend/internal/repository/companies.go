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

type CompanyRepo struct{ db *pgxpool.Pool }

func NewCompanyRepo(db *pgxpool.Pool) *CompanyRepo { return &CompanyRepo{db: db} }

const companyCols = `id, name, domain, phone, website, industry, description, custom_attributes, created_at, updated_at`

var companyAllowedSort = map[string]string{
	"name":       "name",
	"domain":     "domain",
	"created_at": "created_at",
}

func buildCompanyOrderBy(sort string) string {
	dir := "ASC"
	col := sort
	if strings.HasPrefix(sort, "-") {
		dir = "DESC"
		col = sort[1:]
	}
	if mapped, ok := companyAllowedSort[col]; ok {
		return fmt.Sprintf("%s %s", mapped, dir)
	}
	return "name ASC"
}

func scanCompany(row interface{ Scan(...any) error }) (*domain.Company, error) {
	c := &domain.Company{}
	var caRaw []byte
	err := row.Scan(&c.ID, &c.Name, &c.Domain, &c.Phone, &c.Website, &c.Industry, &c.Description, &caRaw, &c.CreatedAt, &c.UpdatedAt, &c.ContactsCount)
	if err != nil {
		return nil, err
	}
	c.CustomAttributes = make(map[string]any)
	if len(caRaw) > 0 {
		_ = json.Unmarshal(caRaw, &c.CustomAttributes)
	}
	return c, nil
}

const companyColsWithCount = `id, name, domain, phone, website, industry, description, custom_attributes, created_at, updated_at, (SELECT COUNT(*) FROM contacts WHERE company_id = companies.id) AS contacts_count`

func (r *CompanyRepo) List(ctx context.Context, q string, page, limit int, sort string) ([]domain.Company, int, error) {
	args := []any{}
	where := ""
	if q != "" {
		where = "WHERE name ILIKE $1"
		args = append(args, "%"+q+"%")
	}

	var total int
	if err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM companies %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := buildCompanyOrderBy(sort)
	offset := (page - 1) * limit
	paramIdx := len(args) + 1
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx,
		fmt.Sprintf("SELECT %s FROM companies %s ORDER BY %s LIMIT $%d OFFSET $%d", companyColsWithCount, where, orderBy, paramIdx, paramIdx+1),
		args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var companies []domain.Company
	for rows.Next() {
		c, err := scanCompany(rows)
		if err != nil {
			return nil, 0, err
		}
		companies = append(companies, *c)
	}
	return companies, total, nil
}

func (r *CompanyRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	c, err := scanCompany(r.db.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM companies WHERE id=$1", companyColsWithCount), id))
	if err != nil {
		return nil, fmt.Errorf("find company: %w", err)
	}
	return c, nil
}

func (r *CompanyRepo) Create(ctx context.Context, name string, domain_, phone, website, industry, description *string) (*domain.Company, error) {
	// Insert then fetch with count
	var newID uuid.UUID
	err := r.db.QueryRow(ctx,
		`INSERT INTO companies (name, domain, phone, website, industry, description)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		name, domain_, phone, website, industry, description).Scan(&newID)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, newID)
}

func (r *CompanyRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]any) (*domain.Company, error) {
	allowed := map[string]bool{"name": true, "domain": true, "phone": true, "website": true, "industry": true, "description": true}
	sets := []string{}
	args := []any{id}
	i := 2
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
	if _, err := r.db.Exec(ctx, fmt.Sprintf("UPDATE companies SET %s WHERE id=$1", strings.Join(sets, ",")), args...); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *CompanyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM companies WHERE id=$1", id)
	return err
}

func (r *CompanyRepo) ListContacts(ctx context.Context, companyID uuid.UUID) ([]domain.Contact, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, email, phone, company, avatar_url, status, score, assigned_to, custom_attributes, created_at, updated_at, company_id
		 FROM contacts WHERE company_id=$1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contacts []domain.Contact
	for rows.Next() {
		c := domain.Contact{}
		var caRaw []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Company, &c.AvatarURL, &c.Status, &c.Score, &c.AssignedTo, &caRaw, &c.CreatedAt, &c.UpdatedAt, &c.CompanyID); err != nil {
			return nil, err
		}
		c.CustomAttributes = make(map[string]any)
		if len(caRaw) > 0 {
			_ = json.Unmarshal(caRaw, &c.CustomAttributes)
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}
