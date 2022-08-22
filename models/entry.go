package models

import "example.com/accounting/database"

type Entry struct {
	database.Model
	Description  string
	PurchaseID   *uint
	Purchase     *Purchase
	Transactions []*Transaction `gorm:"constraint:OnDelete:CASCADE;"`
}

func (e Entry) IsBalanced() bool {
	totalDebit := 0.0
	totalCredit := 0.0
	for _, transaction := range e.Transactions {
		if transaction.Account.TransactionType() == Debit {
			totalDebit += transaction.Value
		} else {
			totalCredit += transaction.Value
		}
	}
	return totalDebit == totalCredit
}

type Transaction struct {
	database.Model
	Value     float64
	AccountID uint
	Account   *Account
	EntryID   uint
	Entry     *Entry
}
