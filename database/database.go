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
	First(dest interface{}) error
	With(relation string) QueryResult
	Where(condition string, value interface{}) QueryResult
	WhereHas(relation, condition string, value interface{}) QueryResult
}

type Repository interface {
	Find(model interface{}) QueryResult
	Create(model interface{}) error
	Update(model interface{}) error
	Delete(model interface{}, id uint) error
	Migrate(model interface{}) error
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
