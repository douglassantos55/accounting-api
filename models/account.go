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
	Asset AccountType = iota
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
	ParentID     *uint       `json:"parent_id" binding:"omitempty"`
	Parent       *Account
	CompanyID    uint
	Company      *Company
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
