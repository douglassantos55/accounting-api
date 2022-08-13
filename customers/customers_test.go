package customers_test

import (
	"testing"

	"example.com/accounting/customers"
	"example.com/accounting/database"
)

func TestCustomers(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()
	db.Migrate(&customers.Customer{})

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		address := &customers.Address{
			"rua abc",
			"7979",
			"Centro",
			"Sao Paulo",
			"SP",
			"13200-000",
		}

		customer, err := customers.Create(
			"John Doe",
			"johndoe@email.com",
			"753.515.151-15",
			"14 2230-6412",
			address,
		)

		if err != nil {
			t.Error(err)
		}
		if customer.ID == 0 {
			t.Error("Should have an ID")
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		var customer *customers.Customer

		if err := customers.Find(1).First(&customer); err != nil {
			t.Error(err)
		}
		if customer.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, customer.ID)
		}
		if customer.Name != "John Doe" {
			t.Errorf("Expected Name %v, got %v", "John Doe", customer.Name)
		}
		if customer.Email != "johndoe@email.com" {
			t.Errorf("Expected Email %v, got %v", "johndoe@email.com", customer.Email)
		}

		if customer.Address.Street != "rua abc" {
			t.Errorf("Expected Street %v, got %v", "rua abc", customer.Address.Street)
		}
	})
}
