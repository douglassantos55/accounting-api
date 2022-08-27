package models

import (
	"time"

	"example.com/accounting/database"
)

type Purchase struct {
	database.Model
	Qty              uint
	Price            float64
	Paid             bool
	PaymentDate      time.Time
	CompanyID        uint
	Company          *Company
	PayableAccountID *uint
	PayableAccount   *Account `gorm:"foreignKey:PayableAccountID;constraint:OnDelete:SET NULL;"`
	PaymentAccountID *uint
	PaymentAccount   *Account `gorm:"foreignKey:PaymentAccountID;constraint:OnDelete:SET NULL;"`
	ProductID        uint
	Product          *Product `gorm:"constraint:OnDelete:CASCADE;"`
	StockEntryID     *uint
	StockEntry       *StockEntry `gorm:"constraint:OnDelete:SET NULL;"`
	PaymentEntryID   *uint
	PaymentEntry     *Entry `gorm:"foreignKey:PaymentEntryID;constraint:OnDelete:CASCADE;"`
	PayableEntryID   *uint
	PayableEntry     *Entry `gorm:"foreignKey:PayableEntryID;constraint:OnDelete:CASCADE;"`
}
