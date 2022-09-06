package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/accounting/api"
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func TestEntries(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Entry{})
	db.AutoMigrate(&models.Account{})
	db.AutoMigrate(&models.Transaction{})

	t.Cleanup(database.Cleanup)

	router := api.GetRouter()

	db.Create(&models.Company{Name: "Testing Company"})

	cash := &models.Account{
		Name:      "Cash",
		Type:      models.Asset,
		CompanyID: 1,
	}

	revenue := &models.Account{
		Name:      "Revenue",
		CompanyID: 1,
		Type:      models.Revenue,
	}

	db.Create(cash)
	db.Create(revenue)

	t.Run("Create", func(t *testing.T) {
		req := Post(t, "/entries", map[string]interface{}{
			"description": "Sales of apples",
			"transactions": []map[string]interface{}{
				{"account_id": cash.ID, "value": 1000},
				{"account_id": revenue.ID, "value": 1000},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var entry models.Entry
		if err := json.Unmarshal(w.Body.Bytes(), &entry); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if entry.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, entry.ID)
		}

		if entry.Description != "Sales of apples" {
			t.Errorf("Expected description %v, got %v", "Sales of apples", entry.Description)
		}

		if len(entry.Transactions) != 2 {
			t.Errorf("Expected %v transactions, got %v", 2, len(entry.Transactions))
		}

		if !entry.IsBalanced() {
			t.Error("Entry should be balanced")
		}
	})

	t.Run("Create unbalanced", func(t *testing.T) {
		req := Post(t, "/entries", map[string]interface{}{
			"description": "Sales of apples",
			"transactions": []map[string]interface{}{
				{"account_id": cash.ID, "value": 1000},
				{"account_id": cash.ID, "value": 1000},
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if response["error"] != api.ErrEntryNotBalanced.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrEntryNotBalanced.Error(), response["error"])
		}
	})

	t.Run("Create without transactions", func(t *testing.T) {
		req := Post(t, "/entries", map[string]interface{}{
			"description":  "Sales of apples",
			"transactions": []map[string]interface{}{},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("List", func(t *testing.T) {
		db.Create(&models.Company{Name: "Other Company"})

		// This entry should not be retrieved
		db.Create(&models.Entry{
			Description: "Sales of oranges",
			CompanyID:   2,
			Transactions: []*models.Transaction{
				{Value: 100, AccountID: cash.ID},
				{Value: 100, AccountID: revenue.ID},
			},
		})

		req := Get(t, "/entries")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var entries []models.Entry
		if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(entries) != 1 {
			t.Errorf("Expected %v entries, got %v", 1, len(entries))
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := Get(t, "/entries/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var entry models.Entry
		if err := json.Unmarshal(w.Body.Bytes(), &entry); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if entry.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, entry.ID)
		}
	})

	t.Run("Get non existent", func(t *testing.T) {
		req := Get(t, "/entries/1533")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get invalid", func(t *testing.T) {
		req := Get(t, "/entries/asoetnuh")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get from another company", func(t *testing.T) {
		req := Get(t, "/entries/2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})
}
