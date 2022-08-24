package purchases

import (
	"errors"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"example.com/accounting/products"
)

var (
	ErrPaymentAccountMissing = errors.New("Payment account is required")
	ErrPayableAccountMissing = errors.New("Payable account is required")
)

func Create(purchase *models.Purchase) (*models.Purchase, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	if err := db.Transaction(func() error {
		var product *models.Product
		if err := products.Find(purchase.ProductID).First(&product); err != nil {
			return err
		}

		purchase.StockEntry = &models.StockEntry{
			Price:     purchase.Price,
			Qty:       purchase.Qty,
			ProductID: purchase.ProductID,
		}

		if purchase.Paid {
			if purchase.PaymentAccountID == nil {
				return ErrPaymentAccountMissing
			}

			purchase.PaymentEntry = &models.Entry{
				Description: "Purchase of product",
				Transactions: []*models.Transaction{
					{Value: purchase.Price * float64(purchase.Qty), AccountID: product.InventoryAccountID},
					{Value: -purchase.Price * float64(purchase.Qty), AccountID: *purchase.PaymentAccountID},
				},
			}
		} else {
			if purchase.PayableAccountID == nil {
				return ErrPayableAccountMissing
			}

			purchase.PayableEntry = &models.Entry{
				Description: "Purchase of product",
				Transactions: []*models.Transaction{
					{Value: purchase.Price * float64(purchase.Qty), AccountID: product.InventoryAccountID},
					{Value: purchase.Price * float64(purchase.Qty), AccountID: *purchase.PayableAccountID},
				},
			}
		}

		return db.Create(purchase)
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
		var product *models.Product
		if err := products.Find(purchase.ProductID).First(&product); err != nil {
			return err
		}

		if purchase.StockEntryID != nil {
			if purchase.StockEntry == nil {
				db.Find(&models.StockEntry{}).Where("ID", purchase.StockEntryID).First(&purchase.StockEntry)
			}

			purchase.StockEntry.Qty = purchase.Qty
			purchase.StockEntry.Price = purchase.Price
			purchase.StockEntry.ProductID = purchase.ProductID
		}

		if purchase.Paid {
			if purchase.PaymentAccountID == nil {
				return ErrPaymentAccountMissing
			}

			if purchase.PaymentEntryID != nil {
				purchase.PaymentEntry.Transactions[0].AccountID = product.InventoryAccountID
				purchase.PaymentEntry.Transactions[0].Value = purchase.Price * float64(purchase.Qty)

				purchase.PaymentEntry.Transactions[1].AccountID = *purchase.PaymentAccountID
				purchase.PaymentEntry.Transactions[1].Value = -purchase.Price * float64(purchase.Qty)
			} else if purchase.PayableEntryID != nil {
				if purchase.PayableAccountID == nil {
					return ErrPayableAccountMissing
				}

				// create payment entry
				purchase.PaymentEntry = &models.Entry{
					Description: "Payment of purchase of product",
					Transactions: []*models.Transaction{
						{Value: -purchase.Price * float64(purchase.Qty), AccountID: *purchase.PayableAccountID},
						{Value: -purchase.Price * float64(purchase.Qty), AccountID: *purchase.PaymentAccountID},
					},
				}
			}

			if purchase.PayableEntryID != nil {
				purchase.PayableEntry.Transactions[0].AccountID = product.InventoryAccountID
				purchase.PayableEntry.Transactions[0].Value = purchase.Price * float64(purchase.Qty)

				purchase.PayableEntry.Transactions[1].AccountID = *purchase.PayableAccountID
				purchase.PayableEntry.Transactions[1].Value = purchase.Price * float64(purchase.Qty)
			}
		} else {
			if purchase.PayableAccountID == nil {
				return ErrPayableAccountMissing
			}

			if purchase.PayableEntryID == nil {
				purchase.PayableEntry = &models.Entry{
					Description: "Purchase of product",
					Transactions: []*models.Transaction{
						{Value: purchase.Price * float64(purchase.Qty), AccountID: product.InventoryAccountID},
						{Value: purchase.Price * float64(purchase.Qty), AccountID: *purchase.PayableAccountID},
					},
				}
			} else {
				purchase.PayableEntry.Transactions[0].AccountID = product.InventoryAccountID
				purchase.PayableEntry.Transactions[0].Value = purchase.Price * float64(purchase.Qty)

				purchase.PayableEntry.Transactions[1].AccountID = *purchase.PayableAccountID
				purchase.PayableEntry.Transactions[1].Value = purchase.Price * float64(purchase.Qty)
			}

			if purchase.PaymentEntryID != nil {
				db.Delete(&models.Entry{}, *purchase.PaymentEntryID)
				purchase.PaymentEntryID = nil
				purchase.PaymentEntry = nil
			}
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
