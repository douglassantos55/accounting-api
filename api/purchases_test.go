package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/accounting/api"
	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/models"
)

func TestPurchase(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Entry{})
	db.AutoMigrate(&models.Product{})
	db.AutoMigrate(&models.StockEntry{})
	db.AutoMigrate(&models.Transaction{})
	db.AutoMigrate(&models.Purchase{})

	t.Cleanup(database.Cleanup)

	router := api.GetRouter()

	events.Handle(events.PurchaseCreated, api.CreateStockEntry)
	events.Handle(events.PurchaseCreated, api.CreateAccountingEntry)

	events.Handle(events.PurchaseUpdated, api.UpdateStockEntry)
	events.Handle(events.PurchaseUpdated, api.UpdateAccountingEntry)

	db.Create(&models.Company{Name: "Testing Company"})

	// ID: 1
	cash := &models.Account{
		Name:      "Cash & Equivalents",
		Type:      models.Asset,
		CompanyID: 1,
	}
	db.Create(&cash)

	// ID: 2
	revenue := &models.Account{
		Name:      "Revenue",
		Type:      models.Revenue,
		CompanyID: 1,
	}
	db.Create(&revenue)

	// ID: 3
	inventory := &models.Account{
		Name:      "Inventory",
		Type:      models.Asset,
		CompanyID: 1,
	}
	db.Create(&inventory)

	// ID: 4
	cogs := &models.Account{
		Name:      "Cost of Goods Sold",
		Type:      models.Expense,
		CompanyID: 1,
	}
	db.Create(&cogs)

	// ID: 5
	receivables := &models.Account{
		Name:      "Receivables",
		Type:      models.Liability,
		CompanyID: 1,
	}
	db.Create(&receivables)

	db.Create(&models.Product{
		CompanyID:           1,
		Name:                "Product",
		Price:               100,
		Purchasable:         true,
		RevenueAccountID:    &revenue.ID,
		InventoryAccountID:  inventory.ID,
		CostOfSaleAccountID: &cogs.ID,
	})

	t.Run("Create without payment account", func(t *testing.T) {
		req := Post(t, "/purchases", map[string]interface{}{
			"Qty":              5,
			"Price":            155.75,
			"Paid":             true,
			"ProductID":        1,
			"PaymentDate":      time.Now(),
			"PaymentAccountID": nil,
			"PayableAccountID": nil,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create without payable account", func(t *testing.T) {
		req := Post(t, "/purchases", map[string]interface{}{
			"Qty":              5,
			"Price":            155.75,
			"Paid":             false,
			"ProductID":        1,
			"PaymentDate":      nil,
			"PaymentAccountID": nil,
			"PayableAccountID": nil,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create paid", func(t *testing.T) {
		req := Post(t, "/purchases", map[string]interface{}{
			"Qty":              5,
			"Price":            155.75,
			"Paid":             true,
			"ProductID":        1,
			"PaymentDate":      time.Now(),
			"PaymentAccountID": cash.ID,
			"PayableAccountID": nil,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchase *models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchase); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if purchase.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, purchase.ID)
		}

		if purchase.Qty != 5 {
			t.Errorf("Expected qty %v, got %v", 5, purchase.Qty)
		}

		if *purchase.PaymentAccountID != cash.ID {
			t.Errorf("Expected payment account %v, got %v", cash.ID, *purchase.PayableAccountID)
		}

		// Check if stock entries are created
		var product *models.Product
		if db.Preload("StockEntries").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 5 {
			t.Errorf("Expected %v stock, got %v", 5, product.Inventory())
		}

		// Check if payment account is reduced
		var payment *models.Account
		if db.Preload("Transactions").First(&payment, cash.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payment.Balance() != -5*155.75 {
			t.Errorf("Expected balance %v, got %v", -5*155.75, payment.Balance())
		}

		// Checks if inventory account is increased
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != 5*155.75 {
			t.Errorf("Expected balance %v, got %v", 5*155.75, inv.Balance())
		}
	})

	t.Run("Create not paid", func(t *testing.T) {
		req := Post(t, "/purchases", map[string]interface{}{
			"Qty":              5,
			"Price":            155.75,
			"Paid":             false,
			"ProductID":        1,
			"PaymentDate":      nil,
			"PaymentAccountID": nil,
			"PayableAccountID": receivables.ID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchase *models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchase); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if purchase.ID != 2 {
			t.Errorf("Expected ID %v, got %v", 2, purchase.ID)
		}

		if purchase.Qty != 5 {
			t.Errorf("Expected qty %v, got %v", 5, purchase.Qty)
		}

		if *purchase.PayableAccountID != receivables.ID {
			t.Errorf("Expected payment account %v, got %v", receivables.ID, *purchase.PayableAccountID)
		}

		// Check if stock entries are created
		var product *models.Product
		if db.Preload("StockEntries").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 10 {
			t.Errorf("Expected %v stock, got %v", 10, product.Inventory())
		}

		// Check if payable account is increased
		var payable *models.Account
		if db.Preload("Transactions").First(&payable, receivables.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payable.Balance() != 5*155.75 {
			t.Errorf("Expected balance %v, got %v", 5*155.75, payable.Balance())
		}

		// Checks if inventory account is increased
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != 10*155.75 {
			t.Errorf("Expected balance %v, got %v", 10*155.75, inv.Balance())
		}
	})

	t.Run("List", func(t *testing.T) {
		db.Create(&models.Company{Name: "Other company"})

		// This should not be retrieved
		db.Create(&models.Purchase{
			Qty:       1,
			Price:     1,
			CompanyID: 2,
			ProductID: 1,
		})

		req := Get(t, "/purchases")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchases []models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchases); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(purchases) != 2 {
			t.Errorf("Expected %v purchases, got %v", 2, len(purchases))
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := Get(t, "/purchases/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchase models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchase); err != nil {
			t.Error("Failed to parse JSON", err)
		}

		if purchase.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, purchase.ID)
		}
	})

	t.Run("Get non existent", func(t *testing.T) {
		req := Get(t, "/purchases/132")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get invalid", func(t *testing.T) {
		req := Get(t, "/purchases/asontehu")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get from another company", func(t *testing.T) {
		req := Get(t, "/purchases/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update paid", func(t *testing.T) {
		req := Put(t, "/purchases/1", map[string]interface{}{
			"Qty":              10,
			"Price":            155.75,
			"Paid":             true,
			"ProductID":        1,
			"PaymentDate":      time.Now(),
			"PaymentAccountID": cash.ID,
			"PayableAccountID": nil,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchase models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchase); err != nil {
			t.Error("failed parsing JSON", err)
		}

		if purchase.Qty != 10 {
			t.Errorf("Expected qty %v, got %v", 10, purchase.Qty)
		}

		// Check if stock entries are updated
		var product *models.Product
		if db.Preload("StockEntries").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 15 {
			t.Errorf("Expected %v stock, got %v", 15, product.Inventory())
		}

		// Check if payment account is updated
		var payment *models.Account
		if db.Preload("Transactions").First(&payment, cash.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payment.Balance() != -10*155.75 {
			t.Errorf("Expected balance %v, got %v", -10*155.75, payment.Balance())
		}

		// Checks if inventory account is updated
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != 15*155.75 {
			t.Errorf("Expected balance %v, got %v", 15*155.75, inv.Balance())
		}
	})

	t.Run("Update not paid", func(t *testing.T) {
		req := Put(t, "/purchases/2", map[string]interface{}{
			"Qty":              10,
			"Price":            155.75,
			"Paid":             false,
			"ProductID":        1,
			"PaymentDate":      nil,
			"PaymentAccountID": nil,
			"PayableAccountID": receivables.ID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchase models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchase); err != nil {
			t.Error("failed parsing JSON", err)
		}

		if purchase.Qty != 10 {
			t.Errorf("Expected qty %v, got %v", 10, purchase.Qty)
		}

		// Check if stock entries are updated
		var product *models.Product
		if db.Preload("StockEntries").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 20 {
			t.Errorf("Expected %v stock, got %v", 20, product.Inventory())
		}

		// Check if payment account remains the same
		var payment *models.Account
		if db.Preload("Transactions").First(&payment, cash.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payment.Balance() != -10*155.75 {
			t.Errorf("Expected balance %v, got %v", -10*155.75, payment.Balance())
		}

		// Check if payable account is updated
		var payable *models.Account
		if db.Preload("Transactions").First(&payable, receivables.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payable.Balance() != 10*155.75 {
			t.Errorf("Expected balance %v, got %v", 10*155.75, payable.Balance())
		}

		// Checks if inventory account is updated
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != 20*155.75 {
			t.Errorf("Expected balance %v, got %v", 20*155.75, inv.Balance())
		}
	})

	t.Run("Update to paid", func(t *testing.T) {
		req := Put(t, "/purchases/2", map[string]interface{}{
			"Qty":              10,
			"Price":            155.75,
			"Paid":             true,
			"ProductID":        1,
			"PaymentDate":      time.Now(),
			"PaymentAccountID": cash.ID,
			"PayableAccountID": receivables.ID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchase models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchase); err != nil {
			t.Error("failed parsing JSON", err)
		}

		if !purchase.Paid {
			t.Error("Expected purchase to be paid")
		}

		// Check if payable account is reduced
		var payable *models.Account
		if db.Preload("Transactions").First(&payable, receivables.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payable.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, payable.Balance())
		}

		// Check if payment account is reduced
		var payment *models.Account
		if db.Preload("Transactions").First(&payment, cash.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payment.Balance() != -20*155.75 {
			t.Errorf("Expected balance %v, got %v", -20*155.75, payment.Balance())
		}

		// Checks if inventory account is updated
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != 20*155.75 {
			t.Errorf("Expected balance %v, got %v", 20*155.75, inv.Balance())
		}
	})

	t.Run("Update to not paid", func(t *testing.T) {
		req := Put(t, "/purchases/1", map[string]interface{}{
			"Qty":              10,
			"Price":            155.75,
			"Paid":             false,
			"ProductID":        1,
			"PaymentDate":      time.Now(),
			"PaymentAccountID": cash.ID,
			"PayableAccountID": receivables.ID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var purchase models.Purchase
		if err := json.Unmarshal(w.Body.Bytes(), &purchase); err != nil {
			t.Error("failed parsing JSON", err)
		}

		if purchase.Paid {
			t.Error("Purchase should not be paid")
		}

		// Check if payable account is increased
		var payable *models.Account
		if db.Preload("Transactions").First(&payable, receivables.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payable.Balance() != 10*155.75 {
			t.Errorf("Expected balance %v, got %v", 10*155.750, payable.Balance())
		}

		// Check if payable account is reduced
		var payment *models.Account
		if db.Preload("Transactions").First(&payment, cash.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payment.Balance() != -10*155.75 {
			t.Errorf("Expected balance %v, got %v", -10*155.75, payment.Balance())
		}

		// Checks if inventory account is updated
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != 20*155.75 {
			t.Errorf("Expected balance %v, got %v", 20*155.75, inv.Balance())
		}
	})

	t.Run("Delete paid", func(t *testing.T) {
		req := Delete(t, "/purchases/2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}

		if db.First(&models.Purchase{}, 2).Error == nil {
			t.Error("Should have deleted purchase")
		}

		// Check if product stock is updated
		var product *models.Product
		db.Preload("StockEntries").First(&product, 1)

		if product.Inventory() != 10 {
			t.Errorf("Expected stock %v, got %v", 10, product.Inventory())
		}

		// Check if payment account is updated
		var payment *models.Account
		db.Preload("Transactions").First(&payment, cash.ID)

		if payment.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, payment.Balance())
		}

		// Check if inventory account is updated
		var inv *models.Account
		db.Preload("Transactions").First(&inv, inventory.ID)

		if inv.Balance() != 10*155.75 {
			t.Errorf("Expected balance %v, got %v", 10*155.75, inv.Balance())
		}
	})

	t.Run("Delete not paid", func(t *testing.T) {
		req := Delete(t, "/purchases/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}

		if db.First(&models.Purchase{}, 1).Error == nil {
			t.Error("Should have deleted purchase")
		}

		// Check if product stock is updated
		var product *models.Product
		db.Preload("StockEntries").First(&product, 1)

		if product.Inventory() != 0 {
			t.Errorf("Expected stock %v, got %v", 0, product.Inventory())
		}

		// Check if payable account is updated
		var payable *models.Account
		db.Preload("Transactions").First(&payable, receivables.ID)

		if payable.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, payable.Balance())
		}

		// Check if inventory account is updated
		var inv *models.Account
		db.Preload("Transactions").First(&inv, inventory.ID)

		if inv.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, inv.Balance())
		}
	})

	t.Run("Delete non existent", func(t *testing.T) {
		req := Delete(t, "/purchases/1423")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete invalid", func(t *testing.T) {
		req := Delete(t, "/purchases/asontehu")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete from another company", func(t *testing.T) {
		req := Delete(t, "/purchases/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})
}
