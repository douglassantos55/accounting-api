package models

import "gorm.io/gorm"

type Service struct {
	gorm.Model
	Name                   string `binding:"required"`
	RevenueAccountID       uint   `json:"revenue_account_id" binding:"required"`
	RevenueAccount         *Account
	CostOfServiceAccountID uint `json:"cost_of_service_account_id" binding:"required"`
	CostOfServiceAccount   *Account
	CompanyID              uint
	Company                *Company
}

type ServicePerformed struct {
	gorm.Model
	Value               float64
	ServiceID           uint `json:"service_id"`
	Service             *Service
	Consumptions        []*Consumption
	CompanyID           uint
	Company             *Company
	PaymentAccountID    *uint `json:"payment_account_id"`
	PaymentAccount      *Account
	ReceivableAccountID *uint `json:"receivable_account_id"`
	ReceivableAccount   *Account
	StockUsages         []*StockUsage `gorm:"polymorphic:Source"`
}

type Consumption struct {
	gorm.Model
	Qty                uint
	ProductID          uint `json:"product_id"`
	Product            *Product
	ServicePerformedID uint
	ServicePerformed   *ServicePerformed
}
