package vendors_test

import (
	"testing"

	"example.com/accounting/database"
	"example.com/accounting/models"
	"example.com/accounting/vendors"
)

func TestVendors(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.Migrate(&models.Vendor{})

	db.Create(&models.Company{
		Name: "Testing Company",
	})

	t.Run("Create", func(t *testing.T) {
		address := &models.Address{}
		vendor, err := vendors.Create(1, "Vendor 1", "15.152.412/4441-12", address)

		if err != nil {
			t.Error(err)
		}

		if vendor.ID == 0 {
			t.Error("Should save vendor")
		}
	})

	t.Run("Create without company", func(t *testing.T) {
		address := &models.Address{}
		if _, err := vendors.Create(0, "Vendor 1", "15.152.412/4441-12", address); err == nil {
			t.Error("Should not create without company")
		}
	})

	t.Run("Create without Address", func(t *testing.T) {
		vendor, err := vendors.Create(1, "Vendor 2", "25.252.212/2441-12", nil)

		if err != nil {
			t.Error(err)
		}

		if vendor.Address != nil {
			t.Error("Should not have address")
		}
	})

	t.Run("List", func(t *testing.T) {
		var items []*models.Vendor
		if err := vendors.List().Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(items))
		}
	})

	t.Run("List by condition", func(t *testing.T) {
		var items []*models.Vendor
		if err := vendors.List().Where("Name LIKE ?", "%Vendor%").Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(items))
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		var vendor *models.Vendor
		if err := vendors.Find(2).First(&vendor); err != nil {
			t.Error(err)
		}

		if vendor == nil {
			t.Error("Should have retrieved vendor")
		}
	})

	t.Run("Get by condition", func(t *testing.T) {
		var vendor *models.Vendor
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
		var vendor *models.Vendor
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

	t.Run("Delete", func(t *testing.T) {
		if err := vendors.Delete(2); err != nil {
			t.Error(err)
		}

		var vendor *models.Vendor
		if err := vendors.Find(2).First(&vendor); err == nil {
			t.Error("Should have deleted vendor")
		}

		if vendor.ID != 0 {
			t.Error("Should have not found vendor")
		}
	})
}
