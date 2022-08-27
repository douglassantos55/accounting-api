package sales_test

import (
	"errors"
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/customers"
	"example.com/accounting/database"
	"example.com/accounting/events"
	"example.com/accounting/models"
	"example.com/accounting/products"
	"example.com/accounting/purchases"
	"example.com/accounting/sales"
)

func TestSales(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.Migrate(&models.Item{})
	db.Migrate(&models.Sale{})
	db.Migrate(&models.Account{})
	db.Migrate(&models.Entry{})
	db.Migrate(&models.Transaction{})
	db.Migrate(&models.Purchase{})
	db.Migrate(&models.Product{})
	db.Migrate(&models.StockEntry{})

	db.Create(&models.Company{
		Name: "Testing Company",
	})

	revenue, _ := accounts.Create(1, "Revenue", models.Revenue, nil)
	costOfSales, _ := accounts.Create(1, "Cost of Sales", models.Expense, nil)

	cash, _ := accounts.Create(1, "Cash", models.Asset, nil)
	inventory, _ := accounts.Create(1, "Inventory", models.Asset, nil)

	events.Handle(events.PurchaseCreated, purchases.CreateStockEntry)
	events.Handle(events.PurchaseCreated, purchases.CreateAccountingEntry)

	events.Handle(events.SaleCreated, sales.CreateAccountingEntry)
	events.Handle(events.SaleCreated, sales.ReduceProductStock)

	t.Run("Create", func(t *testing.T) {
		customer := &models.Customer{Name: "John Doe", CompanyID: 1}

		items := []*models.Item{
			{
				Qty:   1,
				Price: 100,
				Product: &models.Product{
					Name:                "Mouse",
					CompanyID:           1,
					RevenueAccountID:    &revenue.ID,
					CostOfSaleAccountID: &costOfSales.ID,
					InventoryAccountID:  inventory.ID,
					StockEntries: []*models.StockEntry{
						{Qty: 100, Price: 99.3},
					}},
			},
			{
				Qty:   2,
				Price: 30,
				Product: &models.Product{
					Name:                "Mousepad",
					CompanyID:           1,
					RevenueAccountID:    &revenue.ID,
					CostOfSaleAccountID: &costOfSales.ID,
					InventoryAccountID:  inventory.ID,
					StockEntries: []*models.StockEntry{
						{Qty: 100, Price: 29.5},
					}},
			},
		}

		sale := &models.Sale{
			CompanyID:        1,
			Items:            items,
			Paid:             true,
			PaymentAccountID: &cash.ID,
			Customer:         customer,
		}

		err := sales.Create(sale)

		if err != nil {
			t.Error(err)
		} else {
			if sale.ID == 0 {
				t.Error("Should have saved sale")
			}

			if len(sale.Items) != 2 {
				t.Errorf("Expected %v items, got %v", 2, len(sale.Items))
			}
		}
	})

	t.Run("Create without customer", func(t *testing.T) {
		items := []*models.Item{
			{
				Qty:   1,
				Price: 100,
				Product: &models.Product{
					Name:                "Mouse",
					CompanyID:           1,
					RevenueAccountID:    &revenue.ID,
					CostOfSaleAccountID: &costOfSales.ID,
					InventoryAccountID:  inventory.ID,
				},
			},
			{
				Qty:   2,
				Price: 30,
				Product: &models.Product{
					Name:                "Mousepad",
					CompanyID:           1,
					RevenueAccountID:    &revenue.ID,
					CostOfSaleAccountID: &costOfSales.ID,
					InventoryAccountID:  inventory.ID,
				},
			},
		}

		err := sales.Create(&models.Sale{
			CompanyID:        1,
			Items:            items,
			Paid:             true,
			PaymentAccountID: &cash.ID,
		})

		if err == nil {
			t.Error("Should not create without customer")
		}

		if !errors.Is(err, sales.ErrCustomerMissing) {
			t.Errorf("Should return ErrCustomerMissing, got %v", err)
		}
	})

	t.Run("Create without items", func(t *testing.T) {
		err := sales.Create(&models.Sale{
			CompanyID:        1,
			Customer:         &models.Customer{CompanyID: 1},
			Paid:             true,
			PaymentAccountID: &cash.ID,
		})

		if err == nil {
			t.Error("Should not create without items")
		}

		if !errors.Is(err, sales.ErrItemsMissing) {
			t.Errorf("Should return ErrItemsMissing, got %v", err)
		}
	})

	t.Run("Create without stock", func(t *testing.T) {
		err := sales.Create(&models.Sale{
			CompanyID:        1,
			Paid:             true,
			PaymentAccountID: &cash.ID,
			Customer:         &models.Customer{CompanyID: 1},
			Items: []*models.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
						CompanyID:           1,
						Name:                "Mouse",
						RevenueAccountID:    &revenue.ID,
						CostOfSaleAccountID: &costOfSales.ID,
						InventoryAccountID:  inventory.ID,
					},
				},
			},
		})

		if err == nil {
			t.Error("Should not create without stock")
		}

		if !errors.Is(err, sales.ErrNotEnoughStock) {
			t.Errorf("Should return ErrNotEnoughStock, got %v", err)
		}
	})

	t.Run("Create without payment account", func(t *testing.T) {
		err := sales.Create(&models.Sale{
			Paid:      true,
			CompanyID: 1,
			Customer:  &models.Customer{CompanyID: 1},
			Items: []*models.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
						CompanyID:           1,
						Name:                "Mouse",
						RevenueAccountID:    &revenue.ID,
						CostOfSaleAccountID: &costOfSales.ID,
						InventoryAccountID:  inventory.ID,
						StockEntries: []*models.StockEntry{
							{Qty: 100, Price: 99.3},
						},
					},
				},
			},
		})

		if err == nil {
			t.Error("Should not create without payment account")
		}

		if !errors.Is(err, sales.ErrPaymentAccountMissing) {
			t.Errorf("Should return ErrPaymentAccountMissing, got %v", err)
		}
	})

	t.Run("Create without receivable account", func(t *testing.T) {
		err := sales.Create(&models.Sale{
			CompanyID:        1,
			Paid:             false,
			PaymentAccountID: &cash.ID,
			Customer:         &models.Customer{CompanyID: 1},
			Items: []*models.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
						CompanyID:           1,
						Name:                "Mouse",
						RevenueAccountID:    &revenue.ID,
						CostOfSaleAccountID: &costOfSales.ID,
						InventoryAccountID:  inventory.ID,
						StockEntries: []*models.StockEntry{
							{Qty: 100, Price: 99.3},
						},
					},
				},
			},
		})

		if err == nil {
			t.Error("Should not create without payment account")
		}

		if !errors.Is(err, sales.ErrReceivableAccountMissing) {
			t.Errorf("Should return ErrReceivableAccountMissing, got %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		sales.Create(&models.Sale{
			Paid:             true,
			CompanyID:        1,
			PaymentAccountID: &cash.ID,
			Customer:         &models.Customer{Name: "Jane Doe", CompanyID: 1},
			Items: []*models.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
						CompanyID:           1,
						Name:                "Mouse",
						RevenueAccountID:    &revenue.ID,
						InventoryAccountID:  inventory.ID,
						CostOfSaleAccountID: &costOfSales.ID,
						StockEntries: []*models.StockEntry{
							{Qty: 11, Price: 101.37},
						}},
				},
			},
		})

		var items []*models.Sale
		if err := sales.List().Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v item, got %v", 2, len(items))
		}

		for _, item := range items {
			if item.Customer != nil {
				t.Error("Should not have customer")
			}
			if len(item.Items) != 0 {
				t.Error("Should not have items")
			}
		}
	})

	t.Run("List with Customer and Items", func(t *testing.T) {
		var items []*models.Sale

		if err := sales.List().With("Customer", "Items").Get(&items); err != nil {
			t.Error(err)
		}

		for _, item := range items {
			if item.Customer == nil {
				t.Error("Should have customer")
			}
			if len(item.Items) == 0 {
				t.Error("Should have items")
			}
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		var sale *models.Sale
		if err := sales.Find(2).First(&sale); err != nil {
			t.Error(err)
		}

		if sale.ID == 0 {
			t.Error("should get sale")
		}
	})

	t.Run("Get with Customer and Items", func(t *testing.T) {
		var sale *models.Sale
		if err := sales.Find(2).With("Customer", "Items").First(&sale); err != nil {
			t.Error(err)
		}

		if sale.Customer.Name != "Jane Doe" {
			t.Errorf("Expected name %v, got %v", "Jane Doe", sale.Customer.Name)
		}
		if len(sale.Items) != 1 {
			t.Error("Expected items")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := sales.Delete(2); err != nil {
			t.Error(err)
		}

		var sale *models.Sale
		if err := sales.Find(2).First(&sale); err == nil {
			t.Error("Should have deleted sale")
		}
	})

	t.Run("Delete items", func(t *testing.T) {
		if err := sales.Delete(1); err != nil {
			t.Error(err)
		}

		var sale *models.Sale
		if err := sales.Find(1).First(&sale); err == nil {
			t.Error("Should have deleted sale")
		}

		var items []*models.Item
		if err := db.Find(&models.Item{}).Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 0 {
			t.Errorf("Should also delete the items, got %v", len(items))
		}
	})

	t.Run("Reduces product stock", func(t *testing.T) {
		prod := &models.Product{
			Name:                "Prod",
			Price:               15,
			CompanyID:           1,
			RevenueAccountID:    &revenue.ID,
			CostOfSaleAccountID: &costOfSales.ID,
			InventoryAccountID:  inventory.ID,
		}

		if err := products.Create(prod); err != nil {
			t.Error(err)
		}

		if _, err := purchases.Create(&models.Purchase{
			CompanyID:        1,
			ProductID:        prod.ID,
			Qty:              36,
			Paid:             true,
			Price:            55.3,
			PaymentAccountID: &cash.ID,
		}); err != nil {
			t.Error(err)
		}
		if _, err := purchases.Create(&models.Purchase{
			CompanyID:        1,
			ProductID:        prod.ID,
			Qty:              34,
			Paid:             true,
			Price:            55.3,
			PaymentAccountID: &cash.ID,
		}); err != nil {
			t.Error(err)
		}

		customer, err := customers.Create(1, "Customer", "", "", "", nil)

		if err != nil {
			t.Error(err)
		}

		if err := sales.Create(&models.Sale{
			Paid:             true,
			CompanyID:        1,
			PaymentAccountID: &cash.ID,
			Customer:         customer,
			Items:            []*models.Item{{Qty: 50, ProductID: prod.ID}},
		}); err != nil {
			t.Error(err)
		}

		products.Find(prod.ID).With("StockEntries").First(&prod)

		if len(prod.StockEntries) != 1 {
			t.Errorf("Expected %v entry, got %v", 1, len(prod.StockEntries))
		}

		if prod.Inventory() != 20 {
			t.Errorf("Expected stock %v, got %v", 20, prod.Inventory())
		}
	})

	t.Run("Reduces inventory accounts", func(t *testing.T) {
		cogs, _ := accounts.Create(1, "COGS", models.Expense, nil)
		rev, _ := accounts.Create(1, "Revenue", models.Revenue, nil)

		receivable, _ := accounts.Create(1, "Receivable", models.Asset, nil)
		payable, _ := accounts.Create(1, "Payable", models.Liability, nil)

		invAccount1, _ := accounts.Create(1, "Inventory 1", models.Asset, nil)
		invAccount2, _ := accounts.Create(1, "Inventory 2", models.Asset, nil)

		prod1 := &models.Product{
			Price:               25,
			CompanyID:           1,
			Name:                "Product 1",
			CostOfSaleAccountID: &cogs.ID,
			RevenueAccountID:    &rev.ID,
			InventoryAccountID:  invAccount1.ID,
		}

		products.Create(prod1)

		prod2 := &models.Product{
			Price:               50,
			CompanyID:           1,
			Name:                "Product 2",
			CostOfSaleAccountID: &cogs.ID,
			RevenueAccountID:    &rev.ID,
			InventoryAccountID:  invAccount2.ID,
		}

		products.Create(prod2)

		purchases.Create(&models.Purchase{
			CompanyID:        1,
			ProductID:        prod1.ID,
			Qty:              20,
			Paid:             true,
			Price:            25,
			PaymentAccountID: &cash.ID,
		})

		purchases.Create(&models.Purchase{
			CompanyID:        1,
			ProductID:        prod2.ID,
			Qty:              20,
			Paid:             false,
			Price:            50,
			PayableAccountID: &payable.ID,
		})

		if err := sales.Create(&models.Sale{
			Paid:                false,
			CompanyID:           1,
			ReceivableAccountID: &receivable.ID,
			Customer:            &models.Customer{Name: "TNT", CompanyID: 1},
			Items: []*models.Item{
				{
					Qty:       10,
					Price:     25,
					ProductID: prod1.ID,
				},
				{
					Qty:       10,
					Price:     55,
					ProductID: prod2.ID,
				},
			},
		}); err != nil {
			t.Error(err)
		}

		products.Find(prod1.ID).With("StockEntries").First(&prod1)
		if prod1.Inventory() != 10 {
			t.Errorf("Expected stock %v, got %v", 10, prod1.Inventory())
		}

		if err := accounts.Find(invAccount1.ID).With("Transactions").First(&invAccount1); err != nil {
			t.Error(err)
		}
		if invAccount1.Balance() != 250 {
			t.Errorf("Expected balance %v, got %v", 250, invAccount1.Balance())
		}

		products.Find(prod2.ID).With("StockEntries").First(&prod2)
		if prod2.Inventory() != 10 {
			t.Errorf("Expected stock %v, got %v", 10, prod2.Inventory())
		}

		accounts.Find(invAccount2.ID).With("Transactions").First(&invAccount2)
		if invAccount2.Balance() != 500 {
			t.Errorf("Expected balance %v, got %v", 500, invAccount2.Balance())
		}

	})

	t.Run("Increases revenue/expenses accounts", func(t *testing.T) {
		cogs, _ := accounts.Create(1, "COGS", models.Expense, nil)
		rev, _ := accounts.Create(1, "Revenue", models.Revenue, nil)

		payment, _ := accounts.Create(1, "Payment", models.Asset, nil)
		payable, _ := accounts.Create(1, "Payable", models.Liability, nil)
		invAccount, _ := accounts.Create(1, "Inventory", models.Asset, nil)

		prod := &models.Product{
			Price:               25,
			Name:                "Product",
			CompanyID:           1,
			CostOfSaleAccountID: &cogs.ID,
			RevenueAccountID:    &rev.ID,
			InventoryAccountID:  invAccount.ID,
		}

		products.Create(prod)

		purchases.Create(&models.Purchase{
			CompanyID:        1,
			ProductID:        prod.ID,
			Qty:              20,
			Paid:             true,
			Price:            25,
			PaymentAccountID: &payment.ID, // -500
		})

		purchases.Create(&models.Purchase{
			CompanyID:        1,
			ProductID:        prod.ID,
			Qty:              20,
			Paid:             false,
			Price:            30,
			PayableAccountID: &payable.ID, // 600
		})

		if err := sales.Create(&models.Sale{
			CompanyID:        1,
			Paid:             true,
			PaymentAccountID: &payment.ID, // 750 - 500 = 250
			Customer:         &models.Customer{Name: "TNT", CompanyID: 1},
			Items: []*models.Item{
				{
					Qty:       30,
					Price:     25,
					ProductID: prod.ID,
				},
			},
		}); err != nil {
			t.Error(err)
		}

		products.Find(prod.ID).With("StockEntries").First(&prod)
		if prod.Inventory() != 10 {
			t.Errorf("Expected stock %v, got %v", 10, prod.Inventory())
		}

		if err := accounts.Find(invAccount.ID).With("Transactions").First(&invAccount); err != nil {
			t.Error(err)
		}
		if invAccount.Balance() != 300 {
			t.Errorf("Expected balance %v, got %v", 300, invAccount.Balance())
		}

		accounts.Find(cogs.ID).With("Transactions").First(&cogs)
		if cogs.Balance() != 800 {
			t.Errorf("Expected balance %v, got %v", 800, cogs.Balance())
		}

		accounts.Find(rev.ID).With("Transactions").First(&rev)
		if rev.Balance() != 750 {
			t.Errorf("Expected balance %v, got %v", 750, rev.Balance())
		}

		accounts.Find(payment.ID).With("Transactions").First(&payment)
		if payment.Balance() != 250 {
			t.Errorf("Expected balance %v, got %v", 250, payment.Balance())
		}

		accounts.Find(payable.ID).With("Transactions").First(&payable)
		if payable.Balance() != 600 {
			t.Errorf("Expected balance %v, got %v", 600, payable.Balance())
		}
	})
}
