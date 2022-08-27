package services

import (
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func Create(companyId uint, name string, accountID uint) (*models.Service, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	service := &models.Service{
		Name:      name,
		CompanyID: companyId,
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
	return db.Find(&models.Service{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Service{}).Where("ID", id)
}

func Update(service *models.Service) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Update(service)
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Delete(&models.Service{}, id)
}
