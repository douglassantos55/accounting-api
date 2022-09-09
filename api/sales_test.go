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
	db.AutoMigrate(&models.StockUsage{})
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

		if sale.PaymentAccountID == nil {
			t.Error("Should have payment account")
		}

		if sale.ReceivableAccountID != nil {
			t.Errorf("Should not have receivable account, got %v", *sale.ReceivableAccountID)
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
		if db.Preload("StockEntries.StockUsages").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 190 {
			t.Errorf("Expected %v stock, got %v", 190, product.Inventory())
		}

		var prod *models.Product
		if db.Preload("StockEntries.StockUsages").First(&prod, 2).Error != nil {
			t.Error("Should retrieve product")
		}

		if prod.Inventory() != 190 {
			t.Errorf("Expected %v stock, got %v", 190, prod.Inventory())
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

		if sale.PaymentAccountID != nil {
			t.Errorf("Should not have payment account, got %v", *sale.PaymentAccountID)
		}

		if sale.ReceivableAccountID == nil {
			t.Error("Should have receivable account")
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
		if db.Preload("StockEntries.StockUsages").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 180 {
			t.Errorf("Expected %v stock, got %v", 180, product.Inventory())
		}

		var prod *models.Product
		if db.Preload("StockEntries.StockUsages").First(&prod, 2).Error != nil {
			t.Error("Should retrieve product")
		}

		if prod.Inventory() != 180 {
			t.Errorf("Expected %v stock, got %v", 180, prod.Inventory())
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

	t.Run("Consider LIFO", func(t *testing.T) {
		db.Create(&models.Company{
			Name:  "Other company",
			Stock: models.LIFO,
		})

		// ID 6
		payment := &models.Account{
			Name:      "Cash & Equivalents",
			Type:      models.Asset,
			CompanyID: 2,
		}
		db.Create(payment)

		// ID 7
		goods := &models.Account{
			Name:      "Goods",
			Type:      models.Asset,
			CompanyID: 2,
		}
		db.Create(goods)

		// ID 8
		income := &models.Account{
			Name:      "Revenue",
			Type:      models.Revenue,
			CompanyID: 2,
		}
		db.Create(income)

		// ID 9
		expenses := &models.Account{
			Name:      "Cost of sales",
			Type:      models.Expense,
			CompanyID: 2,
		}
		db.Create(expenses)

		db.Create(&models.Product{
			Name:                "Product 5",
			Price:               500,
			CompanyID:           2,
			Purchasable:         true,
			RevenueAccountID:    &income.ID,
			CostOfSaleAccountID: &expenses.ID,
			InventoryAccountID:  goods.ID,
			StockEntries: []*models.StockEntry{
				{Qty: 100, Price: 400},
				{Qty: 100, Price: 450},
			},
		})

		db.Create(&models.Customer{
			Name:      "Customer",
			Email:     "customer@email.com",
			CompanyID: 2,
		})

		req := Post(t, "/sales", map[string]interface{}{
			"paid":                  true,
			"customer_id":           2,
			"payment_account_id":    payment.ID,
			"receivable_account_id": nil,
			"items": []map[string]interface{}{
				{"qty": 10, "price": 600, "product_id": 3},
			},
		})

		req.Header.Set("CompanyID", "2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var sale models.Sale
		if err := json.Unmarshal(w.Body.Bytes(), &sale); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if sale.ID != 3 {
			t.Errorf("Expected ID %v, got %v", 3, sale.ID)
		}

		// Check if product's stock is reduced
		var product *models.Product
		if db.Preload("StockEntries.StockUsages").First(&product, 3).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 190 {
			t.Errorf("Expected %v stock, got %v", 190, product.Inventory())
		}

		// Check if inventory account is reduced
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, goods.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != -4500 {
			t.Errorf("Expected balance %v, got %v", -4500, inv.Balance())
		}

		// Check if revenue account is increased
		var rev *models.Account
		if db.Preload("Transactions").First(&rev, income.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if rev.Balance() != 6000 {
			t.Errorf("Expected balance %v, got %v", 6000, rev.Balance())
		}

		// Check if cost of sales is increased
		var cost *models.Account
		if db.Preload("Transactions").First(&cost, expenses.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if cost.Balance() != 4500 {
			t.Errorf("Expected balance %v, got %v", 4500, cost.Balance())
		}

		// Check if payment account is increased
		var pay *models.Account
		if db.Preload("Transactions").First(&pay, payment.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if pay.Balance() != 6000 {
			t.Errorf("Expected balance %v, got %v", 6000, pay.Balance())
		}
	})

	t.Run("List", func(t *testing.T) {
		req := Get(t, "/sales")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var sales []models.Sale
		if err := json.Unmarshal(w.Body.Bytes(), &sales); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(sales) != 2 {
			t.Errorf("Expected %v sales, got %v", 2, len(sales))
		}

		for idx, sale := range sales {
			if len(sale.Items) == 0 {
				t.Error("Expected items to be retrieved along sale")
			}

			if sale.Customer == nil {
				t.Error("Expected customer to be retrieved along sale")
			}

			if idx == 0 {
				if sale.PaymentAccount == nil {
					t.Error("Expected payment account to be retrieved along sale")
				}
				if sale.ReceivableAccount != nil {
					t.Error("Expected receivable account to not be retrieved along sale")
				}
			}

			if idx == 1 {
				if sale.PaymentAccount != nil {
					t.Error("Expected payment account to not be retrieved along sale")
				}
				if sale.ReceivableAccount == nil {
					t.Error("Expected receivable account to be retrieved along sale")
				}
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := Get(t, "/sales/1")

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

		if len(sale.Items) == 0 {
			t.Error("Expected items to be retrieved along sale")
		}

		if sale.Customer == nil {
			t.Error("Expected customer to be retrieved along sale")
		}

		if sale.PaymentAccount == nil {
			t.Error("Expected payment account to be retrieved along sale")
		}

		if sale.ReceivableAccount != nil {
			t.Error("Expected receivable account to not be retrieved along sale")
		}
	})

	t.Run("Get non existent", func(t *testing.T) {
		req := Get(t, "/sales/1222222")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get invalid", func(t *testing.T) {
		req := Get(t, "/sales/asontehu")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get from another company", func(t *testing.T) {
		req := Get(t, "/sales/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete paid", func(t *testing.T) {
		req := Delete(t, "/sales/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}

		if db.First(&models.Sale{}, 1).Error == nil {
			t.Error("Should not find sale")
		}

		// Check if product's stock is reduced
		var product *models.Product
		if db.Preload("StockEntries.StockUsages").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 190 {
			t.Errorf("Expected %v stock, got %v", 190, product.Inventory())
		}

		var prod *models.Product
		if result := db.Preload("StockEntries.StockUsages").First(&prod, 2); result.Error != nil {
			t.Error("Should retrieve product", result.Error)
		}

		if prod.Inventory() != 190 {
			t.Errorf("Expected %v stock, got %v", 190, prod.Inventory())
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

		if payment.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, payment.Balance())
		}
	})

	t.Run("Delete not paid", func(t *testing.T) {
		req := Delete(t, "/sales/2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}

		if db.First(&models.Sale{}, 2).Error == nil {
			t.Error("Should not find sale")
		}

		// Check if product's stock is reduced
		var product *models.Product
		if db.Preload("StockEntries.StockUsages").First(&product, 1).Error != nil {
			t.Error("Should retrieve product")
		}

		if product.Inventory() != 200 {
			t.Errorf("Expected %v stock, got %v", 200, product.Inventory())
		}

		var prod *models.Product
		if db.Preload("StockEntries.StockUsages").First(&prod, 2).Error != nil {
			t.Error("Should retrieve product")
		}

		if prod.Inventory() != 200 {
			t.Errorf("Expected %v stock, got %v", 200, prod.Inventory())
		}

		// Check if inventory account is reduced
		var inv *models.Account
		if db.Preload("Transactions").First(&inv, inventory.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if inv.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, inv.Balance())
		}

		// Check if receivable account is increased
		var recv *models.Account
		if db.Preload("Transactions").First(&recv, receivables.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if recv.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, recv.Balance())
		}

		// Check if cost of sales is increased
		var cost *models.Account
		if db.Preload("Transactions").First(&cost, cogs.ID).Error != nil {
			t.Error("Should retrieve account")
		}

		if cost.Balance() != 0 {
			t.Errorf("Expected balance %v, got %v", 0, cost.Balance())
		}
	})

	t.Run("Delete non existent", func(t *testing.T) {
		req := Delete(t, "/sales/3215")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete invalid", func(t *testing.T) {
		req := Delete(t, "/sales/aosentuh")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete from another company", func(t *testing.T) {
		req := Delete(t, "/sales/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})
}
