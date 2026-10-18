package model

type AddItemRequest struct {
	VariantID int64 `json:"variant_id"`
	Quantity  int   `json:"quantity"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity"`
}
