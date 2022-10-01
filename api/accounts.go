package api

import (
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterAccountsEndpoints(router *gin.Engine) {
	accounts := router.Group("/accounts")

	accounts.GET("", listAccounts)
	accounts.GET("/:id", viewAccount)
	accounts.POST("", createAccount)
	accounts.PUT("/:id", updateAccount)
	accounts.DELETE("/:id", deleteAccount)
}

func listAccounts(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var items []*models.Account
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(models.FromCompany(companyID)).Preload("Transactions").Find(&items).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, items)
}

func viewAccount(context *gin.Context) {
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

	var account *models.Account
	companyID := context.Value("CompanyID").(uint)

	tx := db.Scopes(models.FromCompany(companyID))
	tx = tx.Joins("Parent").Preload("Children")

	if tx.First(&account, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, account)
}

func createAccount(context *gin.Context) {
	var account *models.Account
	if err := context.ShouldBindJSON(&account); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
		return
	}

	account.CompanyID = context.Value("CompanyID").(uint)

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	if result := db.Create(&account); result.Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	db.Joins("Parent").Preload("Children").First(&account)
	context.JSON(http.StatusOK, account)
}

func updateAccount(context *gin.Context) {
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

	var account *models.Account
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(models.FromCompany(companyID)).First(&account, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&account); err != nil {
		context.JSON(http.StatusBadRequest, Errors(err))
		return
	}

	if result := db.Save(account); result.Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	db.Joins("Parent").Preload("Children").First(&account)
	context.JSON(http.StatusOK, account)
}

func deleteAccount(context *gin.Context) {
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

	if db.Scopes(models.FromCompany(companyID)).First(&models.Account{}, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if result := db.Delete(&models.Account{}, id); result.Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}
