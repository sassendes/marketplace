package model

import "time"

const (
	SagaRunning      = "running"
	SagaCompleted    = "completed"
	SagaCompensating = "compensating"
	SagaFailed       = "failed"
)

type Saga struct {
	ID             int64     `json:"id"`
	OrderID        int64     `json:"order_id"`
	BuyerID        int64     `json:"buyer_id"`
	IdempotencyKey string    `json:"-"`
	PaymentID      *int64    `json:"payment_id,omitempty"`
	Status         string    `json:"status"`
	CurrentStep    string    `json:"current_step"`
	Error          string    `json:"error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SagaStep struct {
	ID        int64     `json:"id"`
	SagaID    int64     `json:"saga_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
