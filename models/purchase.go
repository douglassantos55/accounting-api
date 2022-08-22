package models

import "example.com/accounting/database"

type Purchase struct {
	database.Model
	Qty              uint
	Price            float64
	PaymentAccountID *uint
	PaymentAccount   *Account `gorm:"constraint:OnDelete:SET NULL;"`
	ProductID        uint
	Product          *Product `gorm:"constraint:OnDelete:CASCADE;"`
	StockEntryID     *uint
	StockEntry       *StockEntry `gorm:"constraint:OnDelete:SET NULL;"`
	Entries          []*Entry    `gorm:"constraint:OnDelete:SET NULL;"`
}
