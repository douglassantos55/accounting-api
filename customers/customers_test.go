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
		if customer.Address == nil {
			t.Error("Should have address")
		}
	})

	t.Run("Create without address", func(t *testing.T) {
		customer, err := customers.Create(
			"Mark",
			"mark@email.com",
			"459.149.594-10",
			"69 5412-1925",
			nil,
		)
		if err != nil {
			t.Error(err)
		}
		if customer.Address != nil {
			t.Error("Should not have an address")
		}
	})

	t.Run("List", func(t *testing.T) {
		if _, err := customers.Create(
			"Jane Doe",
			"janedoe@email.com",
			"412.461.592-21",
			"44 2105-6542",
			nil,
		); err != nil {
			t.Error(err)
		}

		var items []*customers.Customer
		if err := customers.List().Get(&items); err != nil {
			t.Error(err)
		}
		if len(items) != 3 {
			t.Errorf("Expected %v items, got %v", 3, len(items))
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

	t.Run("Update", func(t *testing.T) {
		var customer *customers.Customer
		if err := customers.Find(1).First(&customer); err != nil {
			t.Error(err)
		}

		customer.Name = "Updated name"
		customer.Email = "updated@email.com"
		customer.Address.Street = "updated street"

		if err := customers.Update(customer); err != nil {
			t.Error(err)
		}

		customers.Find(1).First(&customer)
		if customer.Name != "Updated name" {
			t.Error("Should have updated name")
		}
		if customer.Email != "updated@email.com" {
			t.Error("Should have updated email")
		}
		if customer.Address.Street != "updated street" {
			t.Error("Should have updated street")
		}
	})

	t.Run("Update without address", func(t *testing.T) {
		var customer *customers.Customer
		if err := customers.Find(1).First(&customer); err != nil {
			t.Error(err)
		}

		customer.Address = nil
		if err := customers.Update(customer); err != nil {
			t.Error(err)
		}

		customers.Find(1).First(&customer)
		if customer.Address.Street != "" {
			t.Error("Should have removed address")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := customers.Delete(3); err != nil {
			t.Error(err)
		}

		if err := customers.Find(3).First(&customers.Customer{}); err == nil {
			t.Error("Customer should have been deleted")
		}
	})
}
