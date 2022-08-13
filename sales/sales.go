package sales

import (
	"errors"

	"example.com/accounting/customers"
	"example.com/accounting/database"
	"example.com/accounting/products"
)

var (
	ErrItemsMissing    = errors.New("Items are required")
	ErrCustomerMissing = errors.New("Customer is required")
)

type Sale struct {
	database.Model
	Items      []*Item
	Customer   *customers.Customer
	CustomerID uint
}

type Item struct {
	database.Model
	Qty       uint
	Price     float64
	ProductID uint
	Product   *products.Product
	SaleID    uint
	Sale      *Sale
}

func Create(customer *customers.Customer, items []*Item) (*Sale, error) {
	if customer == nil {
		return nil, ErrCustomerMissing
	}

	if len(items) == 0 {
		return nil, ErrItemsMissing
	}

	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	sale := &Sale{
		Customer: customer,
		Items:    items,
	}

	if err := db.Create(sale); err != nil {
		return nil, err
	}

	return sale, nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Sale{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Sale{}).Where("ID", id)
}
