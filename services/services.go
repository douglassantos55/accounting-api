package services

import (
	"example.com/accounting/accounts"
	"example.com/accounting/database"
)

type Service struct {
	database.Model
	Name      string
	AccountID uint
	Account   *accounts.Account
}

func Create(name string, accountID uint) (*Service, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	service := &Service{
		Name:      name,
		AccountID: accountID,
	}

	if err := db.Create(service); err != nil {
		return nil, err
	}

	return service, nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Service{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Service{}).Where("ID", id)
}

func Update(service *Service) error {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Update(service)
}
