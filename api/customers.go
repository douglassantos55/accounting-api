package api

import (
	"net/http"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterCustomerEndpoints(router *gin.Engine) {
	group := router.Group("/customers")

	group.POST("", createCustomer)
	group.GET("", listCustomers)
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
