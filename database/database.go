package database

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var connection *gorm.DB

func GetConnection() (*gorm.DB, error) {
	if connection == nil {
		dns := os.Getenv("DB_CONNECTION")
		driver := os.Getenv("DB_DRIVER")

		db, err := gorm.Open(getDialector(driver, dns), &gorm.Config{
			FullSaveAssociations: true,
		})

		if err != nil {
			return nil, err
		}

		if driver == "sqlite" {
			db.Exec("PRAGMA foreign_keys = ON")
		}

		connection = db
	}

	return connection, nil
}

func getDialector(driver string, dns string) gorm.Dialector {
	switch driver {
	case "sqlite":
		return sqlite.Open(dns)
	default:
		return nil
	}
}

func Cleanup() {
	migrator := connection.Migrator()
	tables, _ := migrator.GetTables()

	for _, table := range tables {
		migrator.DropTable(table)
	}
}
