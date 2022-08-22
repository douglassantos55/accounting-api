package purchases

import (
	"example.com/accounting/database"
	"example.com/accounting/models"
	"example.com/accounting/products"
)

func Create(productId, qty uint, price float64, paymentAccountID uint) (*models.Purchase, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	var purchase *models.Purchase

	if err := db.Transaction(func() error {
		var product *models.Product
		if err := products.Find(productId).First(&product); err != nil {
			return err
		}

		purchase = &models.Purchase{
			Qty:              qty,
			Price:            price,
			ProductID:        productId,
			PaymentAccountID: &paymentAccountID,
			StockEntry: &models.StockEntry{
				Price:     price,
				Qty:       qty,
				ProductID: productId,
			},
			Entries: []*models.Entry{{
				Description: "Purchase of product",
				Transactions: []*models.Transaction{
					{Value: price * float64(qty), AccountID: product.InventoryAccountID},
					{Value: -price * float64(qty), AccountID: paymentAccountID},
				},
			}},
		}

		if err := db.Create(purchase); err != nil {
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
	return db.Find(&models.Purchase{}), nil
}

func Find(id uint) (database.QueryResult, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Find(&models.Purchase{}).Where("ID", id), nil
}

func Update(purchase *models.Purchase) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	return db.Transaction(func() error {
		if purchase.StockEntryID != nil {
			if purchase.StockEntry == nil {
				db.Find(&models.StockEntry{}).Where("ID", purchase.StockEntryID).First(&purchase.StockEntry)
			}

			purchase.StockEntry.Qty = purchase.Qty
			purchase.StockEntry.Price = purchase.Price
			purchase.StockEntry.ProductID = purchase.ProductID
		}

		purchase.Entries[0].Transactions[1].AccountID = *purchase.PaymentAccountID
		purchase.Entries[0].Transactions[1].Value = -purchase.Price * float64(purchase.Qty)

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

		var purchase *models.Purchase
		if err := result.First(&purchase); err != nil {
			return err
		}

		if err := db.Delete(&models.Purchase{}, id); err != nil {
			return err
		}

		if err := db.Delete(&models.StockEntry{}, *purchase.StockEntryID); err != nil {
			return err
		}

		return nil
	})
}
