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

	group.POST("/performed", createPerformed)
	group.PUT("/performed/:id", updatePerformed)
	group.DELETE("/performed/:id", deletePerformed)
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

func createPerformed(context *gin.Context) {
	var performed *models.ServicePerformed
	if err := context.ShouldBindJSON(&performed); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	if performed.Paid && performed.PaymentAccountID == nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrPaymentAccountMissing.Error(),
		})
		return
	}

	if !performed.Paid && performed.ReceivableAccountID == nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrReceivableAccountMissing.Error(),
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

	createServiceEntries(performed)
	reduceConsumptionStock(performed)

	context.JSON(http.StatusOK, performed)
}

func updatePerformed(context *gin.Context) {
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

	var performed *models.ServicePerformed
	companyID := context.Value("CompanyID").(uint)

	tx := db.Scopes(models.FromCompany(companyID))
	if tx.Preload("Entries").Preload("StockUsages").First(&performed, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&performed); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	if performed.Paid && performed.PaymentAccountID == nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrPaymentAccountMissing.Error(),
		})
		return
	}

	if !performed.Paid && performed.ReceivableAccountID == nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrReceivableAccountMissing.Error(),
		})
		return
	}

	// Remove current accounting entries
	entryIDs := []uint{}
	for _, entry := range performed.Entries {
		entryIDs = append(entryIDs, entry.ID)
	}
	db.Unscoped().Delete(&performed.Entries, entryIDs)
	performed.Entries = []*models.Entry{}

	// Remove current stock usages
	usageIDs := []uint{}
	for _, usage := range performed.StockUsages {
		usageIDs = append(usageIDs, usage.ID)
	}
	db.Unscoped().Delete(&performed.StockUsages, usageIDs)
	performed.StockUsages = []*models.StockUsage{}

	if db.Save(performed).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	createServiceEntries(performed)
	reduceConsumptionStock(performed)

	context.JSON(http.StatusOK, performed)
}

func deletePerformed(context *gin.Context) {
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

	var performed *models.ServicePerformed
	companyID := context.Value("CompanyID").(uint)

	tx := db.Scopes(models.FromCompany(companyID))
	if tx.First(&performed, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	tx = db.Unscoped().Select("Entries", "StockUsages", "Consumptions")
	if tx.Delete(&performed).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}

func createServiceEntries(performed *models.ServicePerformed) {
	db, _ := database.GetConnection()

	var service *models.Service
	db.First(&service, performed.ServiceID)

	var account uint
	if performed.Paid {
		account = *performed.PaymentAccountID
	} else {
		account = *performed.ReceivableAccountID
	}

	performed.Entries = append(performed.Entries, &models.Entry{
		Description: "Service performed",
		CompanyID:   performed.CompanyID,
		Transactions: []*models.Transaction{
			{AccountID: service.RevenueAccountID, Value: performed.Value},
			{AccountID: account, Value: performed.Value},
		},
	})

	for _, consumption := range performed.Consumptions {
		var prod *models.Product
		db.Joins("Company").Preload("StockEntries").First(&prod, consumption.ProductID)

		cost := prod.Cost(consumption.Qty)

		performed.Entries = append(performed.Entries, &models.Entry{
			Description: "Usage for service",
			CompanyID:   performed.CompanyID,
			Transactions: []*models.Transaction{
				{AccountID: prod.InventoryAccountID, Value: -cost},
				{AccountID: service.CostOfServiceAccountID, Value: cost},
			},
		})
	}

	db.Save(performed)
}

func reduceConsumptionStock(performed *models.ServicePerformed) {
	db, _ := database.GetConnection()

	for _, item := range performed.Consumptions {
		var product *models.Product
		db.Joins("Company").Preload("StockEntries").First(&product, item.ProductID)

		usages := product.Consume(item.Qty)
		performed.StockUsages = append(performed.StockUsages, usages...)
	}

	db.Save(&performed)
}
