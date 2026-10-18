package model

import "time"

type MediaAsset struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	Kind        string    `json:"kind"`
	ObjectKey   string    `json:"object_key"`
	ContentType string    `json:"content_type"`
	Status      string    `json:"status"`
	URL         string    `json:"url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
