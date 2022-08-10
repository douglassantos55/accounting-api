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

type QueryResult interface {
	Get(dest interface{}) error
	With(relation string) QueryResult
}

type Repository interface {
	Find(table string) QueryResult
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
