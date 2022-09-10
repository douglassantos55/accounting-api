package api

import (
	"net/http"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterServicesEndpoints(router *gin.Engine) {
	group := router.Group("/services")
	group.POST("", createService)
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
