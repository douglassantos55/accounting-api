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
