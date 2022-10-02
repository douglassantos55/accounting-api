package models

import (
	"gorm.io/gorm"
)

type AccountType uint
type TransactionType uint

const (
	Debit TransactionType = iota
	Credit
)

const (
	Asset AccountType = iota + 1
	Liability
	Equity
	Dividend
	Expense
	Revenue
)

type Account struct {
	gorm.Model
	Name         string      `binding:"required"`
	Type         AccountType `binding:"required"`
	ParentID     *uint       `binding:"omitempty"`
	Parent       *Account
	CompanyID    uint           `json:"-"`
	Company      *Company       `json:"-"`
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
