package accounts_test

import (
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
)

func TestAccount(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()
	db.Migrate(&accounts.Account{})

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		account, err := accounts.Create("Cash", accounts.Asset, 0)

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
		accounts.Create("Assets", accounts.Asset, 0)
		cash, err := accounts.Create("Receivables", accounts.Asset, 2)

		if err != nil {
			t.Error(err)
		}

		if cash.ID != 3 {
			t.Errorf("Expected ID 3, got %v", cash.ID)
		}

		if cash.ParentID != 2 {
			t.Errorf("Expected ParentID to be 2, got %v", cash.ParentID)
		}
	})

	t.Run("List", func(t *testing.T) {
		var items []*accounts.Account
		err := accounts.List().Get(&items)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("Expected 3, got %v", len(items))
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		var account *accounts.Account
		err := accounts.Find(3).Get(&account)

		if err != nil {
			t.Error(err)
		}

		if account == nil {
			t.Error("Should get account")
		}
	})

	t.Run("Get with Parent", func(t *testing.T) {
		var account *accounts.Account
		err := accounts.Find(3).With("Parent").Get(&account)

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
		var account *accounts.Account
		err := accounts.Find(2).With("Children").Get(&account)

		if err != nil {
			t.Error(err)
		}

		if len(account.Children) != 1 {
			t.Errorf("Expected 1 children, got %v", len(account.Children))
		}
	})
}
