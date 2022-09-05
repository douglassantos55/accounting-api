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
	PaymentDate      time.Time `json:"payment_date"`
	CompanyID        uint
	Company          *Company
	PayableAccountID *uint    `json:"payable_account_id"`
	PayableAccount   *Account `gorm:"foreignKey:PayableAccountID;constraint:OnDelete:SET NULL;"`
	PaymentAccountID *uint    `json:"payment_account_id"`
	PaymentAccount   *Account `gorm:"foreignKey:PaymentAccountID;constraint:OnDelete:SET NULL;"`
	ProductID        uint     `json:"product_id" binding:"required"`
	Product          *Product `gorm:"constraint:OnDelete:CASCADE;"`
	StockEntryID     *uint
	StockEntry       *StockEntry `gorm:"constraint:OnDelete:SET NULL;"`
	PaymentEntryID   *uint
	PaymentEntry     *Entry `gorm:"foreignKey:PaymentEntryID;constraint:OnDelete:CASCADE;"`
	PayableEntryID   *uint
	PayableEntry     *Entry `gorm:"foreignKey:PayableEntryID;constraint:OnDelete:CASCADE;"`
}
