package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Attributes map[string]any

func (a Attributes) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (a *Attributes) Scan(src any) error {
	if src == nil {
		*a = Attributes{}
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for Attributes scan")
	}
	return json.Unmarshal(data, a)
}

type Variant struct {
	ID         int64      `json:"id"`
	ProductID  int64      `json:"product_id"`
	SKU        string     `json:"sku"`
	Price      int64      `json:"price"`
	Currency   string     `json:"currency"`
	Stock      int        `json:"-"`
	Available  int        `json:"available"`
	Attributes Attributes `json:"attributes"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
