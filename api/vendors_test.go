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
		req := Post(t, "/vendors", map[string]interface{}{
			"name": "Vendor 1",
			"cnpj": "87.381.309/0001-57",
			"address": map[string]interface{}{
				"street":       "Street",
				"number":       "211",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "2222222",
			},
		})

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

		if vendor.Name != "Vendor 1" {
			t.Errorf("Expected name %v, got %v", "Vendor 1", vendor.Name)
		}
	})

	t.Run("Validate CNPJ", func(t *testing.T) {
		req := Post(t, "/vendors", map[string]interface{}{
			"name": "Vendor 1",
			"cnpj": "87.381.309/0001-50",
			"address": map[string]interface{}{
				"street":       "Street",
				"number":       "211",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "2222222",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("Could not parse error", err)
		}

		if response["error"] != api.ErrInvalidCNPJ.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrInvalidCNPJ, response["error"])
		}
	})

	t.Run("Address not required", func(t *testing.T) {
		req := Post(t, "/vendors", map[string]interface{}{
			"name": "Vendor 2",
			"cnpj": "33.441.041/0001-72",
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var vendor *models.Vendor
		if err := json.Unmarshal(w.Body.Bytes(), &vendor); err != nil {
			t.Error("Error parsing JSON", err)
		}

		if vendor.ID != 2 {
			t.Errorf("Expected ID %v, got %v", 2, vendor.ID)
		}

		if vendor.Name != "Vendor 2" {
			t.Errorf("Expected name %v, got %v", "Vendor 2", vendor.Name)
		}

		if vendor.Address != nil {
			t.Error("Should not have an address")
		}
	})

	t.Run("List", func(t *testing.T) {
		db, _ := database.GetConnection()

		// Create one for another company
		db.Create(&models.Company{Name: "Other Company"})
		db.Create(&models.Vendor{Name: "Vendor 3", CompanyID: 2})

		req := Get(t, "/vendors")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var vendors []*models.Vendor
		if err := json.Unmarshal(w.Body.Bytes(), &vendors); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(vendors) != 2 {
			t.Errorf("Expected %v vendors, got %v", 2, len(vendors))
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := Get(t, "/vendors/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var vendor *models.Vendor
		if err := json.Unmarshal(w.Body.Bytes(), &vendor); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if vendor == nil {
			t.Error("Should return vendor")
		}

		if vendor.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, vendor.ID)
		}
	})

	t.Run("Get non existent", func(t *testing.T) {
		req := Get(t, "/vendors/1312")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}

		req = Get(t, "/vendors/aosetnh")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get from another company", func(t *testing.T) {
		req := Get(t, "/vendors/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update", func(t *testing.T) {
		req := Put(t, "/vendors/1", map[string]interface{}{
			"name": "Updated vendor",
			"cnpj": "80.529.809/0001-44",
			"address": map[string]interface{}{
				"street": "Another street",
				"city":   "San Francisco",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var vendor *models.Vendor
		if err := json.Unmarshal(w.Body.Bytes(), &vendor); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if vendor.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, vendor.ID)
		}

		if vendor.Name != "Updated vendor" {
			t.Errorf("Expected name %v, got %v", "Updated vendor", vendor.Name)
		}

		if vendor.Cnpj != "80.529.809/0001-44" {
			t.Errorf("Expected CNPJ %v, got %v", "80.529.809/0001-44", vendor.Cnpj)
		}

		if vendor.Address.Street != "Another street" {
			t.Errorf("Expected street %v, got %v", "Another street", vendor.Address.Street)
		}

		if vendor.Address.City != "San Francisco" {
			t.Errorf("Expected city %v, got %v", "San Francisco", vendor.Address.City)
		}
	})

	t.Run("Update validate CNPJ", func(t *testing.T) {
		req := Put(t, "/vendors/2", map[string]interface{}{
			"name": "Updated vendor",
			"cnpj": "80.529.809/0001-04",
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("Could not parse error", err)
		}

		if response["error"] != api.ErrInvalidCNPJ.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrInvalidCNPJ, response["error"])
		}
	})

	t.Run("Update from another company", func(t *testing.T) {
		req := Put(t, "/vendors/3", map[string]interface{}{
			"name": "Updated vendor",
			"cnpj": "80.529.809/0001-44",
			"address": map[string]interface{}{
				"street": "Another street",
				"city":   "San Francisco",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})
}
