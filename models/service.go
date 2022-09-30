package models

import "gorm.io/gorm"

type Service struct {
	gorm.Model
	Name                   string `binding:"required"`
	RevenueAccountID       uint   `binding:"required"`
	RevenueAccount         *Account
	CostOfServiceAccountID uint `binding:"required"`
	CostOfServiceAccount   *Account
	CompanyID              uint
	Company                *Company
}

type ServicePerformed struct {
	gorm.Model
	Paid                bool
	Value               float64
	ServiceID           uint
	Service             *Service
	Consumptions        []*Consumption `binding:"required,dive,required"`
	CompanyID           uint           `json:"-"`
	Company             *Company       `json:"-"`
	PaymentAccountID    *uint          `binding:"required_if=Paid true"`
	PaymentAccount      *Account
	ReceivableAccountID *uint `binding:"required_if=Paid false"`
	ReceivableAccount   *Account
	StockUsages         []*StockUsage `json:"-" gorm:"polymorphic:Source"`
	Entries             []*Entry      `gorm:"polymorphic:Source"`
}

type Consumption struct {
	gorm.Model
	Qty                uint `binding:"required"`
	ProductID          uint `binding:"required"`
	Product            *Product
	ServicePerformedID uint
	ServicePerformed   *ServicePerformed
}
