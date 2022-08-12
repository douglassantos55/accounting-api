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
}
