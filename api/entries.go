package api

import (
	"errors"
	"net/http"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

var ErrEntryNotBalanced = errors.New("Entry transactions are not balanced")

func RegisterEntriesEndpoint(router *gin.Engine) {
	group := router.Group("/entries")

	group.POST("", createEntry)
}

func createEntry(context *gin.Context) {
	var entry *models.Entry
	if err := context.ShouldBindJSON(&entry); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	if !entry.IsBalanced() {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrEntryNotBalanced.Error(),
		})
		return
	}

	entry.CompanyID = context.Value("CompanyID").(uint)

	if db.Create(&entry).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, entry)
}
