package api

import (
	"errors"
	"net/http"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

var ErrInvalidCNPJ = errors.New("CPNJ invalido")

func RegisterVendorEndpoints(router *gin.Engine) {
	group := router.Group("/vendors")

	group.POST("", createVendor)
}

func createVendor(context *gin.Context) {
	var vendor *models.Vendor
	if err := context.ShouldBindJSON(&vendor); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !IsCNPJ(vendor.Cnpj) {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrInvalidCNPJ.Error(),
		})
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
