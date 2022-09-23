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
	group.DELETE("/:id", deleteProduct)
}

func createProduct(context *gin.Context) {
	var product *models.Product
	if err := context.ShouldBindJSON(&product); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
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

	tx := db.Joins("InventoryAccount").Joins("Vendor")
	tx = tx.Joins("RevenueAccount").Joins("CostOfSaleAccount").First(&product)

	context.JSON(http.StatusOK, product)
}

func listProducts(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var products []*models.Product
	companyID := context.Value("CompanyID").(uint)

	tx := db.Scopes(models.FromCompany(companyID))
	tx = tx.Joins("InventoryAccount").Joins("Vendor")
	tx = tx.Joins("RevenueAccount").Joins("CostOfSaleAccount")

	if tx.Find(&products).Error != nil {
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

	tx := db.Scopes(models.FromCompany(companyID))
	tx = db.Joins("InventoryAccount").Joins("Vendor")
	tx = tx.Joins("RevenueAccount").Joins("CostOfSaleAccount")

	if tx.First(&product, id).Error != nil {
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

	if db.Scopes(models.FromCompany(companyID)).First(&product, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&product); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
		return
	}

	if db.Save(&product).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	tx := db.Joins("InventoryAccount").Joins("Vendor")
	tx = tx.Joins("RevenueAccount").Joins("CostOfSaleAccount").First(&product)

	context.JSON(http.StatusOK, product)
}

func deleteProduct(context *gin.Context) {
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

	if db.Scopes(models.FromCompany(companyID)).First(&product, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if db.Scopes(models.FromCompany(companyID)).Delete(&models.Product{}, id).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}
