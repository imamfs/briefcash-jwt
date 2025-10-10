package dto

type JwtRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
}

type JwtRefreshToken struct {
	UserID       string `json:"user_id"`
	Type         string `json:"type"`
	RefreshToken string `json:"refresh_token"`
}
