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

func TestCustomersEndpoint(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "file::memory:?cache=shared")

	db, _ := database.GetConnection()

	db.AutoMigrate(&models.Company{})
	db.AutoMigrate(&models.Customer{})

	t.Cleanup(database.Cleanup)

	if result := db.Create(&models.Company{Name: "Testing Company"}); result.Error != nil {
		t.Error(result.Error)
	}

	router := api.GetRouter()

	t.Run("Create", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "John Doe",
			"email": "johndoe@email.com",
			"phone": "",
			"cpf":   "297.164.260-70",
			"address": map[string]interface{}{
				"street":       "Street",
				"number":       "211",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "2222222",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var customer models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
			t.Error("Failed to parse JSON: ", err)
		}

		if customer.ID == 0 {
			t.Error("Expected ID, got 0")
		}

		if customer.Name != "John Doe" {
			t.Errorf("Expected name %v, got %v", "John Doe", customer.Name)
		}

		if customer.Address == nil {
			t.Error("Should have an address")
		}

		if customer.Address.City != "New York" {
			t.Errorf("Expected state %v, got %v", "New York", customer.Address.City)
		}

		if customer.Address.State != "NY" {
			t.Errorf("Expected state %v, got %v", "NY", customer.Address.State)
		}
	})

	t.Run("Validation", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "",
			"email": "johndoe",
			"phone": "",
			"cpf":   "",
			"address": map[string]interface{}{
				"street":       "",
				"number":       "",
				"city":         "",
				"state":        "",
				"neighborhood": "",
				"postcode":     "",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Address should not be required", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "Jane Doe",
			"email": "janedoe@email.com",
			"phone": "",
			"cpf":   "007.806.010-92",
			"address": map[string]interface{}{
				"street":       "",
				"number":       "",
				"city":         "",
				"state":        "",
				"neighborhood": "",
				"postcode":     "",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var customer models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
			t.Error("Failed to parse JSON", err)
		}

		if customer.Address != nil {
			t.Errorf("Should not have an address, got %v", customer.Address)
		}
	})

	t.Run("Email should be valid", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "Jane Doe",
			"email": "janedoe",
			"phone": "",
			"cpf":   "249.741.710-54",
			"address": map[string]interface{}{
				"street":       "street",
				"number":       "533",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "75397",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Valid CPF", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "Jane Doe",
			"email": "janedoe@email.com",
			"phone": "",
			"cpf":   "234.214.230-22",
			"address": map[string]interface{}{
				"street":       "street",
				"number":       "533",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "75397",
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

		if response["error"] != api.ErrInvalidCPF.Error() {
			t.Errorf("Expected error %v, got %v", api.ErrInvalidCPF.Error(), response["error"])
		}

		req = Put(t, "/customers/2", map[string]interface{}{
			"name":  "Jane Doe",
			"email": "janedoe@email.com",
			"phone": "",
			"cpf":   "234.214.230-22",
			"address": map[string]interface{}{
				"street":       "street",
				"number":       "533",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "75397",
			},
		})

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Unique email", func(t *testing.T) {
		req := Post(t, "/customers", map[string]interface{}{
			"name":  "John Doe",
			"email": "johndoe@email.com",
			"phone": "",
			"cpf":   "736.151.790-05",
			"address": map[string]interface{}{
				"street":       "Street",
				"number":       "211",
				"city":         "New York",
				"state":        "NY",
				"neighborhood": "Brooklyn",
				"postcode":     "2222222",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %v, got %v", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("List", func(t *testing.T) {
		db.Create(&models.Company{Name: "Other Company"})

		// this customer shouldn't be retrieved
		db.Create(&models.Customer{
			Name:      "From another company",
			CompanyID: 2,
		})

		req := Get(t, "/customers")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var items []models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if len(items) != 2 {
			t.Errorf("Expected %v items, got %v", 2, len(items))
		}

		for idx, customer := range items {
			if idx == 0 && customer.Name != "John Doe" {
				t.Errorf("Expected name %v, got %v", "John Doe", customer.Name)
			}

			if idx == 1 && customer.Name != "Jane Doe" {
				t.Errorf("Expected name %v, got %v", "Jane Doe", customer.Name)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := Get(t, "/customers/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var customer *models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if customer.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, customer.ID)
		}
	})

	t.Run("Get non existing", func(t *testing.T) {
		req := Get(t, "/customers/121")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get invalid", func(t *testing.T) {
		req := Get(t, "/customers/aoeu")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get from another company", func(t *testing.T) {
		req := Get(t, "/customers/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update", func(t *testing.T) {
		req := Put(t, "/customers/1", map[string]interface{}{
			"name":  "Jane Doe 2",
			"email": "janedoe2@email.com",
			"phone": "222222222",
			"cpf":   "774.293.940-19",
			"address": map[string]interface{}{
				"street":       "Street",
				"number":       "211",
				"city":         "Sao Paulo",
				"state":        "SP",
				"neighborhood": "Sao Paulo",
				"postcode":     "333333",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
		}

		var customer *models.Customer
		if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
			t.Error("Failed parsing JSON", err)
		}

		if customer.Name != "Jane Doe 2" {
			t.Errorf("Expected name %v, got %v", "Jane Doe 2", customer.Name)
		}
		if customer.Email != "janedoe2@email.com" {
			t.Errorf("Expected email %v, got %v", "janedoe2@email.com", customer.Email)
		}
	})

	t.Run("Update validation", func(t *testing.T) {
		req := Put(t, "/customers/2", map[string]interface{}{
			"name":  "",
			"email": "janedoe",
			"phone": "222222222",
			"cpf":   "603.411.650-34",
			"address": map[string]interface{}{
				"street":       "",
				"number":       "",
				"city":         "",
				"state":        "",
				"neighborhood": "",
				"postcode":     "",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Update from another company", func(t *testing.T) {
		req := Put(t, "/customers/3", map[string]interface{}{
			"name":  "Jane Doe 2",
			"email": "janedoe2@email.com",
			"phone": "222222222",
			"cpf":   "838.654.570-45",
			"address": map[string]interface{}{
				"street":       "Street",
				"number":       "211",
				"city":         "Sao Paulo",
				"state":        "SP",
				"neighborhood": "Sao Paulo",
				"postcode":     "333333",
			},
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		req := Delete(t, "/customers/1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %v, got %v", http.StatusNoContent, w.Code)
		}

		var customer *models.Customer
		if db.First(&customer, 1).Error == nil {
			t.Error("Should delete customer")
		}
	})

	t.Run("Delete non existent", func(t *testing.T) {
		req := Delete(t, "/customers/1251")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete invalid", func(t *testing.T) {
		req := Delete(t, "/customers/stnh")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete from another company", func(t *testing.T) {
		req := Delete(t, "/customers/3")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %v, got %v", http.StatusNotFound, w.Code)
		}
	})
}
