package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simplix/api/internal/domain"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, availability, avatar_url, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Availability, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, availability, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Availability, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, email, password_hash, role, availability, avatar_url, created_at, updated_at
		 FROM users ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []domain.User
	for rows.Next() {
		u := domain.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Availability, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) Create(ctx context.Context, name, email, passwordHash string, role domain.Role) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, email, password_hash, role, availability, avatar_url, created_at, updated_at`,
		name, email, passwordHash, role,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Availability, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepo) Update(ctx context.Context, id uuid.UUID, name string, availability domain.Availability) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx,
		`UPDATE users SET name=$2, availability=$3, updated_at=NOW()
		 WHERE id=$1
		 RETURNING id, name, email, password_hash, role, availability, avatar_url, created_at, updated_at`,
		id, name, availability,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Availability, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, name string, availability domain.Availability, avatarURL *string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx,
		`UPDATE users SET name=$2, availability=$3, avatar_url=COALESCE($4, avatar_url), updated_at=NOW()
		 WHERE id=$1
		 RETURNING id, name, email, password_hash, role, availability, avatar_url, created_at, updated_at`,
		id, name, availability, avatarURL,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Availability, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
