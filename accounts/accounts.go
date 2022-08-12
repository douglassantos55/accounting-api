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
	ParentID uint
	Parent   *Account
	Children []*Account `gorm:"foreignKey:ParentID"`
}

func Create(name string, accType AccountType, parentID uint) (*Account, error) {
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
