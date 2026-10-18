package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"marketplace/pipeline/model"
	"marketplace/platform/database"
)

var ErrDuplicateKey = errors.New("checkout already started with this key")

type SagaRepository struct{ db *sql.DB }

func NewSagaRepository(db *sql.DB) *SagaRepository { return &SagaRepository{db: db} }

const sagaColumns = `id, order_id, buyer_id, COALESCE(idempotency_key, ''), payment_id, status, current_step, error, created_at, updated_at`

func scanSaga(row interface{ Scan(...any) error }) (*model.Saga, error) {
	s := &model.Saga{}
	err := row.Scan(&s.ID, &s.OrderID, &s.BuyerID, &s.IdempotencyKey, &s.PaymentID, &s.Status, &s.CurrentStep,
		&s.Error, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SagaRepository) Create(ctx context.Context, s *model.Saga) error {
	var key any
	if s.IdempotencyKey != "" {
		key = s.IdempotencyKey
	}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO sagas (order_id, buyer_id, idempotency_key, status, current_step, created_at, updated_at)
		 VALUES ($1, $2, $3, 'running', '', now(), now()) RETURNING id, status, created_at, updated_at`,
		s.OrderID, s.BuyerID, key,
	).Scan(&s.ID, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if database.IsUniqueViolation(err) {
		return ErrDuplicateKey
	}
	return err
}

func (r *SagaRepository) SetOrder(ctx context.Context, sagaID, orderID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sagas SET order_id = $1, updated_at = now() WHERE id = $2`, orderID, sagaID)
	return err
}

func (r *SagaRepository) SetPayment(ctx context.Context, sagaID, paymentID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sagas SET payment_id = $1, updated_at = now() WHERE id = $2`, paymentID, sagaID)
	return err
}

func (r *SagaRepository) SetStep(ctx context.Context, sagaID int64, step string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sagas SET current_step = $1, updated_at = now() WHERE id = $2`, step, sagaID)
	return err
}

func (r *SagaRepository) SetStatus(ctx context.Context, sagaID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sagas SET status = $1, updated_at = now() WHERE id = $2`, status, sagaID)
	return err
}

func (r *SagaRepository) SetError(ctx context.Context, sagaID int64, msg string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sagas SET error = $1, updated_at = now() WHERE id = $2`, msg, sagaID)
	return err
}

func (r *SagaRepository) RecordStep(ctx context.Context, sagaID int64, name, status string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO saga_steps (saga_id, name, status, created_at) VALUES ($1, $2, $3, now())`, sagaID, name, status)
	return err
}

func (r *SagaRepository) MarkStepCompensated(ctx context.Context, sagaID int64, name string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE saga_steps SET status = 'compensated' WHERE saga_id = $1 AND name = $2 AND status = 'done'`, sagaID, name)
	return err
}

func (r *SagaRepository) Get(ctx context.Context, id int64) (*model.Saga, error) {
	return scanSaga(r.db.QueryRowContext(ctx, `SELECT `+sagaColumns+` FROM sagas WHERE id = $1`, id))
}

func (r *SagaRepository) FindByKey(ctx context.Context, buyerID int64, key string) (*model.Saga, error) {
	return scanSaga(r.db.QueryRowContext(ctx,
		`SELECT `+sagaColumns+` FROM sagas WHERE buyer_id = $1 AND idempotency_key = $2`, buyerID, key))
}

func (r *SagaRepository) Steps(ctx context.Context, sagaID int64) ([]*model.SagaStep, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, saga_id, name, status, created_at FROM saga_steps WHERE saga_id = $1 ORDER BY id`, sagaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.SagaStep{}
	for rows.Next() {
		st := &model.SagaStep{}
		if err := rows.Scan(&st.ID, &st.SagaID, &st.Name, &st.Status, &st.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (r *SagaRepository) ClaimStale(ctx context.Context, olderThan time.Duration, limit int) ([]*model.Saga, error) {
	rows, err := r.db.QueryContext(ctx,
		`UPDATE sagas SET status = 'compensating', updated_at = now()
		 WHERE id IN (
			SELECT id FROM sagas
			WHERE status IN ('running', 'compensating') AND updated_at < now() - make_interval(secs => $1)
			ORDER BY id LIMIT $2 FOR UPDATE SKIP LOCKED)
		 RETURNING `+sagaColumns, olderThan.Seconds(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Saga
	for rows.Next() {
		s, err := scanSaga(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
