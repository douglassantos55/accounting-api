package models

import "example.com/accounting/database"

type AccountType uint
type TransactionType uint

const (
	Debit TransactionType = iota
	Credit
)

const (
	Asset AccountType = iota
	Liability
	Equity
	Dividend
	Expense
	Revenue
)

type Account struct {
	database.Model
	Name         string
	Type         AccountType
	ParentID     *uint
	Parent       *Account
	Children     []*Account     `gorm:"foreignKey:ParentID; constraint:OnDelete:CASCADE;"`
	Transactions []*Transaction `gorm:"constraint:OnDelete:CASCADE;"`
}

func (a Account) Balance() float64 {
	balance := 0.0
	for _, transaction := range a.Transactions {
		balance += transaction.Value
	}
	return balance
}

func (a Account) TransactionType() TransactionType {
	switch a.Type {
	case Dividend, Expense, Asset:
		return Debit
	case Liability, Equity, Revenue:
		return Credit
	default:
		panic("Invalid account type")
	}
}
