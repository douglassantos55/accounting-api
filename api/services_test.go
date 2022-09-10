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

func TestServices(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	t.Cleanup(database.Cleanup)

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Company{})
	db.AutoMigrate(&models.Service{})
	db.AutoMigrate(&models.Account{})

	db.Create(&models.Company{Name: "Testing Company"})
	db.Create(&models.Account{Name: "Revenue", Type: models.Revenue, CompanyID: 1})

	router := api.GetRouter()

	t.Run("Create", func(t *testing.T) {
		req := Post(t, "/services", map[string]interface{}{
			"name":               "Window cleaning",
			"revenue_account_id": 1,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var service *models.Service
		if err := json.Unmarshal(w.Body.Bytes(), &service); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if service.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, service.ID)
		}

		if service.RevenueAccountID != 1 {
			t.Errorf("EXpected revenue account %v, got %v", 1, service.RevenueAccountID)
		}

		if service.Name != "Window cleaning" {
			t.Errorf("Expected name %v, got %v", "Window cleaning", service.Name)
		}
	})

	t.Run("Validation", func(t *testing.T) {
		req := Post(t, "/services", map[string]interface{}{
			"name":               "",
			"revenue_account_id": "",
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})
}
