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
