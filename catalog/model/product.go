package model

import "time"

type ProductStatus string

const (
	StatusDraft    ProductStatus = "draft"
	StatusActive   ProductStatus = "active"
	StatusInactive ProductStatus = "inactive"
	StatusRemoved  ProductStatus = "removed"
)

func (s ProductStatus) SellerSettable() bool {
	return s == StatusDraft || s == StatusActive || s == StatusInactive
}

type Product struct {
	ID           int64         `json:"id"`
	SellerID     int64         `json:"seller_id"`
	StoreName    string        `json:"store_name"`
	Title        string        `json:"title"`
	Slug         string        `json:"slug"`
	Description  string        `json:"description"`
	CategoryID   *int64        `json:"category_id"`
	CategorySlug string        `json:"category_slug"`
	Brand        string        `json:"brand"`
	ImageURL     string        `json:"image_url"`
	Status       ProductStatus `json:"status"`
	PriceFrom    int64         `json:"price_from"`
	Currency     string        `json:"currency"`
	RatingAvg    float64       `json:"rating_avg"`
	RatingCount  int           `json:"rating_count"`
	Available    int           `json:"available"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	Variants []*Variant `json:"variants,omitempty"`
}
