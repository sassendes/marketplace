package model

import "time"

const (
	StatusPending  = "pending"
	StatusCaptured = "captured"
	StatusFailed   = "failed"
	StatusRefunded = "refunded"
)

type Payment struct {
	ID             int64     `json:"id"`
	IdempotencyKey string    `json:"idempotency_key"`
	OrderID        int64     `json:"order_id"`
	BuyerID        int64     `json:"buyer_id"`
	AmountCents    int64     `json:"amount_cents"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	ProviderRef    string    `json:"provider_ref"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LedgerEntry struct {
	ID          int64     `json:"id"`
	PaymentID   int64     `json:"payment_id"`
	Account     string    `json:"account"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
}
