package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

var (
	ErrReceivableAccountMissing = errors.New("Receivables account is required")
	ErrNotEnoughStock           = errors.New("Not enough stock")
)

func RegisterSalesEndpoints(router *gin.Engine) {
	group := router.Group("/sales")
	group.POST("", createSale)
	group.GET("", listSales)
	group.GET("/:id", viewSale)
	group.PUT("/:id", updateSale)
	group.DELETE("/:id", deleteSale)
}

func CreateAccountingEntries(data interface{}) {
	db, _ := database.GetConnection()

	sale := data.(*models.Sale)
	db.Joins("Company").First(&sale, sale.ID)

	for _, item := range sale.Items {
		var product *models.Product
		db.Joins("Company").Preload("StockEntries").First(&product, item.ProductID)

		costOfSale := product.Cost(item.Qty)

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

		sale.Entries = append(sale.Entries, &models.Entry{
			Description:  "Sale of product",
			CompanyID:    sale.CompanyID,
			Transactions: transactions,
		})
	}

	db.Save(&sale)
}

func ReduceProductStock(data interface{}) {
	sale := data.(*models.Sale)
	db, _ := database.GetConnection()

	for _, item := range sale.Items {
		var product *models.Product
		db.Joins("Company").Preload("StockEntries").First(&product, item.ProductID)

		usages := product.Consume(item.Qty)
		sale.StockUsages = append(sale.StockUsages, usages...)
	}

	db.Save(&sale)
}

func createSale(context *gin.Context) {
	var sale *models.Sale
	if err := context.ShouldBindJSON(&sale); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	for idx, item := range sale.Items {
		if item.Product == nil {
			db.Preload("StockEntries").First(&item.Product, item.ProductID)
		}

		if item.Product.Inventory() < item.Qty {
			context.JSON(http.StatusBadRequest, gin.H{
				fmt.Sprintf("Items.%d.Qty", idx): ErrNotEnoughStock.Error(),
			})
			return
		}
	}

	sale.CompanyID = context.Value("CompanyID").(uint)

	if db.Create(sale).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	events.Dispatch(events.SaleCreated, sale)

	db.Preload("Items.Product").Joins("PaymentAccount").Joins("ReceivableAccount").Joins("Customer").First(&sale)

	context.JSON(http.StatusOK, sale)
}

func listSales(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var sales []*models.Sale
	companyID := context.Value("CompanyID").(uint)

	query := db.Scopes(models.FromCompany(companyID))
	query = query.Joins("PaymentAccount").Joins("ReceivableAccount")

	if query.Preload("Items.Product").Joins("Customer").Find(&sales).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, sales)
}

func viewSale(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var sale *models.Sale
	companyID := context.Value("CompanyID").(uint)

	query := db.Scopes(models.FromCompany(companyID))
	query = query.Joins("PaymentAccount").Joins("ReceivableAccount")

	if query.Preload("Items.Product").Joins("Customer").First(&sale, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, sale)
}

func updateSale(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var sale *models.Sale
	companyID := context.Value("CompanyID").(uint)

	query := db.Scopes(models.FromCompany(companyID))
	query = query.Preload("Items.Product").Preload("Entries").Preload("StockUsages")

	if query.First(&sale, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	// Remove current accounting entries
	entryIDs := []uint{}
	for _, entry := range sale.Entries {
		entryIDs = append(entryIDs, entry.ID)
	}
	db.Unscoped().Delete(&sale.Entries, entryIDs)
	sale.Entries = []*models.Entry{}

	// Remove current stock usages
	usageIDs := []uint{}
	for _, usage := range sale.StockUsages {
		usageIDs = append(usageIDs, usage.ID)
	}
	db.Unscoped().Delete(&sale.StockUsages, usageIDs)
	sale.StockUsages = []*models.StockUsage{}

	if err := context.ShouldBindJSON(&sale); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	if db.Save(sale).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	events.Dispatch(events.SaleUpdated, sale)

	db.Preload("Items.Product").Joins("PaymentAccount").Joins("ReceivableAccount").Joins("Customer").First(&sale)

	context.JSON(http.StatusOK, sale)
}

func deleteSale(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var sale *models.Sale
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(models.FromCompany(companyID)).First(&sale, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if db.Unscoped().Select("StockUsages", "Entries").Delete(&sale).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}
