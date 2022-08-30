package api

import (
	"fmt"
	"net/http"
	"strconv"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

func RegisterAccountsEndpoints() {
	router := GetRouter()
	accounts := router.Group("/accounts")

	accounts.GET("", list)
	accounts.GET("/:id", view)
	accounts.POST("", create)
	accounts.PUT("/:id", update)
	accounts.DELETE("/:id", remove)
}

func list(context *gin.Context) {
	var items []models.Account

	if err := accounts.List().Get(&items); err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, items)
}

func view(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 64)

	if err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	var account models.Account
	if err := accounts.Find(uint(id)).First(&account); err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, account)
}

func create(context *gin.Context) {
	var account *models.Account
	if err := context.ShouldBindJSON(&account); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.CompanyID = context.Value("CompanyID").(uint)

	db, _ := database.GetConnection()
	if err := db.Create(&account); err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(200, account)
}

func update(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 64)

	if err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	var account *models.Account
	if err := accounts.Find(uint(id)).First(&account); err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&account); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}

	if err := accounts.Update(account); err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, account)
}

func remove(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 64)

	if err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := accounts.Delete(uint(id)); err != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.Status(http.StatusNoContent)
}
