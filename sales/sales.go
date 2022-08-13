package sales

import (
	"example.com/accounting/customers"
	"example.com/accounting/database"
	"example.com/accounting/products"
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
