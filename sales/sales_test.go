package sales_test

import (
	"testing"

	"example.com/accounting/customers"
	"example.com/accounting/database"
	"example.com/accounting/products"
	"example.com/accounting/sales"
)

func TestSales(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()

	db.Migrate(&sales.Sale{})
	db.Migrate(&sales.Item{})
	db.Migrate(&products.Product{})

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		customer := &customers.Customer{Name: "John Doe"}

		items := []*sales.Item{
			{
				Qty:     1,
				Price:   100,
				Product: &products.Product{Name: "Mouse"},
			},
			{
				Qty:     2,
				Price:   30,
				Product: &products.Product{Name: "Mousepad"},
			},
		}

		sale, err := sales.Create(customer, items)
		if err != nil {
			t.Error(err)
		}

		if sale.ID == 0 {
			t.Error("Should have saved sale")
		}

		if len(sale.Items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(sale.Items))
		}
	})
}
