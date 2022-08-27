package models

import "example.com/accounting/database"

type Service struct {
	database.Model
	Name      string
	AccountID uint
	Account   *Account
	CompanyID uint
	Company   *Company
}
