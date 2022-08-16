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
		entry, err := entries.Create("Description", []*entries.Transaction{
			{AccountID: 1, Value: 1001},
			{AccountID: 2, Value: 1000},
		})

		if !errors.Is(err, entries.ErrEntryNotBalanced) {
			t.Errorf("Expeted entry not balanced error, got %v", err)
		}

		if entry != nil {
			t.Error("should not save entry")
		}
	})

	t.Run("List", func(t *testing.T) {
		entries.Create("Entry 2", []*entries.Transaction{
			{AccountID: 1, Value: 333},
			{AccountID: 2, Value: 333},
		})

		result, err := entries.List()
		if err != nil {
			t.Error(err)
		}

		var items []*entries.Entry
		if err := result.Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(items))
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		result, err := entries.Find(2)
		if err != nil {
			t.Error(err)
		}

		var entry *entries.Entry
		if err := result.With("Transactions").First(&entry); err != nil {
			t.Error(err)
		}

		if entry == nil || entry.ID == 0 {
			t.Error("Should retrieve entry")
		}

		if entry.Description != "Entry 2" {
			t.Errorf("Expected Entry 2, got %v", entry.Description)
		}

		if len(entry.Transactions) != 2 {
			t.Errorf("Expected %v transactions, got %v", 2, len(entry.Transactions))
		}

		for _, transaction := range entry.Transactions {
			if transaction.Value != 333 {
				t.Errorf("Expected value %v, got %v", 333, transaction.Value)
			}
		}
	})
}
