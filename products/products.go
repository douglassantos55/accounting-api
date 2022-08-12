package products

import (
	"errors"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
)

var ErrRevenueAccountMissing = errors.New("Renevue account is required")

type Product struct {
	database.Model
	Name      string
	Price     float64
	AccountID uint
	Account   *accounts.Account
}

func Create(name string, price float64, accountID uint) (*Product, error) {
	db, err := database.GetConnection()

	if err != nil {
		return nil, err
	}

	if !accountExists(accountID) {
		return nil, ErrRevenueAccountMissing
	}

	product := &Product{
		Name:      name,
		Price:     price,
		AccountID: accountID,
	}

	if err := db.Create(&product); err != nil {
		return nil, err
	}

	return product, nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Product{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Product{}).Where("ID", id)
}

func Update(product *Product) error {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}

	if !accountExists(product.AccountID) {
		return ErrRevenueAccountMissing
	}

	return db.Update(product)
}

func accountExists(accountID uint) bool {
	var account *accounts.Account
	if err := accounts.Find(accountID).First(&account); err != nil {
		return false
	}
	return true
}
