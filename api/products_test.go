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

	t.Cleanup(database.Cleanup)

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
			"inventory_account_id":    "",
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
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

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
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

	t.Run("List", func(t *testing.T) {
		// create another company
		db.Create(&models.Company{Name: "Other Company"})

		// this product should not be retrieved
		db.Create(&models.Product{
			Name:               "Product 3",
			Price:              153.54,
			Purchasable:        false,
			InventoryAccountID: 3,
		})

		req := Get(t, "/products")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var products []*models.Product
		if err := json.Unmarshal(w.Body.Bytes(), &products); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(products) != 2 {
			t.Errorf("Expected %v products, got %v", 2, len(products))
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := Get(t, "/products/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, w.Code)
		}

		var product *models.Product
		if err := json.Unmarshal(w.Body.Bytes(), &product); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if product.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, product.ID)
		}
	})

	t.Run("Get non existent", func(t *testing.T) {
		req := Get(t, "/products/15151")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get invalid", func(t *testing.T) {
		req := Get(t, "/products/astnoheu")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get from another company", func(t *testing.T) {
		req := Get(t, "/products/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update non purchasable", func(t *testing.T) {
		req := Put(t, "/products/1", map[string]interface{}{
			"name":                    "Edited product",
			"price":                   63.64,
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

		var product *models.Product
		if err := json.Unmarshal(w.Body.Bytes(), &product); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if product.Name != "Edited product" {
			t.Errorf("Expected name %v, got %v", "Edited product", product.Name)
		}

		if product.Price != 63.64 {
			t.Errorf("Expected price %v, got %v", 63.64, product.Price)
		}

		if product.Purchasable {
			t.Error("Should not be purchasable")
		}
	})

	t.Run("Update required revenue account", func(t *testing.T) {
		req := Put(t, "/products/2", map[string]interface{}{
			"name":                    "Edited product 2",
			"price":                   63.64,
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

	t.Run("Update required cost of sale account", func(t *testing.T) {
		req := Put(t, "/products/2", map[string]interface{}{
			"name":                    "Edited product 2",
			"price":                   63.64,
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

	t.Run("Update required inventory account", func(t *testing.T) {
		req := Put(t, "/products/2", map[string]interface{}{
			"name":                    "Edited product 2",
			"price":                   63.64,
			"purchasable":             true,
			"revenue_account_id":      1,
			"cost_of_sale_account_id": 2,
			"inventory_account_id":    "",
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Update non existent", func(t *testing.T) {
		req := Put(t, "/products/51515", map[string]interface{}{
			"name":                    "Edited product",
			"price":                   63.64,
			"purchasable":             false,
			"revenue_account_id":      nil,
			"cost_of_sale_account_id": nil,
			"inventory_account_id":    3,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update invalid", func(t *testing.T) {
		req := Put(t, "/products/aoeu", map[string]interface{}{
			"name":                    "Edited product",
			"price":                   63.64,
			"purchasable":             false,
			"revenue_account_id":      nil,
			"cost_of_sale_account_id": nil,
			"inventory_account_id":    3,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update from another company", func(t *testing.T) {
		req := Put(t, "/products/3", map[string]interface{}{
			"name":                    "Edited product",
			"price":                   63.64,
			"purchasable":             false,
			"revenue_account_id":      nil,
			"cost_of_sale_account_id": nil,
			"inventory_account_id":    3,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		req := Delete(t, "/products/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}

		db, _ := database.GetConnection()
		if db.First(&models.Product{}, 1).Error == nil {
			t.Error("Should have deleted product")
		}
	})

	t.Run("Delete non existent", func(t *testing.T) {
		req := Delete(t, "/products/210")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete invalid", func(t *testing.T) {
		req := Delete(t, "/products/aosnetuh")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete from another company", func(t *testing.T) {
		req := Delete(t, "/products/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create with vendor", func(t *testing.T) {
		db, _ := database.GetConnection()

		db.Create(&models.Vendor{
			CompanyID: 1,
			Name:      "Vendor",
			Cnpj:      "70.463.497/0001-60",
		})

		req := Post(t, "/products", map[string]interface{}{
			"name":                    "Product 4",
			"price":                   253.24,
			"purchasable":             true,
			"revenue_account_id":      1,
			"cost_of_sale_account_id": 2,
			"inventory_account_id":    3,
			"vendor_id":               1,
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

		if *prod.VendorID != 1 {
			t.Errorf("Expected vendor %v, got %v", 1, *prod.VendorID)
		}
	})
}
