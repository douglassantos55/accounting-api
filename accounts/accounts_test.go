package accounts_test

import (
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func TestAccount(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()
	db.Migrate(&models.Account{})

	t.Run("Create", func(t *testing.T) {
		account, err := accounts.Create("Cash", models.Asset, nil)

		if err != nil {
			t.Error(err)
		}

		if account.ID == 0 {
			t.Error("Should have an ID")
		}

		if account.Parent != nil {
			t.Error("Should not have a parent")
		}
	})

	t.Run("Create With Parent", func(t *testing.T) {
		accounts.Create("Assets", models.Asset, nil)
		accountID := uint(2)
		cash, err := accounts.Create("Receivables", models.Asset, &accountID)

		if err != nil {
			t.Error(err)
		}

		if cash.ID != 3 {
			t.Errorf("Expected ID 3, got %v", cash.ID)
		}

		if *cash.ParentID != 2 {
			t.Errorf("Expected ParentID to be 2, got %v", cash.ParentID)
		}
	})

	t.Run("List", func(t *testing.T) {
		var items []*models.Account
		err := accounts.List().Get(&items)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("Expected 3, got %v", len(items))
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		var account *models.Account
		err := accounts.Find(3).First(&account)

		if err != nil {
			t.Error(err)
		}

		if account == nil {
			t.Error("Should get account")
		}
	})

	t.Run("Get with Parent", func(t *testing.T) {
		var account *models.Account
		err := accounts.Find(3).With("Parent").First(&account)

		if err != nil {
			t.Error(err)
		}

		if account == nil {
			t.Error("Should get account")
		}

		if account.Parent == nil {
			t.Error("Should have a parent")
		}
	})

	t.Run("Get with children", func(t *testing.T) {
		var account *models.Account
		err := accounts.Find(2).With("Children").First(&account)

		if err != nil {
			t.Error(err)
		}

		if len(account.Children) != 1 {
			t.Errorf("Expected 1 children, got %v", len(account.Children))
		}
	})

	t.Run("Update", func(t *testing.T) {
		var account *models.Account

		if err := accounts.Find(3).First(&account); err != nil {
			t.Error(err)
		}

		previousUpdatedAt := account.UpdatedAt
		account.Name = "Accounts receivable"

		if err := accounts.Update(account); err != nil {
			t.Error(err)
		}

		if account.Name != "Accounts receivable" {
			t.Errorf("Expected 'Accounts receivable', got %v", account.Name)
		}

		if account.UpdatedAt == previousUpdatedAt {
			t.Error("Expected to be updated")
		}
	})

	t.Run("Update Remove Parent", func(t *testing.T) {
		var account *models.Account
		if err := accounts.Find(3).First(&account); err != nil {
			t.Error(err)
		}

		account.ParentID = nil
		if err := accounts.Update(account); err != nil {
			t.Error(err)
		}

		accounts.Find(3).With("Parent").First(&account)
		if account.Parent != nil {
			t.Error("Should remove Parent")
		}
	})

	t.Run("Update With Parent", func(t *testing.T) {
		var account *models.Account
		if err := accounts.Find(3).First(&account); err != nil {
			t.Error(err)
		}

		parent := uint(1)
		account.ParentID = &parent
		if err := accounts.Update(account); err != nil {
			t.Error(err)
		}

		accounts.Find(3).With("Parent").First(&account)
		if account.Parent.ID != 1 {
			t.Errorf("Expected ParentID 1, got %v", account.Parent.ID)
		}
		if account.Parent.Name != "Cash" {
			t.Errorf("Expected 'Cash', got %v", account.Parent.Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := accounts.Delete(3); err != nil {
			t.Error(err)
		}

		var account *models.Account
		if err := accounts.Find(3).First(&account); err == nil {
			t.Error("Account should be deleted")
		}
	})

	t.Run("Delete Non Existing Account", func(t *testing.T) {
		if err := accounts.Delete(153); err == nil {
			t.Error("Should not delete account that does not exist")
		}
	})
}
