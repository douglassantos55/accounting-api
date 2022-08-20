package sales

import (
	"errors"

	"example.com/accounting/customers"
	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/products"
)

var (
	ErrItemsMissing    = errors.New("Items are required")
	ErrCustomerMissing = errors.New("Customer is required")
	ErrNotEnoughStock  = errors.New("There is not enough in stock")
)

type Sale struct {
	database.Model
	Items    []*Item `gorm:"constraint:OnDelete:CASCADE;"`
	Customer *customers.Customer

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

func ReduceProductStock(sale interface{}) {
	db, _ := database.GetConnection()

	for _, item := range sale.(*Sale).Items {
		var product *products.Product
		products.Find(item.ProductID).With("StockEntries").First(&product)

		left := item.Qty
		// TODO: consider FIFO or LIFO
		for _, entry := range product.StockEntries {
			qty := entry.Qty
			if entry.Qty > left {
				entry.Qty -= uint(left)
				db.Update(&entry)
			} else {
				db.Delete(&products.StockEntry{}, entry.ID)
			}
			left -= qty
		}
	}
}

func Create(customer *customers.Customer, items []*Item) (*Sale, error) {
	if customer == nil {
		return nil, ErrCustomerMissing
	}

	if len(items) == 0 {
		return nil, ErrItemsMissing
	}

	for _, item := range items {
		if item.Product == nil {
			products.Find(item.ProductID).With("StockEntries").First(&item.Product)
		}

		if item.Product.Inventory() < item.Qty {
			return nil, ErrNotEnoughStock
		}
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

	events.Dispatch(events.SaleCreated, sale)

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

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Delete(&Sale{}, id)
}
