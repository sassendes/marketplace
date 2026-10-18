package model

type SetStockRequest struct {
	VariantID int64 `json:"variant_id"`
	OnHand    int   `json:"on_hand"`
}

type ReserveRequest struct {
	VariantID int64 `json:"variant_id"`
	OrderID   int64 `json:"order_id"`
	Quantity  int   `json:"quantity"`
}

type Line struct {
	VariantID int64
	Quantity  int
}
