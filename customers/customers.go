package customers

import (
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func Create(name string, email string, cpf string, phone string, address *models.Address) (*models.Customer, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	customer := &models.Customer{
		Name:    name,
		Email:   email,
		Phone:   phone,
		Cpf:     cpf,
		Address: address,
	}
	if err := db.Create(&customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Customer{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Customer{}).Where("ID", id)
}

func Update(customer *models.Customer) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Update(customer)
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Delete(&models.Customer{}, id)
}
