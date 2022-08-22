package models

import "example.com/accounting/database"

type Product struct {
	database.Model
	Name                string
	Price               float64
	Purchasable         bool
	RevenueAccountID    *uint
	RevenueAccount      *Account `gorm:"constraint:OnDelete:SET NULL;"`
	ReceivableAccountID *uint
	ReceivableAccount   *Account `gorm:"constraint:OnDelete:SET NULL;"`
	InventoryAccountID  uint
	InventoryAccount    *Account `gorm:"constraint:OnDelete:SET NULL;"`
	VendorID            *uint
	Vendor              *Vendor       `gorm:"constraint:OnDelete:SET NULL;"`
	StockEntries        []*StockEntry `gorm:"constraint:OnDelete:CASCADE;"`
}

func (p Product) Inventory() uint {
	var inventory uint = 0
	for _, entry := range p.StockEntries {
		inventory += entry.Qty
	}
	return inventory
}

type StockEntry struct {
	database.Model
	Qty       uint
	Price     float64
	ProductID uint
	Product   *Product
}

type Vendor struct {
	database.Model
	Name    string
	Cnpj    string
	Address *Address `gorm:"embedded"`
}
