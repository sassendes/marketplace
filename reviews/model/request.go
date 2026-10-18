package model

type CreateReviewRequest struct {
	ProductID int64  `json:"product_id"`
	Rating    int    `json:"rating"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}
