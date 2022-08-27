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

	db.Migrate(&sales.Item{})
	db.Migrate(&sales.Sale{})
	db.Migrate(&models.Account{})
	db.Migrate(&models.Entry{})
	db.Migrate(&models.Transaction{})
	db.Migrate(&models.Purchase{})
	db.Migrate(&models.Product{})
	db.Migrate(&models.StockEntry{})

	revenue, _ := accounts.Create("Revenue", models.Revenue, nil)
	costOfSales, _ := accounts.Create("Cost of Sales", models.Expense, nil)

	cash, _ := accounts.Create("Cash", models.Asset, nil)
	inventory, _ := accounts.Create("Inventory", models.Asset, nil)

	events.Handle(events.PurchaseCreated, purchases.CreateStockEntry)
	events.Handle(events.PurchaseCreated, purchases.CreateAccountingEntry)

	events.Handle(events.SaleCreated, sales.CreateAccountingEntry)
	events.Handle(events.SaleCreated, sales.ReduceProductStock)

	t.Run("Create", func(t *testing.T) {
		customer := &models.Customer{Name: "John Doe"}

		items := []*sales.Item{
			{
				Qty:   1,
				Price: 100,
				Product: &models.Product{
					Name:                "Mouse",
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
					RevenueAccountID:    &revenue.ID,
					CostOfSaleAccountID: &costOfSales.ID,
					InventoryAccountID:  inventory.ID,
					StockEntries: []*models.StockEntry{
						{Qty: 100, Price: 29.5},
					}},
			},
		}

		sale := &sales.Sale{
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
		items := []*sales.Item{
			{
				Qty:   1,
				Price: 100,
				Product: &models.Product{
					Name:                "Mouse",
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
					RevenueAccountID:    &revenue.ID,
					CostOfSaleAccountID: &costOfSales.ID,
					InventoryAccountID:  inventory.ID,
				},
			},
		}

		err := sales.Create(&sales.Sale{
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
		err := sales.Create(&sales.Sale{
			Customer:         &models.Customer{},
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
		err := sales.Create(&sales.Sale{
			Paid:             true,
			PaymentAccountID: &cash.ID,
			Customer:         &models.Customer{},
			Items: []*sales.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
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
		err := sales.Create(&sales.Sale{
			Paid:     true,
			Customer: &models.Customer{},
			Items: []*sales.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
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
		err := sales.Create(&sales.Sale{
			Paid:             false,
			PaymentAccountID: &cash.ID,
			Customer:         &models.Customer{},
			Items: []*sales.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
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
		sales.Create(&sales.Sale{
			Paid:             true,
			PaymentAccountID: &cash.ID,
			Customer:         &models.Customer{Name: "Jane Doe"},
			Items: []*sales.Item{
				{
					Qty:   1,
					Price: 100,
					Product: &models.Product{
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

		var items []*sales.Sale
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
		var items []*sales.Sale

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
		var sale *sales.Sale
		if err := sales.Find(2).First(&sale); err != nil {
			t.Error(err)
		}

		if sale.ID == 0 {
			t.Error("should get sale")
		}
	})

	t.Run("Get with Customer and Items", func(t *testing.T) {
		var sale *sales.Sale
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

		var sale *sales.Sale
		if err := sales.Find(2).First(&sale); err == nil {
			t.Error("Should have deleted sale")
		}
	})

	t.Run("Delete items", func(t *testing.T) {
		if err := sales.Delete(1); err != nil {
			t.Error(err)
		}

		var sale *sales.Sale
		if err := sales.Find(1).First(&sale); err == nil {
			t.Error("Should have deleted sale")
		}

		var items []*sales.Item
		if err := db.Find(&sales.Item{}).Get(&items); err != nil {
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
			RevenueAccountID:    &revenue.ID,
			CostOfSaleAccountID: &costOfSales.ID,
			InventoryAccountID:  inventory.ID,
		}

		if err := products.Create(prod); err != nil {
			t.Error(err)
		}

		if _, err := purchases.Create(&models.Purchase{
			ProductID:        prod.ID,
			Qty:              36,
			Paid:             true,
			Price:            55.3,
			PaymentAccountID: &cash.ID,
		}); err != nil {
			t.Error(err)
		}
		if _, err := purchases.Create(&models.Purchase{
			ProductID:        prod.ID,
			Qty:              34,
			Paid:             true,
			Price:            55.3,
			PaymentAccountID: &cash.ID,
		}); err != nil {
			t.Error(err)
		}

		customer, err := customers.Create("Customer", "", "", "", nil)

		if err != nil {
			t.Error(err)
		}

		if err := sales.Create(&sales.Sale{
			Paid:             true,
			PaymentAccountID: &cash.ID,
			Customer:         customer,
			Items:            []*sales.Item{{Qty: 50, ProductID: prod.ID}},
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
		cogs, _ := accounts.Create("COGS", models.Expense, nil)
		rev, _ := accounts.Create("Revenue", models.Revenue, nil)

		receivable, _ := accounts.Create("Receivable", models.Asset, nil)
		payable, _ := accounts.Create("Payable", models.Liability, nil)

		invAccount1, _ := accounts.Create("Inventory 1", models.Asset, nil)
		invAccount2, _ := accounts.Create("Inventory 2", models.Asset, nil)

		prod1 := &models.Product{
			Price:               25,
			Name:                "Product 1",
			CostOfSaleAccountID: &cogs.ID,
			RevenueAccountID:    &rev.ID,
			InventoryAccountID:  invAccount1.ID,
		}

		products.Create(prod1)

		prod2 := &models.Product{
			Price:               50,
			Name:                "Product 2",
			CostOfSaleAccountID: &cogs.ID,
			RevenueAccountID:    &rev.ID,
			InventoryAccountID:  invAccount2.ID,
		}

		products.Create(prod2)

		purchases.Create(&models.Purchase{
			ProductID:        prod1.ID,
			Qty:              20,
			Paid:             true,
			Price:            25,
			PaymentAccountID: &cash.ID,
		})

		purchases.Create(&models.Purchase{
			ProductID:        prod2.ID,
			Qty:              20,
			Paid:             false,
			Price:            50,
			PayableAccountID: &payable.ID,
		})

		if err := sales.Create(&sales.Sale{
			Paid:                false,
			ReceivableAccountID: &receivable.ID,
			Customer:            &models.Customer{Name: "TNT"},
			Items: []*sales.Item{
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

	t.Run("Reduces inventory accounts 2", func(t *testing.T) {
		cogs, _ := accounts.Create("COGS", models.Expense, nil)
		rev, _ := accounts.Create("Revenue", models.Revenue, nil)

		payment, _ := accounts.Create("Payment", models.Asset, nil)
		payable, _ := accounts.Create("Payable", models.Liability, nil)
		invAccount, _ := accounts.Create("Inventory", models.Asset, nil)

		prod := &models.Product{
			Price:               25,
			Name:                "Product",
			CostOfSaleAccountID: &cogs.ID,
			RevenueAccountID:    &rev.ID,
			InventoryAccountID:  invAccount.ID,
		}

		products.Create(prod)

		purchases.Create(&models.Purchase{
			ProductID:        prod.ID,
			Qty:              20,
			Paid:             true,
			Price:            25,
			PaymentAccountID: &payment.ID, // -500
		})

		purchases.Create(&models.Purchase{
			ProductID:        prod.ID,
			Qty:              20,
			Paid:             false,
			Price:            30,
			PayableAccountID: &payable.ID, // 600
		})

		if err := sales.Create(&sales.Sale{
			Paid:             true,
			PaymentAccountID: &payment.ID, // 750 - 500 = 250
			Customer:         &models.Customer{Name: "TNT"},
			Items: []*sales.Item{
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
