package entries

import (
	"errors"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
)

var ErrEntryNotBalanced = errors.New("Entry not balanced")

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
			accounts.Find(transaction.AccountID).First(&account)
		}

		if account.TransactionType() == accounts.Debit {
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
	Account   *accounts.Account
	EntryID   uint
	Entry     *Entry
}

func Create(description string, transactions []*Transaction) (*Entry, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	entry := &Entry{
		Description:  description,
		Transactions: transactions,
	}

	if !entry.IsBalanced() {
		return nil, ErrEntryNotBalanced
	}

	if err := db.Create(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func List() (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&Entry{}), nil
}

func Find(id uint) (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&Entry{}).Where("ID", id), nil
}

func Update(entry *Entry) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Update(entry)
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Delete(&Entry{}, id)
}