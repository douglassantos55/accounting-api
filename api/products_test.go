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

func TestProducts(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Account{})
	db.AutoMigrate(&models.Product{})
	db.AutoMigrate(&models.Company{})

	db.Create(&models.Company{Name: "Testing Company"})

	db.Create(&models.Account{
		Name:      "Revenue",
		Type:      models.Revenue,
		CompanyID: 1,
	})

	db.Create(&models.Account{
		Name:      "Cost of Good Sold",
		Type:      models.Expense,
		CompanyID: 1,
	})

	db.Create(&models.Account{
		Name:      "Inventory",
		Type:      models.Asset,
		CompanyID: 1,
	})

	router := api.GetRouter()

	t.Run("Create", func(t *testing.T) {
		req := Post(t, "/products", map[string]interface{}{
			"name":                    "Product 1",
			"price":                   53.54,
			"purchasable":             true,
			"revenue_account_id":      1,
			"cost_of_sale_account_id": 2,
			"inventory_account_id":    3,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var prod *models.Product
		if err := json.Unmarshal(w.Body.Bytes(), &prod); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if prod.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, prod.ID)
		}

		if prod.Name != "Product 1" {
			t.Errorf("Expected name %v, got %v", "Product 1", prod.Name)
		}

		if *prod.RevenueAccountID != 1 {
			t.Errorf("Expected revenuen account %v, got %v", 1, *prod.RevenueAccountID)
		}
	})

	t.Run("Required revenue account", func(t *testing.T) {
		req := Post(t, "/products", map[string]interface{}{
			"name":                    "Product 1",
			"price":                   53.54,
			"purchasable":             true,
			"revenue_account_id":      nil,
			"cost_of_sale_account_id": 2,
			"inventory_account_id":    3,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if response["error"] != api.ErrRevenueAccountMissing.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrRevenueAccountMissing.Error(), response["error"])
		}
	})

	t.Run("Required cost of sale account", func(t *testing.T) {
		req := Post(t, "/products", map[string]interface{}{
			"name":                    "Product 1",
			"price":                   53.54,
			"purchasable":             true,
			"revenue_account_id":      1,
			"cost_of_sale_account_id": nil,
			"inventory_account_id":    3,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if response["error"] != api.ErrCostOfSaleAccountMissing.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrCostOfSaleAccountMissing.Error(), response["error"])
		}
	})

	t.Run("Required inventory account", func(t *testing.T) {
		req := Post(t, "/products", map[string]interface{}{
			"name":                    "Product 1",
			"price":                   53.54,
			"purchasable":             true,
			"revenue_account_id":      1,
			"cost_of_sale_account_id": 2,
			"inventory_account_id":    nil,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %v, got %v", http.StatusInternalServerError, w.Code)
		}

		req = Post(t, "/products", map[string]interface{}{
			"name":                    "Product 1",
			"price":                   53.54,
			"purchasable":             false,
			"revenue_account_id":      nil,
			"cost_of_sale_account_id": nil,
			"inventory_account_id":    nil,
		})

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %v, got %v", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("Create non purchasable", func(t *testing.T) {
		req := Post(t, "/products", map[string]interface{}{
			"name":                    "Product 2",
			"price":                   153.54,
			"purchasable":             false,
			"revenue_account_id":      nil,
			"cost_of_sale_account_id": nil,
			"inventory_account_id":    3,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var prod *models.Product
		if err := json.Unmarshal(w.Body.Bytes(), &prod); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if prod.ID != 2 {
			t.Errorf("Expected ID %v, got %v", 2, prod.ID)
		}

		if prod.RevenueAccountID != nil {
			t.Errorf("Should not have revenue account, got %v", prod.RevenueAccountID)
		}

		if prod.CostOfSaleAccountID != nil {
			t.Errorf("Should not have revenue account, got %v", prod.CostOfSaleAccountID)
		}
	})
}
