package entity

import "time"

type Merchant struct {
	Code        string     `gorm:"column:code;primaryKey"`
	CompanyName string     `gorm:"column:company_name"`
	Address     string     `gorm:"column:address"`
	Email       string     `gorm:"column:email"`
	Phone       string     `gorm:"column:phone"`
	Website     string     `gorm:"column:website"`
	IsActive    bool       `gorm:"column:is_active"`
	DateJoined  *time.Time `gorm:"column:date_joined"`
}
