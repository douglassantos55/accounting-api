package main

import (
	"log"
	"os"

	"example.com/accounting/api"
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func main() {
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_CONNECTION", "./database.sqlite")

	db, _ := database.GetConnection()
	db.AutoMigrate(
		&models.Account{},
		&models.Customer{},
		&models.Company{},
	)

	router := api.GetRouter()
	log.Fatal(router.Run())
}
