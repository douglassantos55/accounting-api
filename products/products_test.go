package products_test

import (
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
	db.Migrate(&accounts.Account{})

	accounts.Create("Revenue", accounts.Revenue, nil)

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		prod, err := products.Create("Keyboard", 350.5, 35, 1, nil)

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
		_, err := products.Create("Coffee Powder", 33.6, 37, 0, nil)

		if err == nil {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("Create With Non Existing Account", func(t *testing.T) {
		_, err := products.Create("Coffee Powder", 33.6, 55, 10, nil)

		if err == nil {
			t.Error("Should not be able to create product without revenue account")
		}
	})

	t.Run("List", func(t *testing.T) {
		products.Create("Monitor", 1350.5, 51, 1, nil)
		products.Create("Mouse", 150.5, 231, 1, nil)

		var items []*products.Product
		err := products.List().Get(&items)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("Expected %v items, got %v", 3, len(items))
		}
	})

	t.Run("List With Account", func(t *testing.T) {
		var items []*products.Product
		err := products.List().With("Account").Get(&items)

		if err != nil {
			t.Error(err)
		}

		for _, product := range items {
			if product.Account == nil {
				t.Error("Should have revenue account")
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

	t.Run("Get With Account", func(t *testing.T) {
		var product *products.Product

		if err := products.Find(3).With("Account").First(&product); err != nil {
			t.Error(err)
		}

		if product.Account == nil {
			t.Error("Should have Account")
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

	t.Run("Update Without Account", func(t *testing.T) {
		var item *products.Product
		if err := products.Find(3).First(&item); err != nil {
			t.Error(err)
		}

		item.AccountID = 0
		if err := products.Update(item); err == nil {
			t.Error("Should not be able to update product without revenue account")
		}

		var product *products.Product
		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.AccountID == 0 {
			t.Error("Should have Account")
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

		product, err := products.Create("Prod", 100, 11, 1, &vendor.ID)
		if err != nil {
			t.Error(err)
		}

		if err := products.Find(3).First(&product); err != nil {
			t.Error(err)
		}

		if product.Name != "Prod" {
			t.Errorf("Expected name %v, got %v", "Prod", product.Name)
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
}
