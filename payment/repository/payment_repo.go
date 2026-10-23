package repository

import (
	"context"
	"database/sql"
	"errors"

	"mercatus/payment/model"
	"mercatus/platform/database"
)

var (
	ErrDuplicateKey   = errors.New("idempotency key already used")
	ErrNotTransitable = errors.New("payment is not in the required state")
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

const paymentColumns = `id, idempotency_key, order_id, buyer_id, amount_cents, currency, status, provider_ref, created_at, updated_at`

func scanPayment(row interface{ Scan(...any) error }) (*model.Payment, error) {
	p := &model.Payment{}
	err := row.Scan(&p.ID, &p.IdempotencyKey, &p.OrderID, &p.BuyerID, &p.AmountCents, &p.Currency,
		&p.Status, &p.ProviderRef, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*model.Payment, error) {
	return scanPayment(r.db.QueryRowContext(ctx, `SELECT `+paymentColumns+` FROM payments WHERE idempotency_key = $1`, key))
}

func (r *PaymentRepository) FindByID(ctx context.Context, id int64) (*model.Payment, error) {
	return scanPayment(r.db.QueryRowContext(ctx, `SELECT `+paymentColumns+` FROM payments WHERE id = $1`, id))
}

func (r *PaymentRepository) FindCapturedForOrder(ctx context.Context, orderID int64) ([]*model.Payment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+paymentColumns+` FROM payments WHERE order_id = $1 AND status IN ('captured', 'pending')`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PaymentRepository) InsertPending(ctx context.Context, p *model.Payment) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO payments (idempotency_key, order_id, buyer_id, amount_cents, currency, status, provider_ref, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 'pending', '', now(), now())
		 RETURNING id, created_at, updated_at`,
		p.IdempotencyKey, p.OrderID, p.BuyerID, p.AmountCents, p.Currency,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if database.IsUniqueViolation(err) {
		return ErrDuplicateKey
	}
	if err != nil {
		return err
	}
	p.Status = model.StatusPending
	return nil
}

func (r *PaymentRepository) transitionWithLedger(ctx context.Context, paymentID int64, from, to, providerRef string, entries []model.LedgerEntry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		`UPDATE payments SET status = $1, provider_ref = COALESCE(NULLIF($2, ''), provider_ref), updated_at = now()
		 WHERE id = $3 AND status = $4`, to, providerRef, paymentID, from)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotTransitable
	}
	var sum int64
	for _, e := range entries {
		sum += e.AmountCents
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO ledger_entries (payment_id, account, amount_cents, currency, created_at)
			 VALUES ($1, $2, $3, $4, now())`, paymentID, e.Account, e.AmountCents, e.Currency); err != nil {
			return err
		}
	}
	if sum != 0 {
		return errors.New("ledger entries do not balance")
	}
	return tx.Commit()
}

func (r *PaymentRepository) Capture(ctx context.Context, paymentID int64, providerRef string, entries []model.LedgerEntry) error {
	return r.transitionWithLedger(ctx, paymentID, model.StatusPending, model.StatusCaptured, providerRef, entries)
}

func (r *PaymentRepository) Refund(ctx context.Context, paymentID int64, entries []model.LedgerEntry) error {
	return r.transitionWithLedger(ctx, paymentID, model.StatusCaptured, model.StatusRefunded, "", entries)
}

func (r *PaymentRepository) SetProviderRef(ctx context.Context, paymentID int64, providerRef string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payments SET provider_ref = $1, updated_at = now() WHERE id = $2 AND status = 'pending'`, providerRef, paymentID)
	return err
}

func (r *PaymentRepository) Fail(ctx context.Context, paymentID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payments SET status = 'failed', updated_at = now() WHERE id = $1 AND status = 'pending'`, paymentID)
	return err
}

func (r *PaymentRepository) Entries(ctx context.Context, paymentID int64) ([]model.LedgerEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, payment_id, account, amount_cents, currency, created_at FROM ledger_entries WHERE payment_id = $1 ORDER BY id`, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.LedgerEntry
	for rows.Next() {
		var e model.LedgerEntry
		if err := rows.Scan(&e.ID, &e.PaymentID, &e.Account, &e.AmountCents, &e.Currency, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PaymentRepository) LedgerBalance(ctx context.Context, account string) (int64, error) {
	var bal int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0) FROM ledger_entries WHERE account = $1`, account).Scan(&bal)
	return bal, err
}

func (r *PaymentRepository) LedgerSumForPayment(ctx context.Context, paymentID int64) (int64, error) {
	var sum int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0) FROM ledger_entries WHERE payment_id = $1`, paymentID).Scan(&sum)
	return sum, err
}

func (r *PaymentRepository) LedgerTotal(ctx context.Context) (int64, error) {
	var sum int64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount_cents), 0) FROM ledger_entries`).Scan(&sum)
	return sum, err
}
