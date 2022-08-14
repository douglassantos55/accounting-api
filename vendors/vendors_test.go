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
		vendor, err := vendors.Create("Vendor 1", "15.152.412/4441-12", nil)

		if err != nil {
			t.Error(err)
		}

		if vendor.Address != nil {
			t.Error("Should not have address")
		}
	})
}
