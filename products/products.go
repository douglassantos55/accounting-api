package products

import (
	"example.com/accounting/database"
)

type Product struct {
	database.Model
	Name      string
	Price     float64
	AccountID uint
}

func Create(name string, price float64, accountID uint) (*Product, error) {
	db, err := database.GetConnection()

	if err != nil {
		return nil, err
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
