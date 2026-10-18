package model

import "time"

type Review struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	UserID    int64     `json:"user_id"`
	Author    string    `json:"author"`
	Rating    int       `json:"rating"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Verified  bool      `json:"verified_purchase"`
	Helpful   int       `json:"helpful"`
	CreatedAt time.Time `json:"created_at"`
}

type RatingSummary struct {
	ProductID int64       `json:"product_id"`
	Average   float64     `json:"average"`
	Count     int         `json:"count"`
	Breakdown map[int]int `json:"breakdown"`
}
