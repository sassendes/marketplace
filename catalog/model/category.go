package model

import "time"

type Category struct {
	ID           int64     `json:"id"`
	ParentID     *int64    `json:"parent_id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	ProductCount int       `json:"product_count"`
	CreatedAt    time.Time `json:"created_at"`
}
