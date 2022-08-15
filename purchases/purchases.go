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

	if err := updateProductStock(productId, qty); err != nil {
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

func Update(purchase *Purchase) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	if err := db.Update(purchase); err != nil {
		return err
	}

	if err := updateProductStock(purchase.ProductID, purchase.Qty); err != nil {
		return err
	}

	return nil
}

func updateProductStock(productID, qty uint) error {
	var product *products.Product
	if err := products.Find(productID).First(&product); err != nil {
		return err
	}

	product.Stock += qty
	if err := products.Update(product); err != nil {
		return err
	}

	return nil
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	result, err := Find(id)
	if err != nil {
		return err
	}

	var purchase *Purchase
	if err := result.First(&purchase); err != nil {
		return err
	}

	if err := db.Delete(&Purchase{}, id); err != nil {
		return err
	}

	return updateProductStock(purchase.ProductID, -purchase.Qty)
}
