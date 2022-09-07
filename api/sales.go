package api

import (
	"math"
	"net/http"

	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterSalesEndpoints(router *gin.Engine) {
	group := router.Group("/sales")
	group.POST("", createSale)
}

func CreateAccountingEntries(data interface{}) {
	db, _ := database.GetConnection()

	sale := data.(*models.Sale)
	db.Joins("Company").Preload("Items").First(&sale, sale.ID)

	for _, item := range sale.Items {
		var product *models.Product
		db.Preload("StockEntries").First(&product, item.ProductID)

		costOfSale := 0.0
		left := item.Qty

		// Invert entries for LIFO
		if sale.Company.Stock == models.LIFO {
			for i, j := 0, len(product.StockEntries)-1; i < j; i, j = i+1, j-1 {
				product.StockEntries[i], product.StockEntries[j] = product.StockEntries[j], product.StockEntries[i]
			}
		}

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
				Value:     item.Subtotal(),
				AccountID: *product.RevenueAccountID,
			},
		}

		if sale.Paid {
			transactions = append(transactions, &models.Transaction{
				Value:     item.Subtotal(),
				AccountID: *sale.PaymentAccountID,
			})
		} else {
			transactions = append(transactions, &models.Transaction{
				Value:     item.Subtotal(),
				AccountID: *sale.ReceivableAccountID,
			})
		}

		db.Create(&models.Entry{
			Description:  "Sale of product",
			CompanyID:    sale.CompanyID,
			Transactions: transactions,
		})
	}
}

func ReduceProductStock(sale interface{}) {
	db, _ := database.GetConnection()

	for _, item := range sale.(*models.Sale).Items {
		var product *models.Product
		db.Joins("Company").Preload("StockEntries").First(&product, item.ProductID)

		// Invert entries for LIFO
		if product.Company.Stock == models.LIFO {
			for i, j := 0, len(product.StockEntries)-1; i < j; i, j = i+1, j-1 {
				product.StockEntries[i], product.StockEntries[j] = product.StockEntries[j], product.StockEntries[i]
			}
		}

		left := item.Qty

		for _, entry := range product.StockEntries {
			qty := entry.Qty

			if entry.Qty > left {
				entry.Qty -= uint(left)
				db.Save(&entry)
				break
			} else {
				db.Delete(&models.StockEntry{}, entry.ID)
			}
			left -= qty
		}
	}
}

func createSale(context *gin.Context) {
	var sale *models.Sale
	if err := context.ShouldBindJSON(&sale); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	sale.CompanyID = context.Value("CompanyID").(uint)

	if db.Create(sale).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	events.Dispatch(events.SaleCreated, sale)

	context.JSON(http.StatusOK, sale)
}
