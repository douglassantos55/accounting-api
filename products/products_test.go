package products_test

import (
	"errors"
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/products"
	"example.com/accounting/vendors"
)

func TestProducts(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()

	db.Migrate(&products.Product{})
	db.Migrate(&products.StockEntry{})
	db.Migrate(&accounts.Account{})

	revenue, _ := accounts.Create("Revenue", accounts.Revenue, nil)
	inventory, _ := accounts.Create("Inventory", accounts.Asset, nil)
	receivables, _ := accounts.Create("Receivables", accounts.Asset, nil)

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		prod := &products.Product{
			Name:                "Keyboard",
			Price:               350.5,
			Purchasable:         true,
			RevenueAccountID:    &revenue.ID,
			InventoryAccountID:  inventory.ID,
			ReceivableAccountID: &receivables.ID,
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

		if *prod.ReceivableAccountID != 3 {
			t.Errorf("Expected AccountID 3, got %v", prod.ReceivableAccountID)
		}
	})

	t.Run("Create Without Revenue Account", func(t *testing.T) {
		err := products.Create(&products.Product{
			Name:                "Coffe Powder",
			Price:               33.6,
			Purchasable:         true,
			ReceivableAccountID: &receivables.ID,
			InventoryAccountID:  inventory.ID,
		})

		if !errors.Is(err, products.ErrRevenueAccountMissing) {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("Create Without Receivable Account", func(t *testing.T) {
		err := products.Create(&products.Product{
			Name:               "Iron plate",
			Price:              330.6,
			Purchasable:        true,
			RevenueAccountID:   &revenue.ID,
			InventoryAccountID: inventory.ID,
		})

		if !errors.Is(err, products.ErrReceivableAccountMissing) {
			t.Error("Should not be able to create product without receivable account")
		}
	})

	t.Run("Create Without Inventory Account", func(t *testing.T) {
		err := products.Create(&products.Product{
			Name:                "Concrete",
			Price:               50.5,
			Purchasable:         true,
			RevenueAccountID:    &revenue.ID,
			ReceivableAccountID: &receivables.ID,
		})

		if err == nil {
			t.Error("Should not be able to create product without inventory account")
		}
	})

	t.Run("Create With Non Existing Revenue Account", func(t *testing.T) {
		fakeId := uint(15115)

		err := products.Create(&products.Product{
			Name:                "Door",
			Price:               70.5,
			Purchasable:         true,
			RevenueAccountID:    &fakeId,
			InventoryAccountID:  inventory.ID,
			ReceivableAccountID: &receivables.ID,
		})

		if err == nil {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("Create With Non Existing Receivable Account", func(t *testing.T) {
		fakeId := uint(15115)

		err := products.Create(&products.Product{
			Name:                "Door knob",
			Price:               20.5,
			Purchasable:         true,
			RevenueAccountID:    &revenue.ID,
			InventoryAccountID:  inventory.ID,
			ReceivableAccountID: &fakeId,
		})

		if err == nil {
			t.Error("Should not be able to create product without receivable account")
		}
	})

	t.Run("Create With Non Existing Inventory Account", func(t *testing.T) {
		fakeId := uint(15115)

		err := products.Create(&products.Product{
			Name:                "Guitar",
			Price:               720.5,
			Purchasable:         true,
			RevenueAccountID:    &revenue.ID,
			InventoryAccountID:  fakeId,
			ReceivableAccountID: &receivables.ID,
		})

		if err == nil {
			t.Error("Should not be able to create product without inventory account")
		}
	})

	t.Run("List", func(t *testing.T) {
		products.Create(&products.Product{
			Name:                "Monitor",
			Price:               1350.5,
			Purchasable:         true,
			RevenueAccountID:    &revenue.ID,
			ReceivableAccountID: &receivables.ID,
			InventoryAccountID:  inventory.ID,
		})

		products.Create(&products.Product{
			Name:                "Mouse",
			Price:               150.5,
			Purchasable:         true,
			RevenueAccountID:    &revenue.ID,
			ReceivableAccountID: &receivables.ID,
			InventoryAccountID:  inventory.ID,
		})

		var items []*products.Product
		err := products.List().Get(&items)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("Expected %v items, got %v", 3, len(items))
		}
	})

	t.Run("List With Accounts", func(t *testing.T) {
		var items []*products.Product
		err := products.List().With("*").Get(&items)

		if err != nil {
			t.Error(err)
		}

		for _, product := range items {
			if product.Purchasable && product.RevenueAccount == nil {
				t.Error("Should have revenue account")
			}

			if product.Purchasable && product.ReceivableAccount == nil {
				t.Error("Should have receivable account")
			}

			if product.InventoryAccount == nil {
				t.Error("Should have inventory account")
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

	t.Run("Get With Accounts", func(t *testing.T) {
		var product *products.Product

		if err := products.Find(3).With("RevenueAccount", "InventoryAccount", "ReceivableAccount").First(&product); err != nil {
			t.Error(err)
		}

		if product.Purchasable && product.RevenueAccount == nil {
			t.Error("Should have revenue account")
		}

		if product.Purchasable && product.ReceivableAccount == nil {
			t.Error("Should have receivable account")
		}

		if product.InventoryAccount == nil {
			t.Error("Should have inventory account")
		}
	})

	t.Run("Update", func(t *testing.T) {
		var item *products.Product
		products.Find(3).First(&item)

		item.Name = "Mousepad"
		item.Price = 30.3

		if err := products.Update(item); err != nil {
			t.Error(err)
		}

		var product *products.Product
		products.Find(3).First(&product)

		if product.Name != "Mousepad" {
			t.Errorf("Expected name %v, got %v", "Mousepad", product.Name)
		}

		if product.Price != 30.3 {
			t.Errorf("Expected price %v, got %v", 30.3, product.Price)
		}
	})

	t.Run("Update Without Revenue Account", func(t *testing.T) {
		var item *products.Product
		if err := products.Find(3).First(&item); err != nil {
			t.Error(err)
		}

		item.RevenueAccountID = nil
		if err := products.Update(item); err == nil {
			t.Error("Should not be able to update product without revenue account")
		}

		var product *products.Product
		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.RevenueAccountID == nil {
			t.Error("Should have Revenue Account")
		}
	})

	t.Run("Update Without Receivable Account", func(t *testing.T) {
		var item *products.Product
		if err := products.Find(3).First(&item); err != nil {
			t.Error(err)
		}

		item.ReceivableAccountID = nil
		if err := products.Update(item); err == nil {
			t.Error("Should not be able to update product without receivable account")
		}

		var product *products.Product
		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.ReceivableAccountID == nil {
			t.Error("Should have Receivable Account")
		}
	})

	t.Run("Update Without Inventory Account", func(t *testing.T) {
		var item *products.Product
		if err := products.Find(3).First(&item); err != nil {
			t.Error(err)
		}

		item.InventoryAccountID = 0
		if err := products.Update(item); err == nil {
			t.Error("Should not be able to update product without inventory account")
		}

		var product *products.Product
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

		var product *products.Product
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
		vendor, err := vendors.Create("Vendor", "", nil)
		if err != nil {
			t.Error(err)
		}

		if err := products.Create(&products.Product{
			Name:               "Prod",
			Price:              100,
			VendorID:           &vendor.ID,
			InventoryAccountID: inventory.ID,
		}); err != nil {
			t.Error(err)
		}

		var product *products.Product
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
		var item *products.Product
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

		var product *products.Product
		products.Find(3).First(&product)

		if product.VendorID != nil {
			t.Error("Should not have Vendor")
		}
	})

	t.Run("Create without purchasable", func(t *testing.T) {
		err := products.Create(&products.Product{
			Name:               "Production Supply",
			Price:              7000,
			Purchasable:        false,
			InventoryAccountID: inventory.ID,
		})

		if err != nil {
			t.Error("Should create without receivables and revenue accounts")
		}
	})
}
