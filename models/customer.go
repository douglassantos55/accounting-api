package models

import (
	"gorm.io/gorm"
)

type Customer struct {
	gorm.Model
	Name      string `binding:"required"`
	Email     string `binding:"omitempty,email"`
	Cpf       string `binding:"required"`
	Phone     string
	Address   *Address `gorm:"embedded"`
	CompanyID uint
	Company   *Company
}

type Address struct {
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	Postcode     string
}
