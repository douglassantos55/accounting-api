package services_test

import (
	"testing"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/services"
)

func TestServices(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_CONNECTION", "../test.sqlite")

	db, _ := database.GetConnection()

	db.Migrate(&services.Service{})
	db.Migrate(&accounts.Account{})

	accounts.Create("Revenue", accounts.Revenue, nil)

	t.Cleanup(db.CleanUp)

	t.Run("Create", func(t *testing.T) {
		service, err := services.Create("Service 1", 1)
		if err != nil {
			t.Error(err)
		}
		if service.ID == 0 {
			t.Error("Should have saved service")
		}
	})

	t.Run("Create without account", func(t *testing.T) {
		if _, err := services.Create("Service 1", 0); err == nil {
			t.Error("Should not create without account")
		}
	})

	t.Run("Create non existing account", func(t *testing.T) {
		if _, err := services.Create("Service 1", 10); err == nil {
			t.Error("Should not create without account")
		}
	})
}
