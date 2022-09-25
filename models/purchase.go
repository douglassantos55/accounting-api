package models

import (
	"time"

	"gorm.io/gorm"
)

type Purchase struct {
	gorm.Model
	Qty              uint    `binding:"required"`
	Price            float64 `binding:"required"`
	Paid             bool
	PaymentDate      time.Time `binding:"required_if=Paid true"`
	CompanyID        uint
	Company          *Company
	PayableAccountID *uint    `binding:"required_if=Paid false"`
	PayableAccount   *Account `gorm:"foreignKey:PayableAccountID;constraint:OnDelete:SET NULL;"`
	PaymentAccountID *uint    `binding:"required_if=Paid true"`
	PaymentAccount   *Account `gorm:"foreignKey:PaymentAccountID;constraint:OnDelete:SET NULL;"`
	ProductID        uint     `binding:"required"`
	Product          *Product `gorm:"constraint:OnDelete:CASCADE;"`
	StockEntryID     *uint
	StockEntry       *StockEntry `gorm:"constraint:OnDelete:SET NULL;"`
	PaymentEntryID   *uint
	PaymentEntry     *Entry `gorm:"foreignKey:PaymentEntryID;constraint:OnDelete:CASCADE;"`
	PayableEntryID   *uint
	PayableEntry     *Entry `gorm:"foreignKey:PayableEntryID;constraint:OnDelete:CASCADE;"`
}
