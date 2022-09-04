package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name                string  `binding:"required"`
	Price               float64 `binding:"required"`
	Purchasable         bool
	RevenueAccountID    *uint    `json:"revenue_account_id"`
	RevenueAccount      *Account `gorm:"constraint:OnDelete:SET NULL;"`
	CostOfSaleAccountID *uint    `json:"cost_of_sale_account_id"`
	CostOfSaleAccount   *Account `gorm:"constraint:OnDelete:SET NULL;"`
	InventoryAccountID  uint     `json:"inventory_account_id"`
	InventoryAccount    *Account `gorm:"constraint:OnDelete:SET NULL;"`
	VendorID            *uint
	Vendor              *Vendor       `gorm:"constraint:OnDelete:SET NULL;"`
	StockEntries        []*StockEntry `gorm:"constraint:OnDelete:CASCADE;"`
	CompanyID           uint
	Company             *Company
}

func (p Product) Inventory() uint {
	var inventory uint = 0
	for _, entry := range p.StockEntries {
		inventory += entry.Qty
	}
	return inventory
}

type StockEntry struct {
	gorm.Model
	Qty       uint
	Price     float64
	ProductID uint
	Product   *Product
}

type Vendor struct {
	gorm.Model
	Name      string
	Cnpj      string
	CompanyID uint
	Company   *Company
	Address   *Address `gorm:"embedded"`
}
