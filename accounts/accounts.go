package accounts

import (
	"example.com/accounting/database"
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

type Entry struct {
	database.Model
	Description  string
	Transactions []*Transaction `gorm:"constraint:OnDelete:CASCADE;"`
}

func (e Entry) IsBalanced() bool {
	totalDebit := 0.0
	totalCredit := 0.0
	for _, transaction := range e.Transactions {
		account := transaction.Account
		if account == nil {
			Find(transaction.AccountID).First(&account)
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
	database.Model
	Value     float64
	AccountID uint
	Account   *Account
	EntryID   uint
	Entry     *Entry
}

func Create(name string, accType AccountType, parentID *uint) (*Account, error) {
	db, err := database.GetConnection()

	if err != nil {
		return nil, err
	}

	account := &Account{
		Name:     name,
		Type:     accType,
		ParentID: parentID,
	}

	if err := db.Create(account); err != nil {
		return nil, err
	}

	return account, nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()

	if err != nil {
		return nil
	}

	return db.Find(&Account{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Account{}).Where("ID", id)
}

func Update(account *Account) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Update(account)
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	if err := Find(id).First(&Account{}); err != nil {
		return err
	}
	return db.Delete(&Account{}, id)
}
