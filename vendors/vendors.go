package vendors

import (
	"example.com/accounting/database"
	"example.com/accounting/models"
)

func Create(companyId uint, name, cnpj string, address *models.Address) (*models.Vendor, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	vendor := &models.Vendor{
		Name:      name,
		Cnpj:      cnpj,
		CompanyID: companyId,
		Address:   address,
	}

	if err := db.Create(vendor); err != nil {
		return nil, err
	}

	return vendor, nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Vendor{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Vendor{}).Where("ID", id)
}

func Update(vendor *models.Vendor) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Update(vendor)
}

func Delete(id uint) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}
	return db.Delete(&models.Vendor{}, id)
}
