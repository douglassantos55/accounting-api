package models

import (
	"example.com/accounting/database"
	"gorm.io/gorm"
)

type Entry struct {
	gorm.Model
	Description  string
	PurchaseID   *uint
	Purchase     *Purchase `gorm:"constraint:OnDelete:CASCADE;"`
	CompanyID    uint
	Company      *Company
	Transactions []*Transaction `gorm:"constraint:OnDelete:CASCADE;"`
}

func (e Entry) IsBalanced() bool {
	totalDebit := 0.0
	totalCredit := 0.0
	for _, transaction := range e.Transactions {
		account := transaction.Account

		if account == nil {
			db, _ := database.GetConnection()
			db.Find(Account{}).Where("ID", transaction.AccountID).First(&account)
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
	Value     float64
	AccountID uint
	Account   *Account
	EntryID   uint
	Entry     *Entry `gorm:"constraint:OnDelete:CASCADE"`
}
