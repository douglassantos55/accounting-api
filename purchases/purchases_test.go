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
}
