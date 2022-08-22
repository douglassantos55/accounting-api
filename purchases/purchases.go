package purchases

import (
	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/entries"
	"example.com/accounting/products"
)

func Create(productId, qty uint, price float64, paymentAccountID uint) (*products.Purchase, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	purchase := &products.Purchase{
		Qty:              qty,
		Price:            price,
		ProductID:        productId,
		PaymentAccountID: &paymentAccountID,
		StockEntry: &products.StockEntry{
			Price:     price,
			Qty:       qty,
			ProductID: productId,
		},
	}

	if err := db.Transaction(func() error {
		if err := db.Create(purchase); err != nil {
			return err
		}

		var product *products.Product
		if err := products.Find(productId).First(&product); err != nil {
			return err
		}

		if _, err := entries.Create("Purchase of product", []*accounts.Transaction{
			{Value: price * float64(qty), AccountID: product.InventoryAccountID},
			{Value: -price * float64(qty), AccountID: paymentAccountID},
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return purchase, nil
}

func List() (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&products.Purchase{}), nil
}

func Find(id uint) (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&products.Purchase{}).Where("ID", id), nil
}

func Update(purchase *products.Purchase) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	return db.Transaction(func() error {
		if purchase.StockEntryID != nil {
			if purchase.StockEntry == nil {
				db.Find(&products.StockEntry{}).Where("ID", purchase.StockEntryID).First(&purchase.StockEntry)
			}

			purchase.StockEntry.Qty = purchase.Qty
			purchase.StockEntry.Price = purchase.Price
			purchase.StockEntry.ProductID = purchase.ProductID
		}

		if err := db.Update(purchase); err != nil {
			return err
		}

		return nil
	})
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	return db.Transaction(func() error {
		result, err := Find(id)
		if err != nil {
			return err
		}

		var purchase *products.Purchase
		if err := result.First(&purchase); err != nil {
			return err
		}

		if err := db.Delete(&products.Purchase{}, id); err != nil {
			return err
		}

		if err := db.Delete(&products.StockEntry{}, *purchase.StockEntryID); err != nil {
			return err
		}

		return nil
	})
}
