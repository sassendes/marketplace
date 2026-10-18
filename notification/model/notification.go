package model

import "time"

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelInApp Channel = "in_app"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

type Notification struct {
	ID        int64      `json:"id" db:"id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	Channel   Channel    `json:"channel" db:"channel"`
	Subject   string     `json:"subject" db:"subject"`
	Body      string     `json:"body" db:"body"`
	Status    Status     `json:"status" db:"status"`
	Read      bool       `json:"read" db:"read"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	SentAt    *time.Time `json:"sent_at" db:"sent_at"`
}
