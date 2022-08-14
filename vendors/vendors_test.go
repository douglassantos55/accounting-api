package vendors_test

import (
	"testing"

	"example.com/accounting/customers"
	"example.com/accounting/database"
	"example.com/accounting/vendors"
)

func TestVendors(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()

	db.Migrate(&vendors.Vendor{})

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		address := &customers.Address{}
		vendor, err := vendors.Create("Vendor 1", "15.152.412/4441-12", address)

		if err != nil {
			t.Error(err)
		}

		if vendor.ID == 0 {
			t.Error("Should save vendor")
		}
	})

	t.Run("Create without Address", func(t *testing.T) {
		vendor, err := vendors.Create("Vendor 2", "25.252.212/2441-12", nil)

		if err != nil {
			t.Error(err)
		}

		if vendor.Address != nil {
			t.Error("Should not have address")
		}
	})

	t.Run("List", func(t *testing.T) {
		var items []*vendors.Vendor
		if err := vendors.List().Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(items))
		}
	})

	t.Run("List by condition", func(t *testing.T) {
		var items []*vendors.Vendor
		if err := vendors.List().Where("Name LIKE ?", "%Vendor%").Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(items))
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		var vendor *vendors.Vendor
		if err := vendors.Find(2).First(&vendor); err != nil {
			t.Error(err)
		}

		if vendor == nil {
			t.Error("Should have retrieved vendor")
		}
	})

	t.Run("Get by condition", func(t *testing.T) {
		var vendor *vendors.Vendor
		if err := vendors.List().Where("Name", "Vendor 1").First(&vendor); err != nil {
			t.Error(err)
		}

		if vendor == nil {
			t.Error("Should have retrieved vendor")
		}
		if vendor.Name != "Vendor 1" {
			t.Errorf("Expected name %v, got %v", "Vendor 1", vendor.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var vendor *vendors.Vendor
		if err := vendors.Find(1).First(&vendor); err != nil {
			t.Error(err)
		}

		previousUpdatedAt := vendor.UpdatedAt
		vendor.Name = "Updated vendor"
		vendor.Address = nil

		if err := vendors.Update(vendor); err != nil {
			t.Error(err)
		}

		vendors.Find(1).First(&vendor)

		if previousUpdatedAt == vendor.UpdatedAt {
			t.Error("Should to have updated")
		}

		if vendor.Name != "Updated vendor" {
			t.Error("Should to have updated name")
		}
		if vendor.Address.Street != "" {
			t.Error("Expected address to be removed")
		}
	})
}
