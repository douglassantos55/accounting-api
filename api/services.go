package api

import (
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterServicesEndpoints(router *gin.Engine) {
	group := router.Group("/services")
	group.POST("", createService)
	group.GET("", listServices)
	group.GET("/:id", viewService)
	group.PUT("/:id", updateService)
	group.DELETE("/:id", deleteService)
	group.POST("/perform", performService)
}

func createService(context *gin.Context) {
	var service *models.Service
	if err := context.ShouldBindJSON(&service); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	service.CompanyID = context.Value("CompanyID").(uint)

	if db.Create(&service).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, service)
}

func listServices(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var services []*models.Service
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(models.FromCompany(companyID)).Find(&services).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, services)
}

func viewService(context *gin.Context) {
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

	var service *models.Service
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(models.FromCompany(companyID)).First(&service, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, service)
}

func updateService(context *gin.Context) {
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

	var service *models.Service
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(models.FromCompany(companyID)).First(&service, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&service); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	if db.Save(&service).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, service)
}

func deleteService(context *gin.Context) {
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

	if db.Scopes(models.FromCompany(companyID)).First(&models.Service{}, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if db.Delete(&models.Service{}, id).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}

func performService(context *gin.Context) {
	var performed *models.ServicePerformed
	if err := context.ShouldBindJSON(&performed); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	performed.CompanyID = context.Value("CompanyID").(uint)

	if db.Create(&performed).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	createEntries(performed)

	context.JSON(http.StatusOK, performed)
}

func createEntries(performed *models.ServicePerformed) {
	db, _ := database.GetConnection()

	var service *models.Service
	db.First(&service, performed.ServiceID)

	db.Create(&models.Entry{
		Description: "Service performed",
		CompanyID:   performed.CompanyID,
		Transactions: []*models.Transaction{
			{AccountID: service.RevenueAccountID, Value: performed.Value},
			{AccountID: *performed.PaymentAccountID, Value: performed.Value},
		},
	})

	for _, consumption := range performed.Consumptions {
		var prod *models.Product
		db.Joins("Company").Preload("StockEntries").First(&prod, consumption.ProductID)

		cost := prod.Cost(consumption.Qty)

		db.Create(&models.Entry{
			Description: "Usage for service",
			CompanyID:   performed.CompanyID,
			Transactions: []*models.Transaction{
				{AccountID: prod.InventoryAccountID, Value: -cost},
				{AccountID: service.CostOfServiceAccountID, Value: cost},
			},
		})
	}
}
