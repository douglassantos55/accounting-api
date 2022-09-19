package api

import (
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterVendorEndpoints(router *gin.Engine) {
	group := router.Group("/vendors")

	group.POST("", createVendor)
	group.GET("", listVendors)
	group.GET("/:id", viewVendor)
	group.PUT("/:id", updateVendor)
	group.DELETE("/:id", deleteVendor)
}

func createVendor(context *gin.Context) {
	var vendor *models.Vendor
	if err := context.ShouldBindJSON(&vendor); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
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

func listVendors(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	companyID := context.Value("CompanyID").(uint)

	var vendors []*models.Vendor
	if db.Scopes(models.FromCompany(companyID)).Find(&vendors).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, vendors)
}

func viewVendor(context *gin.Context) {
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

	var vendor *models.Vendor
	if db.Scopes(models.FromCompany(companyID)).First(&vendor, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, vendor)
}

func updateVendor(context *gin.Context) {
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

	var vendor *models.Vendor
	if db.Scopes(models.FromCompany(companyID)).First(&vendor, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&vendor); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
		return
	}

	if db.Save(&vendor).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, vendor)
}

func deleteVendor(context *gin.Context) {
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

	if db.Scopes(models.FromCompany(companyID)).First(&models.Vendor{}, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if db.Delete(&models.Vendor{}, id).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}
