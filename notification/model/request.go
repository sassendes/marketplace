package model

type SendRequest struct {
	UserID  int64   `json:"user_id"`
	Channel Channel `json:"channel"`
	Subject string  `json:"subject"`
	Body    string  `json:"body"`
}
