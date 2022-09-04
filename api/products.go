package api

import (
	"errors"
	"net/http"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

var (
	ErrRevenueAccountMissing    = errors.New("Revenue account is required")
	ErrCostOfSaleAccountMissing = errors.New("Cost of sale account is required")
)

func RegisterProductEndpoints(router *gin.Engine) {
	group := router.Group("/products")

	group.POST("", createProduct)
}

func createProduct(context *gin.Context) {
	var product *models.Product
	if err := context.ShouldBindJSON(&product); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := validateAccounts(product); err != nil {
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

	product.CompanyID = context.Value("CompanyID").(uint)

	if db.Create(&product).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, product)
}

func validateAccounts(product *models.Product) error {
	if product.Purchasable {
		if product.RevenueAccountID == nil {
			return ErrRevenueAccountMissing
		}

		if product.CostOfSaleAccountID == nil {
			return ErrCostOfSaleAccountMissing
		}
	}

	return nil
}
