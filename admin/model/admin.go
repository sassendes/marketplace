package model

import "time"

type ModerationAction struct {
	ID         int64     `json:"id"`
	AdminID    int64     `json:"admin_id"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Action     string    `json:"action"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

type PlatformStats struct {
	Users            int64 `json:"users"`
	ActiveSellers    int64 `json:"active_sellers"`
	SuspendedSellers int64 `json:"suspended_sellers"`
	ActiveProducts   int64 `json:"active_products"`
	RemovedProducts  int64 `json:"removed_products"`
	PaidOrders       int64 `json:"paid_orders"`
	GMVCents         int64 `json:"gmv_cents"`
	CommissionCents  int64 `json:"commission_cents"`
	TotalActions     int64 `json:"total_actions"`
}
