package models

import (
	"example.com/accounting/database"
	"gorm.io/gorm"
)

type Entry struct {
	gorm.Model
	Description  string    `binding:"required"`
	PurchaseID   *uint     `json:"purchase_id"`
	Purchase     *Purchase `gorm:"constraint:OnDelete:CASCADE;"`
	CompanyID    uint
	Company      *Company       `gorm:"constraint:OnDelete:CASCADE"`
	Transactions []*Transaction `binding:"min=2" gorm:"constraint:OnDelete:CASCADE;"`
}

func (e Entry) IsBalanced() bool {
	totalDebit := 0.0
	totalCredit := 0.0
	for _, transaction := range e.Transactions {
		account := transaction.Account

		if account == nil {
			db, _ := database.GetConnection()
			db.First(&account, transaction.AccountID)
		}

		if account.TransactionType() == Debit {
			totalDebit += transaction.Value
		} else {
			totalCredit += transaction.Value
		}
	}
	return totalDebit == totalCredit
}

type Transaction struct {
	gorm.Model
	Value     float64 `binding:"required"`
	AccountID uint    `json:"account_id" binding:"required"`
	Account   *Account
	EntryID   uint
	Entry     *Entry `gorm:"constraint:OnDelete:CASCADE"`
}
