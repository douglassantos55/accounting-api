package api

import (
	"errors"
	"net/http"
	"strconv"

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
	group.GET("", listProducts)
	group.GET("/:id", viewProduct)
	group.PUT("/:id", updateProduct)
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

func listProducts(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var products []*models.Product
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).Find(&products).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, products)
}

func viewProduct(context *gin.Context) {
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

	var product *models.Product
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).First(&product, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, product)
}

func updateProduct(context *gin.Context) {
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

	var product models.Product
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).First(&product, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&product); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := validateAccounts(&product); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if db.Save(&product).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, product)
}
