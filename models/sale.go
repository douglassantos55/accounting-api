package models

import (
	"gorm.io/gorm"
)

type Sale struct {
	gorm.Model
	Paid              bool
	Items             []*Item  `gorm:"constraint:OnDelete:CASCADE;" binding:"min=1"`
	Entries           []*Entry `gorm:"polymorphic:Source"`
	Customer          *Customer
	Company           *Company
	PaymentAccount    *Account      `gorm:"constraint:OnDelete:SET NULL;"`
	ReceivableAccount *Account      `gorm:"constraint:OnDelete:SET NULL;"`
	StockUsages       []*StockUsage `gorm:"polymorphic:Source"`

	CustomerID          uint `json:"customer_id" binding:"required"`
	CompanyID           uint
	PaymentAccountID    *uint `json:"payment_account_id"`
	ReceivableAccountID *uint `json:"receivable_account_id"`
}

func (s Sale) Total() float64 {
	total := 0.0
	for _, item := range s.Items {
		total += item.Subtotal()
	}
	return total
}

type Item struct {
	gorm.Model
	Qty       uint
	Price     float64
	ProductID uint `json:"product_id"`
	Product   *Product
	SaleID    uint
	Sale      *Sale
}

func (i Item) Subtotal() float64 {
	return float64(i.Qty) * i.Price
}
