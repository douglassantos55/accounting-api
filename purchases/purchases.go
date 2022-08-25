package purchases

import (
	"errors"

	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/models"
	"example.com/accounting/products"
)

var (
	ErrPaymentAccountMissing = errors.New("Payment account is required")
	ErrPayableAccountMissing = errors.New("Payable account is required")
)

func CreateStockEntry(data interface{}) {
	db, _ := database.GetConnection()
	purchase := data.(*models.Purchase)

	purchase.StockEntry = &models.StockEntry{
		Price:     purchase.Price,
		Qty:       purchase.Qty,
		ProductID: purchase.ProductID,
	}

	db.Update(purchase)
}

func UpdateStockEntry(data interface{}) {
	db, _ := database.GetConnection()
	purchase := data.(*models.Purchase)

	if purchase.StockEntryID != nil {
		if purchase.StockEntry == nil {
			db.Find(&models.StockEntry{}).Where("ID", purchase.StockEntryID).First(&purchase.StockEntry)
		}

		purchase.StockEntry.Qty = purchase.Qty
		purchase.StockEntry.Price = purchase.Price
		purchase.StockEntry.ProductID = purchase.ProductID

		db.Update(purchase)
	}
}

func CreateAccountingEntry(data interface{}) {
	db, _ := database.GetConnection()
	purchase := data.(*models.Purchase)

	var product *models.Product
	products.Find(purchase.ProductID).First(&product)

	price := purchase.Price * float64(purchase.Qty)

	if purchase.Paid {
		purchase.PaymentEntry = &models.Entry{
			Description: "Purchase of product",
			Transactions: []*models.Transaction{
				{Value: price, AccountID: product.InventoryAccountID},
				{Value: -price, AccountID: *purchase.PaymentAccountID},
			},
		}
	} else {
		purchase.PayableEntry = &models.Entry{
			Description: "Purchase of product",
			Transactions: []*models.Transaction{
				{Value: price, AccountID: product.InventoryAccountID},
				{Value: price, AccountID: *purchase.PayableAccountID},
			},
		}
	}

	db.Update(purchase)
}

func UpdateAccountingEntry(data interface{}) {
	db, _ := database.GetConnection()
	purchase := data.(*models.Purchase)

	var product *models.Product
	products.Find(purchase.ProductID).First(&product)

	price := purchase.Price * float64(purchase.Qty)

	if purchase.PayableEntryID != nil {
		// update existing payable entry
		purchase.PayableEntry.Transactions[0].AccountID = product.InventoryAccountID
		purchase.PayableEntry.Transactions[0].Value = price

		purchase.PayableEntry.Transactions[1].AccountID = *purchase.PayableAccountID
		purchase.PayableEntry.Transactions[1].Value = price
	}

	if purchase.Paid {
		if purchase.PaymentEntryID != nil {
			// update existing payment entry
			purchase.PaymentEntry.Transactions[0].AccountID = product.InventoryAccountID
			purchase.PaymentEntry.Transactions[0].Value = price

			purchase.PaymentEntry.Transactions[1].AccountID = *purchase.PaymentAccountID
			purchase.PaymentEntry.Transactions[1].Value = -price
		} else {
			// create payment entry
			purchase.PaymentEntry = &models.Entry{
				Description: "Payment of purchase of product",
				Transactions: []*models.Transaction{
					{Value: -price, AccountID: *purchase.PayableAccountID},
					{Value: -price, AccountID: *purchase.PaymentAccountID},
				},
			}
		}
	} else {
		if purchase.PayableEntryID == nil {
			purchase.PayableEntry = &models.Entry{
				Description: "Purchase of product",
				Transactions: []*models.Transaction{
					{Value: price, AccountID: product.InventoryAccountID},
					{Value: price, AccountID: *purchase.PayableAccountID},
				},
			}
		}

		if purchase.PaymentEntryID != nil {
			db.Delete(&models.Entry{}, *purchase.PaymentEntryID)
			purchase.PaymentEntryID = nil
			purchase.PaymentEntry = nil
		}
	}

	db.Update(purchase)
}

func Create(purchase *models.Purchase) (*models.Purchase, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	if purchase.Paid && purchase.PaymentAccountID == nil {
		return nil, ErrPaymentAccountMissing
	}

	if !purchase.Paid && purchase.PayableAccountID == nil {
		return nil, ErrPayableAccountMissing
	}

	if err := db.Create(purchase); err != nil {
		return nil, err
	}

	events.Dispatch(events.PurchaseCreated, purchase)

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

	if purchase.Paid && purchase.PaymentAccountID == nil {
		return ErrPaymentAccountMissing
	}

	if !purchase.Paid && purchase.PayableAccountID == nil {
		return ErrPayableAccountMissing
	}

	if err := db.Update(purchase); err != nil {
		return err
	}

	events.Dispatch(events.PurchaseUpdated, purchase)

	return nil
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
