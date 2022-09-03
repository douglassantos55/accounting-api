package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/accounting/api"
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func TestVendors(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Vendor{})
	db.AutoMigrate(&models.Company{})

	if result := db.Create(&models.Company{Name: "Testing Company"}); result.Error != nil {
		t.Error(result.Error)
	}

	router := api.GetRouter()

	t.Run("Create", func(t *testing.T) {
		req := Post(t, "/vendors", map[string]interface{}{})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var vendor *models.Vendor
		if err := json.Unmarshal(w.Body.Bytes(), &vendor); err != nil {
			t.Error("Error parsing JSON", err)
		}

		if vendor.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, vendor.ID)
		}
	})
}
