package purchases

import (
	"example.com/accounting/database"
	"example.com/accounting/products"
)

type Purchase struct {
	database.Model
	Qty       uint
	ProductID uint
	Product   *products.Product
}

func Create(productId, qty uint) (*Purchase, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	purchase := &Purchase{
		ProductID: productId,
		Qty:       qty,
	}

	if err := db.Create(purchase); err != nil {
		return nil, err
	}

	var product *products.Product
	if err := products.Find(productId).First(&product); err != nil {
		return purchase, err
	}

	product.Stock += qty
	if err := products.Update(product); err != nil {
		return purchase, err
	}

	return purchase, nil
}

func List() (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&Purchase{}), nil
}

func Find(id uint) (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&Purchase{}).Where("ID", id), nil
}
