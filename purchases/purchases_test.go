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
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()

	db.Migrate(&products.Product{})
	db.Migrate(&purchases.Purchase{})

	accounts.Create("Revenue", accounts.Revenue, nil)
	products.Create("Prod 1", 100, 322, 1, nil)

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		purchase, err := purchases.Create(1, 5)
		if err != nil {
			t.Error(err)
		}
		if purchase.ID == 0 {
			t.Error("Should have saved purchase")
		}
	})

	t.Run("Create without product", func(t *testing.T) {
		if _, err := purchases.Create(0, 5); err == nil {
			t.Error("Should not save without product")
		}
	})

	t.Run("Increase product stock", func(t *testing.T) {
		products.Create("Prod 2", 16, 8, 1, nil)

		_, err := purchases.Create(2, 10)
		if err != nil {
			t.Error(err)
		}

		var product *products.Product
		if err := products.Find(2).First(&product); err != nil {
			t.Error(err)
		}

		if product.Stock != 18 {
			t.Errorf("Expected %v stock, got %v", 18, product.Stock)
		}
	})

	t.Run("List", func(t *testing.T) {
		result, err := purchases.List()
		if err != nil {
			t.Error(err)
		}

		var items []*purchases.Purchase
		if err := result.Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v purchases, got %v", 2, len(items))
		}
	})

	t.Run("List with product", func(t *testing.T) {
		result, err := purchases.List()
		if err != nil {
			t.Error(err)
		}

		var items []*purchases.Purchase
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
		result, err := purchases.Find(2)
		if err != nil {
			t.Error(err)
		}

		var purchase *purchases.Purchase
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
		var items []*purchases.Purchase
		if err := result.Where("Qty > ?", 5).Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 1 {
			t.Errorf("Expected %v items, got %v", 1, len(items))
		}
	})
}
