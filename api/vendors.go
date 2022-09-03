package api

import (
	"errors"
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

var ErrInvalidCNPJ = errors.New("CPNJ invalido")

func RegisterVendorEndpoints(router *gin.Engine) {
	group := router.Group("/vendors")

	group.POST("", createVendor)
	group.GET("", listVendors)
	group.GET("/:id", viewVendor)
	group.PUT("/:id", updateVendor)
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

func listVendors(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	companyID := context.Value("CompanyID").(uint)

	var vendors []*models.Vendor
	if db.Scopes(database.FromCompany(companyID)).Find(&vendors).Error != nil {
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
	if db.Scopes(database.FromCompany(companyID)).First(&vendor, id).Error != nil {
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
	if db.Scopes(database.FromCompany(companyID)).First(&vendor, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

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

	if db.Save(&vendor).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, vendor)
}
