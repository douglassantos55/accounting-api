package models

import (
	"gorm.io/gorm"
)

type Sale struct {
	gorm.Model
	Paid              bool
	Items             []*Item `gorm:"constraint:OnDelete:CASCADE;"`
	Customer          *Customer
	Company           *Company
	PaymentAccount    *Account `gorm:"constraint:OnDelete:SET NULL;"`
	ReceivableAccount *Account `gorm:"constraint:OnDelete:SET NULL;"`

	CustomerID          uint
	CompanyID           uint
	PaymentAccountID    *uint
	ReceivableAccountID *uint
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
	ProductID uint
	Product   *Product
	SaleID    uint
	Sale      *Sale
}

func (i Item) Subtotal() float64 {
	return float64(i.Qty) * i.Price
}
