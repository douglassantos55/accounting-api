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
		&models.Vendor{},
		&models.Product{},
		&models.Service{},
		&models.Purchase{},
		&models.StockEntry{},
		&models.Transaction{},
		&models.Entry{},
		&models.Sale{},
		&models.Item{},
	)

	api.RegisterEvents()

	router := api.GetRouter()
	log.Fatal(router.Run())
}
