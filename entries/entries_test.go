package entries_test

import (
	"errors"
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/entries"
	"example.com/accounting/models"
)

func TestEntries(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.Migrate(&models.Entry{})
	db.Migrate(&models.Account{})
	db.Migrate(&models.Transaction{})

	revenue, _ := accounts.Create("Revenue", models.Revenue, nil)
	cash, _ := accounts.Create("Cash", models.Asset, nil)

	t.Run("Create", func(t *testing.T) {
		entry, err := entries.Create("Description", []*models.Transaction{
			{Account: revenue, Value: 1000},
			{Account: cash, Value: 1000},
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
		entry, err := entries.Create("Description", []*models.Transaction{
			{Account: revenue, Value: 1001},
			{Account: cash, Value: 1000},
		})

		if !errors.Is(err, entries.ErrEntryNotBalanced) {
			t.Errorf("Expeted entry not balanced error, got %v", err)
		}

		if entry != nil {
			t.Error("should not save entry")
		}
	})

	t.Run("List", func(t *testing.T) {
		entries.Create("Entry 2", []*models.Transaction{
			{Account: revenue, Value: 333},
			{Account: cash, Value: 333},
		})

		result, err := entries.List()
		if err != nil {
			t.Error(err)
		}

		var items []*models.Entry
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

		var entry *models.Entry
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

		var items []*models.Entry
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

		var items []*models.Entry
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

		var entry *models.Entry
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

		var updated *models.Entry
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

		var entry *models.Entry
		if err := result.First(&entry); err == nil {
			t.Error("Should not find deleted entry")
		}

		// check if transactions are deleted
		var transactions []*models.Transaction
		if err := db.Find(&models.Transaction{}).Get(&transactions); err != nil {
			t.Error(err)
		}

		if len(transactions) != 2 {
			t.Errorf("Expected %v, got %v", 2, len(transactions))
		}
	})
}
