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
	db.AutoMigrate(&models.Product{})
	db.AutoMigrate(&models.StockEntry{})
	db.AutoMigrate(&models.StockUsage{})
	db.AutoMigrate(&models.Transaction{})
	db.AutoMigrate(&models.Consumption{})
	db.AutoMigrate(&models.ServicePerformed{})

	db.Create(&models.Company{Name: "Testing Company"})
	db.Create(&models.Account{Name: "Revenue", Type: models.Revenue, CompanyID: 1})
	db.Create(&models.Account{Name: "Inventory", Type: models.Asset, CompanyID: 1})
	db.Create(&models.Account{Name: "Cost of Service", Type: models.Expense, CompanyID: 1})

	cash := &models.Account{
		CompanyID: 1,
		Name:      "Cash",
		Type:      models.Asset,
	}
	db.Create(cash)

	router := api.GetRouter()

	t.Run("Create", func(t *testing.T) {
		req := Post(t, "/services", map[string]interface{}{
			"name":                       "Window cleaning",
			"revenue_account_id":         1,
			"cost_of_service_account_id": 3,
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
			"name":                       "",
			"revenue_account_id":         "",
			"cost_of_service_account_id": "",
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

		db.Create(&models.Service{
			Name:                   "Renting",
			RevenueAccountID:       1,
			CostOfServiceAccountID: 3,
			CompanyID:              2,
		})

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
			"name":                       "General Cleaning",
			"revenue_account_id":         5,
			"cost_of_service_account_id": 3,
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

		if service.RevenueAccountID != 5 {
			t.Errorf("Expected account %v, got %v", 5, service.RevenueAccountID)
		}
	})

	t.Run("Perform", func(t *testing.T) {
		db.Create(&models.Product{
			Name:               "Sponge",
			Price:              5,
			Purchasable:        false,
			InventoryAccountID: 2,
			CompanyID:          1,
			StockEntries: []*models.StockEntry{
				{Qty: 200, Price: 3},
				{Qty: 200, Price: 2},
			},
		})

		db.Create(&models.Product{
			Name:               "Alcohol",
			Price:              7,
			Purchasable:        false,
			InventoryAccountID: 2,
			CompanyID:          1,
			StockEntries: []*models.StockEntry{
				{Qty: 200, Price: 5},
				{Qty: 200, Price: 4},
			},
		})

		req := Post(t, "/services/performed", map[string]interface{}{
			"paid":                  true,
			"value":                 122,
			"service_id":            1,
			"payment_account_id":    &cash.ID,
			"receivable_account_id": nil,
			"consumptions": []map[string]interface{}{
				{"product_id": 1, "qty": 10},
				{"product_id": 2, "qty": 10},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var performed *models.ServicePerformed
		if err := json.Unmarshal(w.Body.Bytes(), &performed); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if performed.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, performed.ID)
		}

		if len(performed.Consumptions) != 2 {
			t.Errorf("Expected %v consumptions, got %v", 2, len(performed.Consumptions))
		}

		for idx, consumption := range performed.Consumptions {
			if consumption.ProductID != uint(idx+1) {
				t.Errorf("Expected product %v, got %v", idx+1, consumption.ProductID)
			}
			if consumption.Qty != 10 {
				t.Errorf("Expected qty %v, got %v", 10, consumption.Qty)
			}
		}

		// Check if revenue account is increased
		var rev *models.Account
		if result := db.Preload("Transactions").First(&rev, 5); result.Error != nil {
			t.Error("Should retrieve revenue account", result.Error)
		}

		if rev.Balance() != 122 {
			t.Errorf("Expected balance %v, got %v", 122, rev.Balance())
		}

		// Check if payment account is increased
		var pay *models.Account
		if result := db.Preload("Transactions").First(&pay, cash.ID); result.Error != nil {
			t.Error("Should retrieve revenue account", result.Error)
		}

		if pay.Balance() != 122 {
			t.Errorf("Expected balance %v, got %v", 122, pay.Balance())
		}

		// Check if products inventory is reduced
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, 2).Error != nil {
			t.Error("Should retrieve inventory account")
		}

		if inv.Balance() != -80 {
			t.Errorf("Expected balance %v, got %v", -80, inv.Balance())
		}

		// Check if products stock are decreased
		var sponge *models.Product
		if db.Preload("StockEntries.StockUsages").First(&sponge, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if sponge.Inventory() != 390 {
			t.Errorf("Expected stock %v, got %v", 390, sponge.Inventory())
		}

		var alcohol *models.Product
		if db.Preload("StockEntries.StockUsages").First(&alcohol, 2).Error != nil {
			t.Error("Should retrieve product")
		}

		if alcohol.Inventory() != 390 {
			t.Errorf("Expected stock %v, got %v", 390, alcohol.Inventory())
		}
	})

	t.Run("Perform paid without payment account", func(t *testing.T) {
		req := Post(t, "/services/performed", map[string]interface{}{
			"paid":                  true,
			"value":                 122,
			"service_id":            1,
			"payment_account_id":    nil,
			"receivable_account_id": nil,
			"consumptions": []map[string]interface{}{
				{"product_id": 1, "qty": 10},
				{"product_id": 2, "qty": 10},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if response["error"] != api.ErrPaymentAccountMissing.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrPaymentAccountMissing, response["error"])
		}
	})

	t.Run("Perform not paid without receivable account", func(t *testing.T) {
		req := Post(t, "/services/performed", map[string]interface{}{
			"paid":                  false,
			"value":                 122,
			"service_id":            1,
			"payment_account_id":    nil,
			"receivable_account_id": nil,
			"consumptions": []map[string]interface{}{
				{"product_id": 1, "qty": 10},
				{"product_id": 2, "qty": 10},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if response["error"] != api.ErrReceivableAccountMissing.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrReceivableAccountMissing, response["error"])
		}
	})

	t.Run("Update performed", func(t *testing.T) {
		bank := &models.Account{Name: "Bank", Type: models.Asset, CompanyID: 1}
		db.Create(bank)

		req := Put(t, "/services/performed/1", map[string]interface{}{
			"paid":                  true,
			"value":                 222,
			"service_id":            1,
			"payment_account_id":    &bank.ID,
			"receivable_account_id": nil,
			"consumptions": []map[string]interface{}{
				{"product_id": 1, "qty": 20},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var performed *models.ServicePerformed
		if err := json.Unmarshal(w.Body.Bytes(), &performed); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if performed.Value != 222 {
			t.Errorf("Expected value %v, got %v", 222, performed.Value)
		}

		if len(performed.Consumptions) != 1 {
			t.Errorf("Expected %v item, got %v", 1, len(performed.Consumptions))
		}

		// Check if revenue account is updated
		var rev *models.Account
		if result := db.Preload("Transactions").First(&rev, 5); result.Error != nil {
			t.Error("Should retrieve revenue account", result.Error)
		}

		if rev.Balance() != 222 {
			t.Errorf("Expected balance %v, got %v", 222, rev.Balance())
		}

		// Check if cash account is updated
		var prev *models.Account
		if result := db.Preload("Transactions").First(&prev, cash.ID); result.Error != nil {
			t.Error("Should retrieve revenue account", result.Error)
		}

		if prev.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, prev.Balance())
		}

		// Check if payment account is updated
		var pay *models.Account
		if result := db.Preload("Transactions").First(&pay, bank.ID); result.Error != nil {
			t.Error("Should retrieve revenue account", result.Error)
		}

		if pay.Balance() != 222 {
			t.Errorf("Expected balance %v, got %v", 222, pay.Balance())
		}

		// Check if products inventory is updated
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, 2).Error != nil {
			t.Error("Should retrieve inventory account")
		}

		if inv.Balance() != -60 {
			t.Errorf("Expected balance %v, got %v", -60, inv.Balance())
		}

		// Check if products stock are updated
		var sponge *models.Product
		if db.Preload("StockEntries.StockUsages").First(&sponge, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if sponge.Inventory() != 380 {
			t.Errorf("Expected stock %v, got %v", 380, sponge.Inventory())
		}

		var alcohol *models.Product
		if db.Preload("StockEntries.StockUsages").First(&alcohol, 2).Error != nil {
			t.Error("Should retrieve product")
		}

		if alcohol.Inventory() != 400 {
			t.Errorf("Expected stock %v, got %v", 400, alcohol.Inventory())
		}
	})

	t.Run("Update to not paid", func(t *testing.T) {
		receivable := &models.Account{Name: "Receivable", Type: models.Asset, CompanyID: 1}
		db.Create(receivable)

		req := Put(t, "/services/performed/1", map[string]interface{}{
			"paid":                  false,
			"value":                 422,
			"service_id":            1,
			"payment_account_id":    nil,
			"receivable_account_id": &receivable.ID,
			"consumptions": []map[string]interface{}{
				{"product_id": 1, "qty": 20},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var performed *models.ServicePerformed
		if err := json.Unmarshal(w.Body.Bytes(), &performed); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if performed.PaymentAccountID != nil {
			t.Errorf("Should not have payment account, got %v", *performed.PaymentAccountID)
		}

		// Check if revenue account is updated
		var rev *models.Account
		if result := db.Preload("Transactions").First(&rev, 5); result.Error != nil {
			t.Error("Should retrieve revenue account", result.Error)
		}

		if rev.Balance() != 422 {
			t.Errorf("Expected balance %v, got %v", 422, rev.Balance())
		}

		// Check if payment account is updated
		var pay *models.Account
		if result := db.Preload("Transactions").First(&pay, 6); result.Error != nil {
			t.Error("Should retrieve payment account", result.Error)
		}

		if pay.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, pay.Balance())
		}

		// Check if receivable account is updated
		var recv *models.Account
		if result := db.Preload("Transactions").First(&recv, receivable.ID); result.Error != nil {
			t.Error("Should retrieve receivable account", result.Error)
		}

		if recv.Balance() != 422 {
			t.Errorf("Expected balance %v, got %v", 422, recv.Balance())
		}
	})

	t.Run("Update to paid", func(t *testing.T) {
		req := Put(t, "/services/performed/1", map[string]interface{}{
			"paid":                  true,
			"value":                 522,
			"service_id":            1,
			"payment_account_id":    &cash.ID,
			"receivable_account_id": nil,
			"consumptions": []map[string]interface{}{
				{"product_id": 1, "qty": 20},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var performed *models.ServicePerformed
		if err := json.Unmarshal(w.Body.Bytes(), &performed); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if performed.ReceivableAccountID != nil {
			t.Errorf("Should not have receivable account, got %v", *performed.ReceivableAccountID)
		}

		// Check if revenue account is updated
		var rev *models.Account
		if result := db.Preload("Transactions").First(&rev, 5); result.Error != nil {
			t.Error("Should retrieve revenue account", result.Error)
		}

		if rev.Balance() != 522 {
			t.Errorf("Expected balance %v, got %v", 522, rev.Balance())
		}

		// Check if payment account is updated
		var pay *models.Account
		if result := db.Preload("Transactions").First(&pay, cash.ID); result.Error != nil {
			t.Error("Should retrieve payment account", result.Error)
		}

		if pay.Balance() != 522 {
			t.Errorf("Expected balance %v, got %v", 522, pay.Balance())
		}

		// Check if receivable account is updated
		var recv *models.Account
		if result := db.Preload("Transactions").First(&recv, 7); result.Error != nil {
			t.Error("Should retrieve receivable account", result.Error)
		}

		if recv.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, recv.Balance())
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

	t.Run("Delete", func(t *testing.T) {
		req := Delete(t, "/services/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}

		var service *models.Service
		if db.First(&service, 1).Error == nil {
			t.Error("Should have deleted service")
		}
	})

	t.Run("Delete non existent", func(t *testing.T) {
		req := Delete(t, "/services/4121")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete invalid", func(t *testing.T) {
		req := Delete(t, "/services/snth")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete from another company", func(t *testing.T) {
		req := Delete(t, "/services/2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete deleted", func(t *testing.T) {
		req := Delete(t, "/services/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

}
