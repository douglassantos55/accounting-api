package accounts

import (
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func Create(companyId uint, name string, accType models.AccountType, parentID *uint) (*models.Account, error) {
	db, err := database.GetConnection()

	if err != nil {
		return nil, err
	}

	account := &models.Account{
		Name:      name,
		Type:      accType,
		ParentID:  parentID,
		CompanyID: companyId,
	}

	if err := db.Create(account); err != nil {
		return nil, err
	}

	return account, nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()

	if err != nil {
		return nil
	}

	return db.Find(&models.Account{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Account{}).Where("ID", id)
}

func Update(account *models.Account) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Update(account)
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	if err := Find(id).First(&models.Account{}); err != nil {
		return err
	}
	return db.Delete(&models.Account{}, id)
}
