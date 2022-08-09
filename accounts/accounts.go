package accounts

import (
	"example.com/accounting/database"
)

type AccountType uint

const (
	Asset AccountType = iota
	Liability
	Equity
	Expense
	Revenue
)

type Account struct {
	database.Model
	Name     string
	Type     AccountType
	ParentID int
	Parent   *Account
}

func Create(name string, accType AccountType, parent *Account) (*Account, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	account := &Account{
		Name:   name,
		Type:   accType,
		Parent: parent,
	}

	if err := db.Create(account); err != nil {
		return nil, err
	}

	return account, nil
}
