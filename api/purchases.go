package api

import (
	"errors"
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

var (
	ErrPaymentAccountMissing = errors.New("Payment account is required")
	ErrPayableAccountMissing = errors.New("Payable account is required")
)

func RegisterPurchaseEndpoints(router *gin.Engine) {
	group := router.Group("/purchases")

	group.POST("", createPurchase)
	group.GET("", listPurchases)
	group.GET("/:id", viewPurchase)
	group.PUT("/:id", updatePurchase)
	group.DELETE("/:id", deletePurchase)
}

func CreateStockEntry(data interface{}) {
	db, _ := database.GetConnection()
	purchase := data.(*models.Purchase)

	purchase.StockEntry = &models.StockEntry{
		Price:     purchase.Price,
		Qty:       purchase.Qty,
		ProductID: purchase.ProductID,
	}

	db.Save(&purchase)
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

		db.Save(purchase)
	}
}

func CreateAccountingEntry(data interface{}) {
	db, _ := database.GetConnection()
	purchase := data.(*models.Purchase)

	var product *models.Product
	db.First(&product, purchase.ProductID)

	price := purchase.Price * float64(purchase.Qty)

	if purchase.Paid {
		purchase.PaymentEntry = &models.Entry{
			CompanyID:   purchase.CompanyID,
			Description: "Purchase of product",
			Transactions: []*models.Transaction{
				{Value: price, AccountID: product.InventoryAccountID},
				{Value: -price, AccountID: *purchase.PaymentAccountID},
			},
		}
	} else {
		purchase.PayableEntry = &models.Entry{
			CompanyID:   purchase.CompanyID,
			Description: "Purchase of product",
			Transactions: []*models.Transaction{
				{Value: price, AccountID: product.InventoryAccountID},
				{Value: price, AccountID: *purchase.PayableAccountID},
			},
		}
	}

	db.Save(purchase)
}

func UpdateAccountingEntry(data interface{}) {
	db, _ := database.GetConnection()
	purchase := data.(*models.Purchase)

	var product *models.Product
	db.First(&product, purchase.ProductID)

	price := purchase.Price * float64(purchase.Qty)

	if purchase.PayableEntry != nil {
		// update existing payable entry
		purchase.PayableEntry.Transactions[0].AccountID = product.InventoryAccountID
		purchase.PayableEntry.Transactions[0].Value = price

		purchase.PayableEntry.Transactions[1].AccountID = *purchase.PayableAccountID
		purchase.PayableEntry.Transactions[1].Value = price
	}

	if purchase.Paid {
		if purchase.PaymentEntry != nil {
			// update existing payment entry
			purchase.PaymentEntry.Transactions[0].AccountID = product.InventoryAccountID
			purchase.PaymentEntry.Transactions[0].Value = price

			purchase.PaymentEntry.Transactions[1].AccountID = *purchase.PaymentAccountID
			purchase.PaymentEntry.Transactions[1].Value = -price
		} else {
			// create payment entry
			purchase.PaymentEntry = &models.Entry{
				CompanyID:   purchase.CompanyID,
				Description: "Payment of purchase of product",
				Transactions: []*models.Transaction{
					{Value: -price, AccountID: *purchase.PayableAccountID},
					{Value: -price, AccountID: *purchase.PaymentAccountID},
				},
			}
		}
	} else {
		if purchase.PayableEntry == nil {
			purchase.PayableEntry = &models.Entry{
				CompanyID:   purchase.CompanyID,
				Description: "Purchase of product",
				Transactions: []*models.Transaction{
					{Value: price, AccountID: product.InventoryAccountID},
					{Value: price, AccountID: *purchase.PayableAccountID},
				},
			}
		}

		if purchase.PaymentEntry != nil {
			// Remove for real instead of soft-deleting so it cascades through
			// the transactions
			db.Unscoped().Delete(&models.Entry{}, purchase.PaymentEntry.ID)
			purchase.PaymentEntry = nil
		}
	}

	db.Save(purchase)
}

func createPurchase(context *gin.Context) {
	var purchase *models.Purchase
	if err := context.ShouldBindJSON(&purchase); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	purchase.CompanyID = context.Value("CompanyID").(uint)

	if result := db.Create(&purchase); result.Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	events.Dispatch(events.PurchaseCreated, purchase)

	db.
		Joins("PaymentAccount").
		Joins("PayableAccount").
		Preload("Product.Vendor").
		Preload("PaymentEntry.Transactions.Account").
		Preload("PayableEntry.Transactions.Account").
		First(&purchase)

	context.JSON(http.StatusOK, purchase)
}

func listPurchases(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var purchases []models.Purchase
	companyID := context.Value("CompanyID").(uint)

	tx := db.Scopes(models.FromCompany(companyID))
	tx = tx.Preload("PaymentEntry.Transactions.Account")
	tx = tx.Preload("PayableEntry.Transactions.Account")
	tx = tx.Preload("Product.Vendor").Joins("PaymentAccount").Joins("PayableAccount")

	if tx.Find(&purchases).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, purchases)
}

func viewPurchase(context *gin.Context) {
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

	var purchase models.Purchase
	companyID := context.Value("CompanyID").(uint)

	tx := db.Scopes(models.FromCompany(companyID))
	tx = tx.Preload("PaymentEntry.Transactions.Account")
	tx = tx.Preload("PayableEntry.Transactions.Account")
	tx = tx.Preload("Product.Vendor").Joins("PaymentAccount").Joins("PayableAccount")

	if tx.First(&purchase, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, purchase)
}

func updatePurchase(context *gin.Context) {
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

	var purchase *models.Purchase
	companyID := context.Value("CompanyID").(uint)

	query := db.Scopes(models.FromCompany(companyID))
	query = query.Preload("PaymentEntry.Transactions")
	query = query.Preload("PayableEntry.Transactions")

	if query.First(&purchase, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&purchase); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
		return
	}

	if db.Save(&purchase).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	events.Dispatch(events.PurchaseUpdated, purchase)

	db.
		Joins("PaymentAccount").
		Joins("PayableAccount").
		Preload("Product.Vendor").
		Preload("PaymentEntry.Transactions.Account").
		Preload("PayableEntry.Transactions.Account").
		First(&purchase)

	context.JSON(http.StatusOK, purchase)
}

func deletePurchase(context *gin.Context) {
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

	var purchase *models.Purchase
	companyID := context.Value("CompanyID").(uint)

	query := db.Scopes(models.FromCompany(companyID))
	query = query.Preload("PaymentEntry.Transactions")
	query = query.Preload("PayableEntry.Transactions")

	if query.First(&purchase, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if purchase.StockEntryID != nil {
		db.Unscoped().Delete(&models.StockEntry{}, *purchase.StockEntryID)
	}

	if purchase.PaymentEntry != nil {
		db.Unscoped().Delete(purchase.PaymentEntry)
		for _, transaction := range purchase.PaymentEntry.Transactions {
			db.Unscoped().Delete(&models.Transaction{}, transaction.ID)
		}
	}

	if purchase.PayableEntry != nil {
		db.Unscoped().Delete(purchase.PayableEntry)
		for _, transaction := range purchase.PayableEntry.Transactions {
			db.Unscoped().Delete(&models.Transaction{}, transaction.ID)
		}
	}

	if db.Unscoped().Delete(&models.Purchase{}, id).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}
