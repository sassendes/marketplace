package model

import "time"

type Stock struct {
	VariantID int64     `json:"variant_id"`
	OnHand    int       `json:"on_hand"`
	Reserved  int       `json:"reserved"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Stock) Available() int {
	return s.OnHand - s.Reserved
}

type Reservation struct {
	ID        int64     `json:"id"`
	VariantID int64     `json:"variant_id"`
	OrderID   int64     `json:"order_id"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
