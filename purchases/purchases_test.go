package purchases_test

import (
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/products"
	"example.com/accounting/purchases"
)

func TestPurchases(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.Migrate(&accounts.Account{})
	db.Migrate(&accounts.Entry{})
	db.Migrate(&accounts.Transaction{})
	db.Migrate(&products.Product{})
	db.Migrate(&products.Purchase{})

	cash, _ := accounts.Create("Cash & Equivalents", accounts.Asset, nil)
	revenue, _ := accounts.Create("Revenue", accounts.Revenue, nil)
	receivable, _ := accounts.Create("Receivables", accounts.Asset, nil)
	inventory, _ := accounts.Create("Inventory", accounts.Asset, nil)

	products.Create(&products.Product{
		Name:                "Prod 1",
		Price:               100,
		Purchasable:         true,
		RevenueAccountID:    &revenue.ID,
		ReceivableAccountID: &receivable.ID,
		InventoryAccountID:  inventory.ID,
	})

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		purchase, err := purchases.Create(1, 5, 155.75, cash.ID)
		if err != nil {
			t.Error(err)
		}
		if purchase.ID == 0 {
			t.Error("Should have saved purchase")
		}
		if purchase.Price != 155.75 {
			t.Errorf("Expected price %v, got %v", 155.75, purchase.Price)
		}
	})

	t.Run("Create without product", func(t *testing.T) {
		if _, err := purchases.Create(0, 5, 15.33, cash.ID); err == nil {
			t.Error("Should not save without product")
		}
	})

	t.Run("Create stock entry", func(t *testing.T) {
		products.Create(&products.Product{
			Name:               "Prod 2",
			Price:              16,
			InventoryAccountID: inventory.ID,
		})

		purchases.Create(2, 4, 153.22, cash.ID)
		purchases.Create(2, 4, 163.22, cash.ID)
		purchases.Create(2, 10, 157.11, cash.ID)

		var product *products.Product
		if err := products.Find(2).With("StockEntries").First(&product); err != nil {
			t.Error(err)
		}

		if product.Inventory() != 18 {
			t.Errorf("Expected %v stock, got %v", 18, product.Inventory())
		}
	})

	t.Run("List", func(t *testing.T) {
		result, err := purchases.List()
		if err != nil {
			t.Error(err)
		}

		var items []*products.Purchase
		if err := result.Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 4 {
			t.Errorf("Expected %v purchases, got %v", 4, len(items))
		}
	})

	t.Run("List with product", func(t *testing.T) {
		result, err := purchases.List()
		if err != nil {
			t.Error(err)
		}

		var items []*products.Purchase
		if err := result.With("Product").Get(&items); err != nil {
			t.Error(err)
		}

		for _, item := range items {
			if item.Product == nil {
				t.Error("Should have product")
			}
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		result, err := purchases.Find(4)
		if err != nil {
			t.Error(err)
		}

		var purchase *products.Purchase
		if err := result.First(&purchase); err != nil {
			t.Error(err)
		}

		if purchase.ID == 0 {
			t.Error("Should retrieve purchase")
		}
		if purchase.ProductID != 2 {
			t.Errorf("Expected product %v, got %v", 2, purchase.ProductID)
		}
		if purchase.Qty != 10 {
			t.Errorf("Expected qty %v, got %v", 10, purchase.Qty)
		}
	})

	t.Run("List with condition", func(t *testing.T) {
		result, err := purchases.List()
		if err != nil {
			t.Error(err)
		}
		var items []*products.Purchase
		if err := result.Where("Qty > ?", 5).Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 1 {
			t.Errorf("Expected %v items, got %v", 1, len(items))
		}
	})

	t.Run("Update", func(t *testing.T) {
		result, err := purchases.Find(4)
		if err != nil {
			t.Error(err)
		}

		var purchase *products.Purchase
		if err := result.First(&purchase); err != nil {
			t.Error(err)
		}

		prevUpdate := purchase.UpdatedAt
		prevProduct := purchase.ProductID
		prevPrice := purchase.Price
		prevQty := purchase.Qty

		purchase.ProductID = 1
		purchase.Price = 355
		purchase.Qty = 55

		if err := purchases.Update(purchase); err != nil {
			t.Error(err)
		}

		result, _ = purchases.Find(4)
		result.First(&purchase)

		if prevUpdate == purchase.UpdatedAt {
			t.Error("Should have updated")
		}

		if prevProduct == purchase.ProductID {
			t.Errorf("Expected product %v, got %v", 1, purchase.ProductID)
		}

		if prevPrice == purchase.Price {
			t.Error("Should have updated price")
		}

		if prevQty == purchase.Qty {
			t.Error("Should have updated qty")
		}
	})

	t.Run("Update without product", func(t *testing.T) {
		result, err := purchases.Find(4)
		if err != nil {
			t.Error(err)
		}

		var purchase *products.Purchase
		if err := result.First(&purchase); err != nil {
			t.Error(err)
		}

		purchase.ProductID = 1110

		if err := purchases.Update(purchase); err == nil {
			t.Error("Should not update without product")
		}
	})

	t.Run("Updates stock entry", func(t *testing.T) {
		result, err := purchases.Find(4)
		if err != nil {
			t.Error(err)
		}

		var purchase *products.Purchase
		if err := result.First(&purchase); err != nil {
			t.Error(err)
		}

		purchase.Qty = 8
		purchase.ProductID = 2

		if err := purchases.Update(purchase); err != nil {
			t.Error(err)
		}

		var prod *products.Product
		products.Find(2).With("StockEntries").First(&prod)

		if prod.Inventory() != 16 {
			t.Errorf("Expected stock %v , got %v", 16, prod.Inventory())
		}

		var prod1 *products.Product
		products.Find(1).With("StockEntries").First(&prod1)

		if prod1.Inventory() != 5 {
			t.Errorf("Expected stock %v , got %v", 5, prod1.Inventory())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := purchases.Delete(4); err != nil {
			t.Error(err)
		}

		result, err := purchases.Find(4)
		if err != nil {
			t.Error(err)
		}

		var purchase *products.Purchase
		if err := result.First(&purchase); err == nil {
			t.Error("Should have deleted purchase")
		}

		var prod *products.Product
		products.Find(2).With("StockEntries").First(&prod)

		if prod.Inventory() != 8 {
			t.Errorf("Expected stock %v , got %v", 8, prod.Inventory())
		}
	})

	t.Run("Increase inventory account, reduce payment account", func(t *testing.T) {
		account, err := accounts.Create("Inv", accounts.Asset, nil)
		payment, err := accounts.Create("Cash", accounts.Asset, nil)

		if err != nil {
			t.Error(err)
		}

		products.Create(&products.Product{
			Name:               "Mice",
			Price:              33.5,
			InventoryAccountID: account.ID,
		})

		if _, err := purchases.Create(3, 100, 26.5, payment.ID); err != nil {
			t.Error(err)
		}

		var inventory *accounts.Account
		if err := accounts.Find(account.ID).With("Transactions").First(&inventory); err != nil {
			t.Error(err)
		}

		if inventory.Balance() != 2650.0 {
			t.Errorf("Expected balance %v, got %v", 2650.0, inventory.Balance())
		}

		if err := accounts.Find(payment.ID).With("Transactions").First(&payment); err != nil {
			t.Error(err)
		}

		if payment.Balance() != -2650.0 {
			t.Errorf("Expected balance %v, got %v", -2650.0, payment.Balance())
		}
	})
}
