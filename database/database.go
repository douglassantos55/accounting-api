package database

import (
	"os"
	"time"
)

type Model struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Repository interface {
	Create(value interface{}) error
	Update(value interface{}) error
	Delete(value interface{}) error
	Migrate(value interface{}) error
	CleanUp()
}

func GetConnection() (Repository, error) {
	dns := os.Getenv("DB_CONNECTION")
	driver := os.Getenv("DB_DRIVER")

	return CreateGormRepository(driver, dns)
}
