package model

import "time"

type SellerStatus string

const (
	StatusPending   SellerStatus = "pending"
	StatusActive    SellerStatus = "active"
	StatusSuspended SellerStatus = "suspended"
)

type Seller struct {
	ID              int64        `json:"id"`
	UserID          int64        `json:"user_id"`
	StoreName       string       `json:"store_name"`
	Slug            string       `json:"slug"`
	Description     string       `json:"description"`
	Status          SellerStatus `json:"status"`
	CommissionRate  int          `json:"commission_rate"`
	StripeAccountID string       `json:"stripe_account_id"`
	PayoutEnabled   bool         `json:"payout_enabled"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type PublicSeller struct {
	ID          int64        `json:"id"`
	StoreName   string       `json:"store_name"`
	Slug        string       `json:"slug"`
	Description string       `json:"description"`
	Status      SellerStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
}

func (s *Seller) Public() *PublicSeller {
	return &PublicSeller{ID: s.ID, StoreName: s.StoreName, Slug: s.Slug, Description: s.Description, Status: s.Status, CreatedAt: s.CreatedAt}
}

type Payout struct {
	ID        int64     `json:"id"`
	SellerID  int64     `json:"seller_id"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
