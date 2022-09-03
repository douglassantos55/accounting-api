package api

import (
	"net/http"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterVendorEndpoints(router *gin.Engine) {
	group := router.Group("/vendors")

	group.POST("", createVendor)
}

func createVendor(context *gin.Context) {
	var vendor *models.Vendor
	if err := context.ShouldBindJSON(&vendor); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	vendor.CompanyID = context.Value("CompanyID").(uint)

	if db.Create(&vendor).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, vendor)
}
