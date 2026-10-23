package repository

import (
	"context"
	"database/sql"
	"errors"

	"mercatus/auth/model"
)

var ErrTokenNotActive = errors.New("refresh token is not active")

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateRefreshToken(ctx context.Context, t *model.RefreshToken) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at, revoked, created_at)
		 VALUES ($1, $2, $3, false, $4) RETURNING id`,
		t.UserID, t.TokenHash, t.ExpiresAt, t.CreatedAt,
	).Scan(&t.ID)
}

func (r *TokenRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	t := &model.RefreshToken{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, expires_at, revoked, created_at FROM refresh_tokens WHERE token = $1`, tokenHash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.Revoked, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TokenRepository) Rotate(ctx context.Context, oldHash string, next *model.RefreshToken) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE token = $1 AND revoked = false AND expires_at > now()`, oldHash)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return ErrTokenNotActive
	}
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at, revoked, created_at)
		 VALUES ($1, $2, $3, false, $4) RETURNING id`,
		next.UserID, next.TokenHash, next.ExpiresAt, next.CreatedAt,
	).Scan(&next.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked = true WHERE token = $1`, tokenHash)
	return err
}

func (r *TokenRepository) RevokeAllForUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`, userID)
	return err
}

func (r *TokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < now() - interval '1 day'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *TokenRepository) CreateEmailVerification(ctx context.Context, ev *model.EmailVerification) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO email_verifications (user_id, code, expires_at, used, created_at)
		 VALUES ($1, $2, $3, false, $4) RETURNING id`,
		ev.UserID, ev.Code, ev.ExpiresAt, ev.CreatedAt,
	).Scan(&ev.ID)
}

func (r *TokenRepository) ConsumeEmailVerification(ctx context.Context, userID int64, code string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		`UPDATE email_verifications SET used = true
		 WHERE user_id = $1 AND code = $2 AND used = false AND expires_at > now()`, userID, code)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET email_verified = true, updated_at = now() WHERE id = $1`, userID); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
