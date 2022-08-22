package entries

import (
	"errors"

	"example.com/accounting/database"
	"example.com/accounting/models"
)

var ErrEntryNotBalanced = errors.New("Entry not balanced")

func Create(description string, transactions []*models.Transaction) (*models.Entry, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	entry := &models.Entry{
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
	return db.Find(&models.Entry{}), nil
}

func Find(id uint) (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&models.Entry{}).Where("ID", id), nil
}

func Update(entry *models.Entry) error {
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
	return db.Delete(&models.Entry{}, id)
}
