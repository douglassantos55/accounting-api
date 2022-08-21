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

	db.Migrate(&accounts.Entry{})
	db.Migrate(&accounts.Account{})
	db.Migrate(&accounts.Transaction{})

	accounts.Create("Revenue", accounts.Revenue, nil)
	accounts.Create("Cash", accounts.Asset, nil)

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		entry, err := entries.Create("Description", []*accounts.Transaction{
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
		entry, err := entries.Create("Description", []*accounts.Transaction{
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
		entries.Create("Entry 2", []*accounts.Transaction{
			{AccountID: 1, Value: 333},
			{AccountID: 2, Value: 333},
		})

		result, err := entries.List()
		if err != nil {
			t.Error(err)
		}

		var items []*accounts.Entry
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

		var entry *accounts.Entry
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

	t.Run("Filter by non existing account", func(t *testing.T) {
		result, err := entries.List()
		if err != nil {
			t.Error(err)
		}

		var items []*accounts.Entry
		result = result.WhereHas("transactions", "entries.id = transactions.entry_id AND transactions.account_id = ?", 5)
		if err := result.Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 0 {
			t.Errorf("Should not find anything with account 5, got %v", items)
		}
	})

	t.Run("Filter by account", func(t *testing.T) {
		result, err := entries.List()
		if err != nil {
			t.Error(err)
		}

		var items []*accounts.Entry
		result = result.WhereHas("transactions", "entries.id = transactions.entry_id AND transactions.account_id = ?", 2)
		if err := result.Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v entries, got %v", 2, len(items))
		}
	})

	t.Run("Update", func(t *testing.T) {
		result, err := entries.Find(2)
		if err != nil {
			t.Error(err)
		}

		var entry *accounts.Entry
		if err := result.With("Transactions").First(&entry); err != nil {
			t.Error(err)
		}

		entry.Description = "Updated entry"
		entry.Transactions[0].Value = 11
		entry.Transactions[0].AccountID = 2
		entry.Transactions[1].Value = 11
		entry.Transactions[1].AccountID = 1

		if err := entries.Update(entry); err != nil {
			t.Error(err)
		}

		result, _ = entries.Find(2)

		var updated *accounts.Entry
		result.With("Transactions").First(&updated)

		if updated.Description != "Updated entry" {
			t.Errorf("Expected %v, got %v", "Updated entry", updated.Description)
		}

		for i, transaction := range updated.Transactions {
			if int(transaction.AccountID) != 2/(i+1) {
				t.Errorf("Expected %v, got %v", 2/(i+1), transaction.AccountID)
			}
			if transaction.Value != 11 {
				t.Errorf("Expected %v, got %v", 11, transaction.Value)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := entries.Delete(2); err != nil {
			t.Error(err)
		}

		result, err := entries.Find(2)
		if err != nil {
			t.Error(err)
		}

		var entry *accounts.Entry
		if err := result.First(&entry); err == nil {
			t.Error("Should not find deleted entry")
		}

		// check if transactions are deleted
		var transactions []*accounts.Transaction
		if err := db.Find(&accounts.Transaction{}).Get(&transactions); err != nil {
			t.Error(err)
		}

		if len(transactions) != 2 {
			t.Errorf("Expected %v, got %v", 2, len(transactions))
		}
	})
}
