package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/accounting/api"
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func Request(t *testing.T, method, url string, data interface{}) *http.Request {
	t.Helper()
	jsonBytes, err := json.Marshal(data)

	if err != nil {
		t.Error(err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(jsonBytes))
	req.Header.Set("content-type", "application/json")

	return req
}

func Get(t *testing.T, url string) *http.Request {
	return Request(t, "GET", url, nil)
}

func Post(t *testing.T, url string, data interface{}) *http.Request {
	return Request(t, "POST", url, data)
}

func Put(t *testing.T, url string, data interface{}) *http.Request {
	return Request(t, "PUT", url, data)
}

func Delete(t *testing.T, url string) *http.Request {
	return Request(t, "DELETE", url, nil)
}

func TestAccountsEndpoint(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Account{})
	db.AutoMigrate(&models.Company{})

	t.Cleanup(database.Cleanup)

	if result := db.Create(&models.Company{Name: "Testing Company"}); result.Error != nil {
		t.Error(result.Error)
	}

	router := api.GetRouter()

	t.Run("Create", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := Post(t, "/accounts", map[string]interface{}{
			"name":      "Payables",
			"type":      1,
			"parent_id": nil,
		})
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var account models.Account
		if err := json.Unmarshal(w.Body.Bytes(), &account); err != nil {
			t.Error(err)
		}

		if account.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, account.ID)
		}

		if account.Name != "Payables" {
			t.Errorf("Expectd name %v, got %v", "Payables", account.Name)
		}

		if account.Type != models.Liability {
			t.Errorf("Expected type %v, got %v", models.Liability, account.Type)
		}
	})

	t.Run("Create with parent", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := Post(t, "/accounts", map[string]interface{}{
			"name":      "Suppliers",
			"type":      1,
			"parent_id": 1,
		})
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var account models.Account
		if err := json.Unmarshal(w.Body.Bytes(), &account); err != nil {
			t.Error(err)
		}

		if account.ID != 2 {
			t.Errorf("Expected ID %v, got %v", 2, account.ID)
		}

		if account.Name != "Suppliers" {
			t.Errorf("Expectd name %v, got %v", "Suppliers", account.Name)
		}

		if account.Type != models.Liability {
			t.Errorf("Expected type %v, got %v", models.Liability, account.Type)
		}

		if account.ParentID == nil {
			t.Error("Should have a parent account")
		}
	})

	t.Run("Create with invalid parent", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := Post(t, "/accounts", map[string]interface{}{
			"name":      "Others",
			"type":      1,
			"parent_id": 22,
		})
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %v, got %v", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("Update", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := Put(t, "/accounts/2", map[string]interface{}{
			"name":      "Suppliers Payable",
			"parent_id": nil,
		})
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var account *models.Account
		if result := db.First(&account, 2); result.Error != nil {
			t.Error(result.Error)
		}

		if account.Name != "Suppliers Payable" {
			t.Errorf("Expected name %v, got %v", "Suppliers Payable", account.Name)
		}

		if account.Type != models.Liability {
			t.Errorf("Expected type %v, got %v", models.Liability, account.Type)
		}

		if account.ParentID != nil {
			t.Errorf("Expected no parent, got %v", account.ParentID)
		}
	})

	t.Run("Update non existing account", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := Put(t, "/accounts/2444", map[string]interface{}{
			"name": "Suppliers Payable",
		})

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update without account", func(t *testing.T) {
		req := Put(t, "/accounts/", map[string]interface{}{
			"name": "Suppliers Payable",
		})

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("List", func(t *testing.T) {
		req := Get(t, "/accounts")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expect status %v, got %v", http.StatusOK, w.Code)
		}

		var accounts []models.Account
		if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
			t.Error(err)
		}

		if len(accounts) != 2 {
			t.Errorf("Expected %v accounts, got %v", 2, len(accounts))
		}

		for idx, acc := range accounts {
			if idx == 0 && acc.Name != "Payables" {
				t.Errorf("Expected name %v, got %v", "Payables", acc.Name)
			}

			if idx == 1 && acc.Name != "Suppliers Payable" {
				t.Errorf("Expected name %v, got %v", "Suppliers Payables", acc.Name)
			}
		}
	})

	t.Run("View", func(t *testing.T) {
		req := Get(t, "/accounts/2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var account models.Account
		if err := json.Unmarshal(w.Body.Bytes(), &account); err != nil {
			t.Error(err)
		}

		if account.Name != "Suppliers Payable" {
			t.Errorf("Expected name %v, got %v", "Suppliers Payable", account.Name)
		}
	})

	t.Run("View non existing", func(t *testing.T) {
		req := Get(t, "/accounts/5155")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("View no account ID", func(t *testing.T) {
		req := Get(t, "/accounts/ ")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		req := Delete(t, "/accounts/2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}
	})

	t.Run("Delete non existing account", func(t *testing.T) {
		req := Delete(t, "/accounts/2215")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete invalid account", func(t *testing.T) {
		req := Delete(t, "/accounts/aoeu")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})
}
