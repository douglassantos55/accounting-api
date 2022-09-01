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

func TestCustomersEndpoint(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Company{})
	db.AutoMigrate(&models.Customer{})

	if result := db.Create(&models.Company{Name: "Testing Company"}); result.Error != nil {
		t.Error(result.Error)
	}

	router := api.GetRouter()

	t.Run("Create", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "John Doe",
			"email": "johndoe@email.com",
			"phone": "",
			"cpf":   "151.515.423-22",
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

		var customer models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
			t.Error("Failed to parse JSON: ", err)
		}

		if customer.ID == 0 {
			t.Error("Expected ID, got 0")
		}

		if customer.Name != "John Doe" {
			t.Errorf("Expected name %v, got %v", "John Doe", customer.Name)
		}

		if customer.Address == nil {
			t.Error("Should have an address")
		}

		if customer.Address.City != "New York" {
			t.Errorf("Expected state %v, got %v", "New York", customer.Address.City)
		}

		if customer.Address.State != "NY" {
			t.Errorf("Expected state %v, got %v", "NY", customer.Address.State)
		}
	})

	t.Run("Validation", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "",
			"email": "johndoe",
			"phone": "",
			"cpf":   "",
			"address": map[string]interface{}{
				"street":       "",
				"number":       "",
				"city":         "",
				"state":        "",
				"neighborhood": "",
				"postcode":     "",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Address should not be required", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "Jane Doe",
			"email": "janedoe@email.com",
			"phone": "",
			"cpf":   "515.151.222-66",
			"address": map[string]interface{}{
				"street":       "",
				"number":       "",
				"city":         "",
				"state":        "",
				"neighborhood": "",
				"postcode":     "",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var customer models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
			t.Error("Failed to parse JSON", err)
		}

		if customer.Address != nil {
			t.Errorf("Should not have an address, got %v", customer.Address)
		}
	})

	t.Run("Email should be valid", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "Jane Doe",
			"email": "janedoe",
			"phone": "",
			"cpf":   "515.151.222-66",
			"address": map[string]interface{}{
				"street":       "street",
				"number":       "533",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "75397",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("List", func(t *testing.T) {
		db.Create(&models.Company{Name: "Other Company"})

		// this customer shouldn't be retrieved
		db.Create(&models.Customer{
			Name:      "From another company",
			CompanyID: 2,
		})

		req := Get(t, "/customers")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var items []models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(items))
		}

		for idx, customer := range items {
			if idx == 0 && customer.Name != "John Doe" {
				t.Errorf("Expected name %v, got %v", "John Doe", customer.Name)
			}

			if idx == 1 && customer.Name != "Jane Doe" {
				t.Errorf("Expected name %v, got %v", "Jane Doe", customer.Name)
			}
		}
	})
}
