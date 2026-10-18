package model

type ModerateRequest struct {
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	Action     string `json:"action"`
	Reason     string `json:"reason"`
}
