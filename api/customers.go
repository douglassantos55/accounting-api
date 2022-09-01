package api

import (
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterCustomerEndpoints(router *gin.Engine) {
	group := router.Group("/customers")

	group.POST("", createCustomer)
	group.GET("", listCustomers)
	group.GET("/:id", viewCustomer)
	group.PUT("/:id", updateCustomer)
	group.DELETE("/:id", deleteCustomer)
}

func createCustomer(context *gin.Context) {
	var customer *models.Customer
	if err := context.ShouldBindJSON(&customer); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ignore everything if there's no postcode
	if customer.Address.Postcode == "" {
		customer.Address = nil
	}

	customer.CompanyID = context.Value("CompanyID").(uint)

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	if result := db.Create(customer); result.Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, customer)
}

func listCustomers(context *gin.Context) {

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var items []*models.Customer
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).Find(&items).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, items)
}

func viewCustomer(context *gin.Context) {
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

	companyID := context.Value("CompanyID").(uint)
	var customer *models.Customer

	if db.Scopes(database.FromCompany(companyID)).First(&customer, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, customer)
}

func updateCustomer(context *gin.Context) {
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

	companyID := context.Value("CompanyID").(uint)
	var customer *models.Customer

	if db.Scopes(database.FromCompany(companyID)).First(&customer, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&customer); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if db.Save(&customer).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, customer)
}

func deleteCustomer(context *gin.Context) {
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

	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).First(&models.Customer{}, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if result := db.Scopes(database.FromCompany(companyID)).Delete(&models.Customer{}, id); result.Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}
