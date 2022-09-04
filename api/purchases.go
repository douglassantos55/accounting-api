package api

import (
	"errors"
	"fmt"
	"net/http"

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

func createPurchase(context *gin.Context) {
	var purchase *models.Purchase
	if err := context.ShouldBindJSON(&purchase); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if purchase.Paid && purchase.PaymentAccountID == nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrPaymentAccountMissing,
		})
		return
	}

	if !purchase.Paid && purchase.PayableAccountID == nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrPayableAccountMissing,
		})
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	purchase.CompanyID = context.Value("CompanyID").(uint)

	if result := db.Create(&purchase); result.Error != nil {
		fmt.Printf("result.Error.Error(): %v\n", result.Error.Error())
		context.Status(http.StatusInternalServerError)
		return
	}

	events.Dispatch(events.PurchaseCreated, purchase)

	context.JSON(http.StatusOK, purchase)
}
