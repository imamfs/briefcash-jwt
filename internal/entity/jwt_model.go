package entity

import "time"

type JwtToken struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	MerchantID   string    `gorm:"column:merchant_settings_id"`
	AccessToken  string    `gorm:"column:access_token"`
	RefreshToken string    `gorm:"column:refresh_token"`
	ExpiresAt    time.Time `gorm:"column:expires_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	IsRevoke     bool      `gorm:"column:is_revoke"`
}
