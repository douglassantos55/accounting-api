package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/accounting/api"
	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/models"
)

func TestSales(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Sale{})
	db.AutoMigrate(&models.Item{})
	db.AutoMigrate(&models.Account{})
	db.AutoMigrate(&models.Company{})
	db.AutoMigrate(&models.Customer{})
	db.AutoMigrate(&models.StockEntry{})
	db.AutoMigrate(&models.Transaction{})

	t.Cleanup(database.Cleanup)

	events.Handle(events.SaleCreated, api.ReduceProductStock)
	events.Handle(events.SaleCreated, api.CreateAccountingEntries)

	if result := db.Create(&models.Company{Name: "Testing Company"}); result.Error != nil {
		t.Error(result.Error)
	}

	db.Create(&models.Customer{
		Name:      "Customer",
		CompanyID: 1,
	})

	cash := &models.Account{
		Name:      "Cash",
		Type:      models.Asset,
		CompanyID: 1,
	}
	db.Create(cash)

	receivables := &models.Account{
		Name:      "Receivables",
		Type:      models.Asset,
		CompanyID: 1,
	}
	db.Create(receivables)

	revenue := &models.Account{
		Name:      "Revenue",
		Type:      models.Revenue,
		CompanyID: 1,
	}
	db.Create(revenue)

	cogs := &models.Account{
		Name:      "Cost of goods sold",
		Type:      models.Expense,
		CompanyID: 1,
	}
	db.Create(cogs)

	inventory := &models.Account{
		Name:      "Inventory",
		Type:      models.Asset,
		CompanyID: 1,
	}
	db.Create(inventory)

	db.Create(&models.Product{
		Name:                "Product 1",
		Price:               150,
		CompanyID:           1,
		InventoryAccountID:  inventory.ID,
		CostOfSaleAccountID: &cogs.ID,
		RevenueAccountID:    &revenue.ID,
		Purchasable:         true,
		StockEntries: []*models.StockEntry{
			{Price: 100, Qty: 100}, // 10000
			{Price: 90, Qty: 100},  // 9000
		},
	})

	db.Create(&models.Product{
		Name:                "Product 2",
		Price:               250,
		CompanyID:           1,
		InventoryAccountID:  inventory.ID,
		CostOfSaleAccountID: &cogs.ID,
		RevenueAccountID:    &revenue.ID,
		Purchasable:         true,
		StockEntries: []*models.StockEntry{
			{Price: 200, Qty: 100}, // 20000
			{Price: 190, Qty: 100}, // 19000
		},
	})

	router := api.GetRouter()

	t.Run("Create paid", func(t *testing.T) {
		req := Post(t, "/sales", map[string]interface{}{
			"paid":                  true,
			"customer_id":           1,
			"payment_account_id":    cash.ID,
			"receivable_account_id": nil,
			"items": []map[string]interface{}{
				{"qty": 10, "price": 200, "product_id": 1},
				{"qty": 10, "price": 250, "product_id": 2},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var sale models.Sale
		if err := json.Unmarshal(w.Body.Bytes(), &sale); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if sale.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, sale.ID)
		}

		if !sale.Paid {
			t.Error("Should be paid")
		}

		if sale.CustomerID != 1 {
			t.Errorf("Expected user %v, got %v", 1, sale.CustomerID)
		}

		if len(sale.Items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(sale.Items))
		}

		if sale.Total() != 4500 {
			t.Errorf("Expected total %v, got %v", 4500, sale.Total())
		}

		for idx, item := range sale.Items {
			if idx == 0 {
				if item.Price != 200 {
					t.Errorf("Expected price %v, got %v", 200, item.Price)
				}
				if item.ProductID != 1 {
					t.Errorf("Expected product %v, got %v", 1, item.ProductID)
				}
			}

			if idx == 1 {
				if item.Price != 250 {
					t.Errorf("Expected price %v, got %v", 250, item.Price)
				}
				if item.ProductID != 2 {
					t.Errorf("Expected product %v, got %v", 2, item.ProductID)
				}
			}
		}

		// Check if product's stock is reduced
		var product *models.Product
		if db.Preload("StockEntries").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 190 {
			t.Errorf("Expected %v stock, got %v", 190, product.Inventory())
		}

		// Check if inventory account is reduced
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != -3000 {
			t.Errorf("Expected balance %v, got %v", -3000, inv.Balance())
		}

		// Check if revenue account is increased
		var rev *models.Account
		if db.Preload("Transactions").First(&rev, revenue.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if rev.Balance() != 4500 {
			t.Errorf("Expected balance %v, got %v", 4500, rev.Balance())
		}

		// Check if cost of sales is increased
		var cost *models.Account
		if db.Preload("Transactions").First(&cost, cogs.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if cost.Balance() != 3000 {
			t.Errorf("Expected balance %v, got %v", 3000, cost.Balance())
		}

		// Check if payment account is increased
		var payment *models.Account
		if db.Preload("Transactions").First(&payment, cash.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payment.Balance() != 4500 {
			t.Errorf("Expected balance %v, got %v", 4500, payment.Balance())
		}
	})

	t.Run("Create not paid", func(t *testing.T) {
		req := Post(t, "/sales", map[string]interface{}{
			"paid":                  false,
			"customer_id":           1,
			"payment_account_id":    nil,
			"receivable_account_id": receivables.ID,
			"items": []map[string]interface{}{
				{"qty": 10, "price": 200, "product_id": 1},
				{"qty": 10, "price": 250, "product_id": 2},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var sale models.Sale
		if err := json.Unmarshal(w.Body.Bytes(), &sale); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if sale.ID != 2 {
			t.Errorf("Expected ID %v, got %v", 2, sale.ID)
		}

		if sale.Paid {
			t.Error("Should not be paid")
		}

		if sale.CustomerID != 1 {
			t.Errorf("Expected user %v, got %v", 1, sale.CustomerID)
		}

		if len(sale.Items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(sale.Items))
		}

		if sale.Total() != 4500 {
			t.Errorf("Expected total %v, got %v", 4500, sale.Total())
		}

		for idx, item := range sale.Items {
			if idx == 0 {
				if item.Price != 200 {
					t.Errorf("Expected price %v, got %v", 200, item.Price)
				}
				if item.ProductID != 1 {
					t.Errorf("Expected product %v, got %v", 1, item.ProductID)
				}
			}

			if idx == 1 {
				if item.Price != 250 {
					t.Errorf("Expected price %v, got %v", 250, item.Price)
				}
				if item.ProductID != 2 {
					t.Errorf("Expected product %v, got %v", 2, item.ProductID)
				}
			}
		}

		// Check if product's stock is reduced
		var product *models.Product
		if db.Preload("StockEntries").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 180 {
			t.Errorf("Expected %v stock, got %v", 180, product.Inventory())
		}

		// Check if inventory account is reduced
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != -6000 {
			t.Errorf("Expected balance %v, got %v", -6000, inv.Balance())
		}

		// Check if revenue account is increased
		var rev *models.Account
		if db.Preload("Transactions").First(&rev, revenue.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if rev.Balance() != 9000 {
			t.Errorf("Expected balance %v, got %v", 9000, rev.Balance())
		}

		// Check if cost of sales is increased
		var cost *models.Account
		if db.Preload("Transactions").First(&cost, cogs.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if cost.Balance() != 6000 {
			t.Errorf("Expected balance %v, got %v", 6000, cost.Balance())
		}

		// Check if payment account remains the same
		var payment *models.Account
		if db.Preload("Transactions").First(&payment, cash.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if payment.Balance() != 4500 {
			t.Errorf("Expected balance %v, got %v", 4500, payment.Balance())
		}

		// Check if receivable account is increased
		var recv *models.Account
		if db.Preload("Transactions").First(&recv, receivables.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if recv.Balance() != 4500 {
			t.Errorf("Expected balance %v, got %v", 4500, recv.Balance())
		}
	})

	t.Run("Create without enough stock", func(t *testing.T) {
		req := Post(t, "/sales", map[string]interface{}{
			"paid":                  true,
			"customer_id":           1,
			"payment_account_id":    cash.ID,
			"receivable_account_id": nil,
			"items": []map[string]interface{}{
				{"qty": 500, "price": 200, "product_id": 1},
				{"qty": 10, "price": 250, "product_id": 2},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("failed parsing JSON", err)
		}

		if response["error"] != api.ErrNotEnoughStock.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrNotEnoughStock.Error(), response["error"])
		}
	})

	t.Run("Create without items", func(t *testing.T) {
		req := Post(t, "/sales", map[string]interface{}{
			"paid":                  true,
			"customer_id":           1,
			"payment_account_id":    cash.ID,
			"receivable_account_id": nil,
			"items":                 []map[string]interface{}{},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})
}
