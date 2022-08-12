package products_test

import (
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/products"
)

func TestProducts(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()

	db.Migrate(&products.Product{})
	db.Migrate(&accounts.Account{})

	accounts.Create("Revenue", accounts.Revenue, 0)

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		prod, err := products.Create("Keyboard", 350.5, 1)

		if err != nil {
			t.Error(err)
		}

		if prod.ID != 1 {
			t.Errorf("Expected ID 1, got %v", prod.ID)
		}

		if prod.Price != 350.5 {
			t.Errorf("Expected price 350.5, got %v", prod.Price)
		}

		if prod.AccountID != 1 {
			t.Errorf("Expected AccountID 1, got %v", prod.AccountID)
		}
	})

	t.Run("Create Without Account", func(t *testing.T) {
		_, err := products.Create("Coffee Powder", 33.6, 0)

		if err == nil {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("Create With Non Existing Account", func(t *testing.T) {
		_, err := products.Create("Coffee Powder", 33.6, 10)

		if err == nil {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("List", func(t *testing.T) {
		products.Create("Monitor", 1350.5, 1)
		products.Create("Mouse", 150.5, 1)

		var items []*products.Product
		err := products.List().Get(&items)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("Expected %v items, got %v", 3, len(items))
		}
	})

	t.Run("List With Account", func(t *testing.T) {
		var items []*products.Product
		err := products.List().With("Account").Get(&items)

		if err != nil {
			t.Error(err)
		}

		for _, product := range items {
			if product.Account == nil {
				t.Error("Should have revenue account")
			}
		}
	})

	t.Run("Get By ID", func(t *testing.T) {
		var product *products.Product

		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.Name != "Mouse" {
			t.Errorf("Expected name %v, got %v", "Mouse", product.Name)
		}

		if product.Price != 150.5 {
			t.Errorf("Expected Price %v, got %v", 150.5, product.Price)
		}
	})

	t.Run("Get With Account", func(t *testing.T) {
		var product *products.Product

		if err := products.Find(3).With("Account").First(&product); err != nil {
			t.Error(err)
		}

		if product.Account == nil {
			t.Error("Should have Account")
		}
	})
}
