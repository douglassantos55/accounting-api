package models

import (
	"gorm.io/gorm"
)

type Service struct {
	gorm.Model
	Name      string
	AccountID uint
	Account   *Account
	CompanyID uint
	Company   *Company
}
