package vendors

import (
	"example.com/accounting/customers"
	"example.com/accounting/database"
)

type Vendor struct {
	database.Model
	Name    string
	Cnpj    string
	Address *customers.Address `gorm:"embedded"`
}

func Create(name, cnpj string, address *customers.Address) (*Vendor, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	vendor := &Vendor{
		Name:    name,
		Cnpj:    cnpj,
		Address: address,
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
	return db.Find(&Vendor{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Vendor{}).Where("ID", id)
}

func Update(vendor *Vendor) error {
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
	return db.Delete(&Vendor{}, id)
}
