package sales_test

import (
	"errors"
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

	db.Migrate(&sales.Item{})
	db.Migrate(&sales.Sale{})
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

	t.Run("Create without customer", func(t *testing.T) {
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

		_, err := sales.Create(nil, items)
		if err == nil {
			t.Error("Should not create without customer")
		}
		if !errors.Is(err, sales.ErrCustomerMissing) {
			t.Errorf("Should return ErrCustomerMissing, got %v", err)
		}
	})

	t.Run("Create without items", func(t *testing.T) {
		_, err := sales.Create(&customers.Customer{}, nil)
		if err == nil {
			t.Error("Should not create without items")
		}
		if !errors.Is(err, sales.ErrItemsMissing) {
			t.Errorf("Should return ErrItemsMissing, got %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		sales.Create(&customers.Customer{Name: "Jane Doe"}, []*sales.Item{
			{
				Qty:     1,
				Price:   100,
				Product: &products.Product{Name: "Mouse"},
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

		if err := sales.List().With("Customer").With("Items").Get(&items); err != nil {
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
		if err := sales.Find(2).With("Customer").With("Items").First(&sale); err != nil {
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
}
