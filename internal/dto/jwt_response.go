package dto

type JwtResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    string `json:"created_at"`
	ExpiresAt    string `json:"expires_at"`
}
