package api_test

import (
	"encoding/json"
	"fmt"
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

	db.Create(&models.Company{Name: "Testing Company"})

	// ID: 1
	cash := &models.Account{
		Name:      "Cash & Equivalents",
		Type:      models.Asset,
		CompanyID: 1,
	}
	db.Create(&cash)
	fmt.Printf("cash.ID: %v\n", cash.ID)

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
			"qty":                5,
			"price":              155.75,
			"paid":               true,
			"product_id":         1,
			"payment_date":       time.Now(),
			"payment_account_id": nil,
			"payable_account_id": nil,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create without payable account", func(t *testing.T) {
		req := Post(t, "/purchases", map[string]interface{}{
			"qty":                5,
			"price":              155.75,
			"paid":               false,
			"product_id":         1,
			"payment_date":       nil,
			"payment_account_id": nil,
			"payable_account_id": nil,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create paid", func(t *testing.T) {
		req := Post(t, "/purchases", map[string]interface{}{
			"qty":                5,
			"price":              155.75,
			"paid":               true,
			"product_id":         1,
			"payment_date":       time.Now(),
			"payment_account_id": cash.ID,
			"payable_account_id": nil,
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
			"qty":                5,
			"price":              155.75,
			"paid":               false,
			"product_id":         1,
			"payment_date":       nil,
			"payment_account_id": nil,
			"payable_account_id": receivables.ID,
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
}
