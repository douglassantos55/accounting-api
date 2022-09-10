package models

import "gorm.io/gorm"

type Service struct {
	gorm.Model
	Name             string `binding:"required"`
	RevenueAccountID uint   `json:"revenue_account_id" binding:"required"`
	RevenueAccount   *Account
	CompanyID        uint
	Company          *Company
}

type ServicePerformed struct {
	gorm.Model
	ServiceID    uint
	Service      *Service
	Consumptions []*Consumption
}

type Consumption struct {
	gorm.Model
	Qty                uint
	ProductID          uint
	Product            *Product
	ServicePerformedID uint
	ServicePerformed   *ServicePerformed
}
