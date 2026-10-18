package repository

import (
	"context"
	"database/sql"
	"errors"

	"marketplace/auth/model"
	"marketplace/platform/database"
)

var ErrDuplicateEmail = errors.New("email already registered")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `id, username, email, password_hash, role, email_verified, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	u := &model.User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, u *model.User) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (username, email, password_hash, role, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		u.Username, u.Email, u.PasswordHash, u.Role, u.EmailVerified, u.CreatedAt, u.UpdatedAt,
	).Scan(&u.ID)
	if database.IsUniqueViolation(err) {
		return ErrDuplicateEmail
	}
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email))
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}
