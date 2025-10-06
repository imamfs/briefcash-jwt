package entity

import "time"

type MerchantSettings struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	MerchantCode string `gorm:"column:merchant_code"`
	ApiKey       string `gorm:"column:api_key"`
	ApiSecret    string `gorm:"column:api_secret"`
	ChannelId    string `gorm:"column:channel_id"`
	PartnerId    string `gorm:"column:partner_id"`
}

type JwtToken struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	MerchantID   string    `gorm:"column:merchant_settings_id"`
	AccessToken  string    `gorm:"column:access_token"`
	RefreshToken string    `gorm:"column:refresh_token"`
	ExpiresAt    time.Time `gorm:"column:expires_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	IsRevoke     string    `gorm:"column:is_revoke"`
}
