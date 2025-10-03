package entity

type MerchantSettings struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	MerchantCode string `gorm:"column:merchant_code"`
	ApiKey       string `gorm:"column:api_key"`
	ApiSecret    string `gorm:"column:api_secret"`
	ChannelId    string `gorm:"column:channel_id"`
	PartnerId    string `gorm:"column:partner_id"`
}

type JwtToken struct {
	ID        int64            `gorm:"column:id;primaryKey"`
	Merchant  MerchantSettings `gorm:"column:merchant_settings_id;embedded"`
	Token     string           `gorm:"column:token"`
	Jti       string           `gorm:"column:jti"`
	ExpiresAt string           `gorm:"column:expires_at"`
	CreatedAt string           `gorm:"column:created_at"`
	IsRevoke  string           `gorm:"column:is_revoke"`
}
