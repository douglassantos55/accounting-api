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

	accounts.Create("Revenue", models.Revenue, nil)
	cash, _ := accounts.Create("Cash", models.Asset, nil)
	inventory, _ := accounts.Create("Inventory", models.Asset, nil)

	events.Handle(events.SaleCreated, sales.ReduceProductStock)
	events.Handle(events.PurchaseCreated, purchases.CreateStockEntry)

	t.Run("Create", func(t *testing.T) {
		customer := &models.Customer{Name: "John Doe"}

		items := []*sales.Item{
			{
				Qty:   1,
				Price: 100,
				Product: &models.Product{
					Name:               "Mouse",
					InventoryAccountID: inventory.ID,
					StockEntries: []*models.StockEntry{
						{Qty: 100, Price: 99.3},
					}},
			},
			{
				Qty:   2,
				Price: 30,
				Product: &models.Product{
					Name:               "Mousepad",
					InventoryAccountID: inventory.ID,
					StockEntries: []*models.StockEntry{
						{Qty: 100, Price: 29.5},
					}},
			},
		}

		sale, err := sales.Create(customer, items)
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
					Name:               "Mouse",
					InventoryAccountID: inventory.ID,
				},
			},
			{
				Qty:   2,
				Price: 30,
				Product: &models.Product{
					Name:               "Mousepad",
					InventoryAccountID: inventory.ID,
				},
			},
		}

		_, err := sales.Create(nil, items)
		if err == nil {
			t.Error("Should not create without customer")
		}
		if !errors.Is(err, sales.ErrCustomerMissing) {
			t.Errorf("Should return ErrCustomerMissing, got %v", err)
		}
	})

	t.Run("Create without items", func(t *testing.T) {
		_, err := sales.Create(&models.Customer{}, nil)
		if err == nil {
			t.Error("Should not create without items")
		}
		if !errors.Is(err, sales.ErrItemsMissing) {
			t.Errorf("Should return ErrItemsMissing, got %v", err)
		}
	})

	t.Run("Create without stock", func(t *testing.T) {
		_, err := sales.Create(&models.Customer{}, []*sales.Item{
			{
				Qty:   1,
				Price: 100,
				Product: &models.Product{
					Name:               "Mouse",
					InventoryAccountID: inventory.ID,
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

	t.Run("List", func(t *testing.T) {
		sales.Create(&models.Customer{Name: "Jane Doe"}, []*sales.Item{
			{
				Qty:   1,
				Price: 100,
				Product: &models.Product{
					Name:               "Mouse",
					InventoryAccountID: inventory.ID,
					StockEntries: []*models.StockEntry{
						{Qty: 11, Price: 101.37},
					}},
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
			Name:               "Prod",
			Price:              15,
			InventoryAccountID: inventory.ID,
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

		if _, err := sales.Create(customer, []*sales.Item{{Qty: 50, ProductID: prod.ID}}); err != nil {
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
}
