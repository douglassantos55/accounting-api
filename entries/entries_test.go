package entries_test

import (
	"errors"
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/entries"
)

func TestEntries(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()

	db.Migrate(&entries.Entry{})
	db.Migrate(&accounts.Account{})
	db.Migrate(&entries.Transaction{})

	accounts.Create("Revenue", accounts.Revenue, nil)
	accounts.Create("Cash", accounts.Asset, nil)

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		entry, err := entries.Create("Description", []*entries.Transaction{
			{AccountID: 1, Value: 1000},
			{AccountID: 2, Value: 1000},
		})

		if err != nil {
			t.Error(err)
		}

		if entry.ID == 0 {
			t.Error("Should have saved entry")
		}

		if entry.Description != "Description" {
			t.Errorf("Expected Description, got %v", entry.Description)
		}

		for _, transaction := range entry.Transactions {
			if transaction.EntryID == 0 {
				t.Error("Should associate with entry")
			}
		}
	})

	t.Run("Unbalanced", func(t *testing.T) {
		_, err := entries.Create("Description", []*entries.Transaction{
			{AccountID: 1, Value: 1001},
			{AccountID: 2, Value: 1000},
		})

		if !errors.Is(err, entries.ErrEntryNotBalanced) {
			t.Errorf("Expeted entry not balanced error, got %v", err)
		}
	})
}
