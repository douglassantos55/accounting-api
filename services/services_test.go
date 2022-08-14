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

	t.Run("List", func(t *testing.T) {
		services.Create("Service 2", 1)
		services.Create("Service 3", 1)

		var items []*services.Service
		if err := services.List().Get(&items); err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("Expected %v items, got %v", 3, len(items))
		}
	})

	t.Run("Get by ID", func(t *testing.T) {
		var service *services.Service
		if err := services.Find(3).First(&service); err != nil {
			t.Error(err)
		}

		if service.ID != 3 {
			t.Error("Should retrieve service")
		}
	})

	t.Run("Update", func(t *testing.T) {
		var service *services.Service
		if err := services.Find(3).First(&service); err != nil {
			t.Error(err)
		}

		prevUpdate := service.UpdatedAt
		service.Name = "Updated service"

		if err := services.Update(service); err != nil {
			t.Error(err)
		}

		services.Find(3).First(&service)
		if prevUpdate == service.UpdatedAt {
			t.Error("Should have updated")
		}

		if service.Name != "Updated service" {
			t.Error("Should have updated name")
		}
	})

	t.Run("Update without account", func(t *testing.T) {
		var service *services.Service
		if err := services.Find(3).First(&service); err != nil {
			t.Error(err)
		}

		prevUpdate := service.UpdatedAt
		service.AccountID = 0

		if prevUpdate != service.UpdatedAt {
			t.Error("Should not have updated")
		}

		if err := services.Update(service); err == nil {
			t.Error("Should not have updated without account")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := services.Delete(3); err != nil {
			t.Error(err)
		}

		var service *services.Service
		if err := services.Find(3).First(&service); err == nil {
			t.Error("Should have deleted service")
		}
	})
}
