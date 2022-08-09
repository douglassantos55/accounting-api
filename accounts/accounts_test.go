package accounts_test

import (
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
)

func TestAccount(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "test.sqlite")

	db, _ := database.GetConnection()
	db.Migrate(&accounts.Account{})

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		account, err := accounts.Create("Cash", accounts.Asset, nil)

		if err != nil {
			t.Error(err)
		}

		if account.ID == 0 {
			t.Error("Should have an ID")
		}
	})

	t.Run("With Parent", func(t *testing.T) {
		parent, _ := accounts.Create("Assets", accounts.Asset, nil)
		cash, err := accounts.Create("Cash", accounts.Asset, parent)

		if err != nil {
			t.Error(err)
		}

		if cash.ID != 3 {
			t.Errorf("Expected ID 3, got %v", cash.ID)
		}

		if cash.ParentID != 2 {
			t.Errorf("Expected ParentID to be 2, got %v", cash.ParentID)
		}

		if cash.Parent == nil {
			t.Error("Should have a parent")
		}
	})
}
