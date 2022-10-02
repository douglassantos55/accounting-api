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
	PayableAccount   *Account `gorm:"foreignKey:PayableAccountID;"`
	PaymentAccountID *uint    `binding:"required_if=Paid true"`
	PaymentAccount   *Account `gorm:"foreignKey:PaymentAccountID;"`
	ProductID        uint     `binding:"required"`
	Product          *Product
	StockEntryID     *uint
	StockEntry       *StockEntry `gorm:"constraint:OnDelete:CASCADE;"`
	PaymentEntry     *Entry      `gorm:"polymorphic:Source;polymorphicValue:PurchasePayment;constraint:OnDelete:CASCADE;"`
	PayableEntry     *Entry      `gorm:"polymorphic:Source;polymorphicValue:PurchasePayable;constraint:OnDelete:CASCADE;"`
}
