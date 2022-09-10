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

	t.Run("List services", func(t *testing.T) {
		// Create for another company
		db.Create(&models.Company{Name: "Other company"})
		db.Create(&models.Service{Name: "Renting", RevenueAccountID: 1, CompanyID: 2})

		req := Get(t, "/services")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var services []*models.Service
		if err := json.Unmarshal(w.Body.Bytes(), &services); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(services) != 1 {
			t.Errorf("Expected %v services, got %v", 1, len(services))
		}

		for _, service := range services {
			if service.CompanyID != 1 {
				t.Errorf("Expected Company ID %v, got %v", 1, service.CompanyID)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := Get(t, "/services/1")

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
	})

	t.Run("Get non existent", func(t *testing.T) {
		req := Get(t, "/services/1221")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get invalid", func(t *testing.T) {
		req := Get(t, "/services/asonethu")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get from another company", func(t *testing.T) {
		req := Get(t, "/services/2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update", func(t *testing.T) {
		db.Create(&models.Account{
			Name:      "Cleaning revenue",
			Type:      models.Revenue,
			CompanyID: 1,
		})

		req := Put(t, "/services/1", map[string]interface{}{
			"name":               "General Cleaning",
			"revenue_account_id": 2,
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

		if service.Name != "General Cleaning" {
			t.Errorf("Expected name %v, got %v", "General Cleaning", service.Name)
		}

		if service.RevenueAccountID != 2 {
			t.Errorf("Expected account %v, got %v", 2, service.RevenueAccountID)
		}
	})

	t.Run("Update non existent", func(t *testing.T) {
		req := Put(t, "/services/4202", map[string]interface{}{
			"name":               "Renting",
			"revenue_account_id": 2,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update invalid", func(t *testing.T) {
		req := Put(t, "/services/sths", map[string]interface{}{
			"name":               "Renting",
			"revenue_account_id": 2,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update from another company", func(t *testing.T) {
		req := Put(t, "/services/2", map[string]interface{}{
			"name":               "Renting",
			"revenue_account_id": 2,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})
}
