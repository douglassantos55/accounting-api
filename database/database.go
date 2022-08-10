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
	Where(column string, value interface{}) QueryResult
}

type Repository interface {
	Find(table string) QueryResult
	Create(value interface{}) error
	Update(value interface{}) error
	Delete(value interface{}) error
	Migrate(value interface{}) error
	CleanUp()
}

var connection Repository

func GetConnection() (Repository, error) {
	if connection == nil {
		dns := os.Getenv("DB_CONNECTION")
		driver := os.Getenv("DB_DRIVER")

		repository, err := CreateGormRepository(driver, dns)

		if err != nil {
			return nil, err
		}

		connection = repository
	}

	return connection, nil
}
