package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"github.com/gin-gonic/gin"
)

var ErrEntryNotBalanced = errors.New("Entry transactions are not balanced")

func RegisterEntriesEndpoint(router *gin.Engine) {
	group := router.Group("/entries")

	group.POST("", createEntry)
	group.GET("", listEntries)
	group.GET("/:id", viewEntry)
	group.PUT("/:id", updateEntry)
	group.DELETE("/:id", deleteEntry)
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

func listEntries(context *gin.Context) {
	db, err := database.GetConnection()
	if err != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	var entries []models.Entry
	companyID := context.Value("CompanyID").(uint)

	if result := db.Scopes(database.FromCompany(companyID)).Find(&entries); result.Error != nil {
		log.Print("Could not find entries", result.Error)
	}

	context.JSON(http.StatusOK, entries)
}

func viewEntry(context *gin.Context) {
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

	var entry models.Entry
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).First(&entry, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	context.JSON(http.StatusOK, entry)
}

func updateEntry(context *gin.Context) {
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

	var entry *models.Entry
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).First(&entry, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if err := context.ShouldBindJSON(&entry); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !entry.IsBalanced() {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": ErrEntryNotBalanced.Error(),
		})
		return
	}

	if db.Save(entry).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.JSON(http.StatusOK, entry)
}

func deleteEntry(context *gin.Context) {
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

	var entry *models.Entry
	companyID := context.Value("CompanyID").(uint)

	if db.Scopes(database.FromCompany(companyID)).First(&entry, id).Error != nil {
		context.Status(http.StatusNotFound)
		return
	}

	if db.Delete(&models.Entry{}, id).Error != nil {
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusNoContent)
}
