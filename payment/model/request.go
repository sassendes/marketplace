package model

type ChargeRequest struct {
	IdempotencyKey string
	OrderID        int64
	AmountCents    int64
	Currency       string
	Splits         []SellerSplit
}

type SellerSplit struct {
	SellerID        int64
	SubtotalCents   int64
	CommissionCents int64
}
