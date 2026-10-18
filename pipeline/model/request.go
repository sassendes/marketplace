package model

type StartCheckoutRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
}

type CheckoutResult struct {
	SagaID  int64  `json:"saga_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}
