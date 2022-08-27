package products_test

import (
	"errors"
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/models"
	"example.com/accounting/products"
	"example.com/accounting/vendors"
)

func TestProducts(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.Migrate(&models.Product{})
	db.Migrate(&models.StockEntry{})
	db.Migrate(&models.Account{})

	db.Create(&models.Company{
		Name: "Testing Company",
	})

	revenue, _ := accounts.Create(1, "Revenue", models.Revenue, nil)
	inventory, _ := accounts.Create(1, "Inventory", models.Asset, nil)

	t.Run("Create", func(t *testing.T) {
		prod := &models.Product{
			Name:               "Keyboard",
			Price:              350.5,
			Purchasable:        true,
			CompanyID:          1,
			RevenueAccountID:   &revenue.ID,
			InventoryAccountID: inventory.ID,
		}

		if err := products.Create(prod); err != nil {
			t.Error(err)
		}

		if prod.ID != 1 {
			t.Errorf("Expected ID 1, got %v", prod.ID)
		}

		if prod.Price != 350.5 {
			t.Errorf("Expected price 350.5, got %v", prod.Price)
		}

		if *prod.RevenueAccountID != 1 {
			t.Errorf("Expected AccountID 1, got %v", prod.RevenueAccountID)
		}

		if prod.InventoryAccountID != 2 {
			t.Errorf("Expected AccountID 2, got %v", prod.InventoryAccountID)
		}
	})

	t.Run("Create Without Revenue Account", func(t *testing.T) {
		err := products.Create(&models.Product{
			Name:               "Coffe Powder",
			Price:              33.6,
			Purchasable:        true,
			CompanyID:          1,
			InventoryAccountID: inventory.ID,
		})

		if !errors.Is(err, products.ErrRevenueAccountMissing) {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("Create Without Inventory Account", func(t *testing.T) {
		err := products.Create(&models.Product{
			Name:             "Concrete",
			Price:            50.5,
			CompanyID:        1,
			Purchasable:      true,
			RevenueAccountID: &revenue.ID,
		})

		if err == nil {
			t.Error("Should not be able to create product without inventory account")
		}
	})

	t.Run("Create With Non Existing Revenue Account", func(t *testing.T) {
		fakeId := uint(15115)

		err := products.Create(&models.Product{
			Name:               "Door",
			CompanyID:          1,
			Price:              70.5,
			Purchasable:        true,
			RevenueAccountID:   &fakeId,
			InventoryAccountID: inventory.ID,
		})

		if err == nil {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("Create With Non Existing Inventory Account", func(t *testing.T) {
		fakeId := uint(15115)

		err := products.Create(&models.Product{
			Name:               "Guitar",
			CompanyID:          1,
			Price:              720.5,
			Purchasable:        true,
			RevenueAccountID:   &revenue.ID,
			InventoryAccountID: fakeId,
		})

		if err == nil {
			t.Error("Should not be able to create product without inventory account")
		}
	})

	t.Run("List", func(t *testing.T) {
		products.Create(&models.Product{
			Name:               "Monitor",
			CompanyID:          1,
			Price:              1350.5,
			Purchasable:        true,
			RevenueAccountID:   &revenue.ID,
			InventoryAccountID: inventory.ID,
		})

		products.Create(&models.Product{
			Name:               "Mouse",
			Price:              150.5,
			Purchasable:        true,
			CompanyID:          1,
			RevenueAccountID:   &revenue.ID,
			InventoryAccountID: inventory.ID,
		})

		var items []*models.Product
		err := products.List().Get(&items)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("Expected %v items, got %v", 3, len(items))
		}
	})

	t.Run("List With Accounts", func(t *testing.T) {
		var items []*models.Product
		err := products.List().With("*").Get(&items)

		if err != nil {
			t.Error(err)
		}

		for _, product := range items {
			if product.Purchasable && product.RevenueAccount == nil {
				t.Error("Should have revenue account")
			}

			if product.InventoryAccount == nil {
				t.Error("Should have inventory account")
			}
		}
	})

	t.Run("Get By ID", func(t *testing.T) {
		var product *models.Product

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

	t.Run("Get With Accounts", func(t *testing.T) {
		var product *models.Product

		if err := products.Find(3).With("RevenueAccount", "InventoryAccount").First(&product); err != nil {
			t.Error(err)
		}

		if product.Purchasable && product.RevenueAccount == nil {
			t.Error("Should have revenue account")
		}

		if product.InventoryAccount == nil {
			t.Error("Should have inventory account")
		}
	})

	t.Run("Update", func(t *testing.T) {
		var item *models.Product
		products.Find(3).First(&item)

		item.Name = "Mousepad"
		item.Price = 30.3

		if err := products.Update(item); err != nil {
			t.Error(err)
		}

		var product *models.Product
		products.Find(3).First(&product)

		if product.Name != "Mousepad" {
			t.Errorf("Expected name %v, got %v", "Mousepad", product.Name)
		}

		if product.Price != 30.3 {
			t.Errorf("Expected price %v, got %v", 30.3, product.Price)
		}
	})

	t.Run("Update Without Revenue Account", func(t *testing.T) {
		var item *models.Product
		if err := products.Find(3).First(&item); err != nil {
			t.Error(err)
		}

		item.RevenueAccountID = nil
		if err := products.Update(item); err == nil {
			t.Error("Should not be able to update product without revenue account")
		}

		var product *models.Product
		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.RevenueAccountID == nil {
			t.Error("Should have Revenue Account")
		}
	})

	t.Run("Update Without Inventory Account", func(t *testing.T) {
		var item *models.Product
		if err := products.Find(3).First(&item); err != nil {
			t.Error(err)
		}

		item.InventoryAccountID = 0
		if err := products.Update(item); err == nil {
			t.Error("Should not be able to update product without inventory account")
		}

		var product *models.Product
		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.InventoryAccountID == 0 {
			t.Error("Should have Inventory Account")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := products.Delete(3); err != nil {
			t.Error(err)
		}

		var product *models.Product
		if err := products.Find(3).First(&product); err == nil {
			t.Error("Product should be deleted")
		}
	})

	t.Run("Delete Non Existing Product", func(t *testing.T) {
		if err := products.Delete(151); err == nil {
			t.Error("Should not delete product that does not exist")
		}
	})

	t.Run("Create with Vendor", func(t *testing.T) {
		vendor, err := vendors.Create(1, "Vendor", "", nil)
		if err != nil {
			t.Error(err)
		}

		if err := products.Create(&models.Product{
			CompanyID:          1,
			Name:               "Prod",
			Price:              100,
			VendorID:           &vendor.ID,
			InventoryAccountID: inventory.ID,
		}); err != nil {
			t.Error(err)
		}

		var product *models.Product
		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.Name != "Prod" {
			t.Errorf("Expected name %v, got %v", "Prod", product.Name)
		}

		if *product.VendorID != vendor.ID {
			t.Errorf("Expected vendor %v, got %v", vendor.ID, *product.VendorID)
		}
	})

	t.Run("Update without vendor", func(t *testing.T) {
		var item *models.Product
		if err := products.Find(3).First(&item); err != nil {
			t.Error(err)
		}

		if *item.VendorID != 1 {
			t.Error("Should have vendor")
		}

		item.VendorID = nil
		if err := products.Update(item); err != nil {
			t.Error("Should be able to update product without vendor")
		}

		var product *models.Product
		products.Find(3).First(&product)

		if product.VendorID != nil {
			t.Error("Should not have Vendor")
		}
	})

	t.Run("Create without purchasable", func(t *testing.T) {
		err := products.Create(&models.Product{
			Name:               "Production Supply",
			Price:              7000,
			Purchasable:        false,
			InventoryAccountID: inventory.ID,
			CompanyID:          1,
		})

		if err != nil {
			t.Error("Should create without receivables and revenue accounts")
		}
	})
}
