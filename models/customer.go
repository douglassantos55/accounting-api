package models

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
