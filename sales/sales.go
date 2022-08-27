package sales

import (
	"errors"
	"math"

	"example.com/accounting/database"
	"example.com/accounting/entries"
	"example.com/accounting/events"
	"example.com/accounting/models"
	"example.com/accounting/products"
)

var (
	ErrItemsMissing             = errors.New("Items are required")
	ErrCustomerMissing          = errors.New("Customer is required")
	ErrNotEnoughStock           = errors.New("There is not enough in stock")
	ErrPaymentAccountMissing    = errors.New("Payment account is required")
	ErrReceivableAccountMissing = errors.New("Receivable account is required")
)

func CreateAccountingEntry(data interface{}) {
	sale := data.(*models.Sale)

	for _, item := range sale.Items {
		var product *models.Product
		products.Find(item.ProductID).With("StockEntries").First(&product)

		costOfSale := 0.0
		left := item.Qty

		// TODO: consider FIFO or LIFO
		for _, entry := range product.StockEntries {
			qty := math.Min(float64(left), float64(entry.Qty))
			costOfSale += entry.Price * qty
			left -= uint(qty)

			if left <= 0 {
				break
			}
		}

		transactions := []*models.Transaction{
			{
				Value:     -costOfSale,
				AccountID: product.InventoryAccountID,
			},
			{
				Value:     costOfSale,
				AccountID: *product.CostOfSaleAccountID,
			},
			{
				Value:     sale.Total(),
				AccountID: *product.RevenueAccountID,
			},
		}

		if sale.Paid {
			transactions = append(transactions, &models.Transaction{
				Value:     sale.Total(),
				AccountID: *sale.PaymentAccountID,
			})
		} else {
			transactions = append(transactions, &models.Transaction{
				Value:     sale.Total(),
				AccountID: *sale.ReceivableAccountID,
			})
		}

		entries.Create(sale.CompanyID, "Sale of product", transactions)
	}
}

func ReduceProductStock(sale interface{}) {
	db, _ := database.GetConnection()

	for _, item := range sale.(*models.Sale).Items {
		var product *models.Product
		products.Find(item.ProductID).With("StockEntries").First(&product)

		left := item.Qty
		// TODO: consider FIFO or LIFO
		for _, entry := range product.StockEntries {
			qty := entry.Qty
			if entry.Qty > left {
				entry.Qty -= uint(left)
				db.Update(&entry)
			} else {
				db.Delete(&models.StockEntry{}, entry.ID)
			}
			left -= qty
		}
	}
}

func Create(sale *models.Sale) error {
	if sale.Customer == nil {
		return ErrCustomerMissing
	}

	if len(sale.Items) == 0 {
		return ErrItemsMissing
	}

	if sale.Paid && sale.PaymentAccountID == nil {
		return ErrPaymentAccountMissing
	} else if !sale.Paid && sale.ReceivableAccountID == nil {
		return ErrReceivableAccountMissing
	}

	for _, item := range sale.Items {
		if item.Product == nil {
			products.Find(item.ProductID).With("StockEntries").First(&item.Product)
		}

		if item.Product.Inventory() < item.Qty {
			return ErrNotEnoughStock
		}
	}

	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	if err := db.Create(sale); err != nil {
		return err
	}

	events.Dispatch(events.SaleCreated, sale)

	return nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Sale{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Sale{}).Where("ID", id)
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Delete(&models.Sale{}, id)
}
