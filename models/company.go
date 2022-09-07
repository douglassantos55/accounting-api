package models

import (
	"gorm.io/gorm"
)

type StockOption int

const (
	FIFO StockOption = iota
	LIFO
)

type Company struct {
	gorm.Model
	Name  string
	Stock StockOption
}

type ForCompany struct {
	gorm.Model
	CompanyID uint
	Company   *Company
}

func FromCompany(companyID uint) func(d *gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		return d.Where(&ForCompany{CompanyID: companyID})
	}
}
