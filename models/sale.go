package models

import (
	"gorm.io/gorm"
)

type Sale struct {
	gorm.Model
	Paid              bool
	Items             []*Item  `gorm:"constraint:OnDelete:CASCADE;" binding:"min=1,required,dive,required"`
	Entries           []*Entry `gorm:"polymorphic:Source"`
	Customer          *Customer
	Company           *Company
	PaymentAccount    *Account      `gorm:"constraint:OnDelete:SET NULL;"`
	ReceivableAccount *Account      `gorm:"constraint:OnDelete:SET NULL;"`
	StockUsages       []*StockUsage `gorm:"polymorphic:Source"`

	CustomerID          uint `binding:"required"`
	CompanyID           uint
	PaymentAccountID    *uint `binding:"required_if=Paid true"`
	ReceivableAccountID *uint `binding:"required_if=Paid false"`
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
	Qty       uint    `binding:"required,min=1"`
	Price     float64 `binding:"required"`
	ProductID uint    `binding:"required"`
	Product   *Product
	SaleID    uint
	Sale      *Sale
}

func (i Item) Subtotal() float64 {
	return float64(i.Qty) * i.Price
}
