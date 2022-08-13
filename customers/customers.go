package customers

import "example.com/accounting/database"

type Customer struct {
	database.Model
	Name    string
	Email   string
	Cpf     string
	Phone   string
	Address *Address `gorm:"embedded"`
}

type Address struct {
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	Postcode     string
}

func Create(name string, email string, cpf string, phone string, address *Address) (*Customer, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	customer := &Customer{
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
	return db.Find(&Customer{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Customer{}).Where("ID", id)
}
