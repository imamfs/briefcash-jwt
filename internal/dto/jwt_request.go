package dto

type JwtRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
}
