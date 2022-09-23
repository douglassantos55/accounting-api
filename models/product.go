package models

import (
	"math"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name                string  `binding:"required"`
	Price               float64 `binding:"required"`
	Purchasable         bool
	RevenueAccountID    *uint    `binding:"required_if=Purchasable true"`
	RevenueAccount      *Account `gorm:"constraint:OnDelete:SET NULL;"`
	CostOfSaleAccountID *uint    `binding:"required_if=Purchasable true"`
	CostOfSaleAccount   *Account `gorm:"constraint:OnDelete:SET NULL;"`
	InventoryAccountID  uint     `binding:"required"`
	InventoryAccount    *Account `gorm:"constraint:OnDelete:SET NULL;"`
	VendorID            *uint
	Vendor              *Vendor       `gorm:"constraint:OnDelete:SET NULL;"`
	StockEntries        []*StockEntry `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	CompanyID           uint          `json:"-"`
	Company             *Company      `json:"-"`
}

func (p *Product) Inventory() uint {
	var inventory uint = 0
	for _, entry := range p.StockEntries {
		inventory += entry.Stock()
	}
	return inventory
}

func (p *Product) Consume(qty uint) []*StockUsage {
	left := qty
	var usages []*StockUsage

	// Invert entries for LIFO
	if p.Company.Stock == LIFO {
		for i, j := 0, len(p.StockEntries)-1; i < j; i, j = i+1, j-1 {
			p.StockEntries[i], p.StockEntries[j] = p.StockEntries[j], p.StockEntries[i]
		}
	}

	for _, entry := range p.StockEntries {
		qty := math.Min(float64(left), float64(entry.Qty))

		usages = append(usages, &StockUsage{
			Qty:          uint(qty),
			StockEntryID: entry.ID,
		})

		left -= uint(qty)

		if left <= 0 {
			break
		}
	}

	return usages
}

func (p *Product) Cost(qty uint) float64 {
	cost := 0.0
	left := qty

	// Invert entries for LIFO
	if p.Company.Stock == LIFO {
		for i, j := 0, len(p.StockEntries)-1; i < j; i, j = i+1, j-1 {
			p.StockEntries[i], p.StockEntries[j] = p.StockEntries[j], p.StockEntries[i]
		}
	}

	for _, entry := range p.StockEntries {
		qty := math.Min(float64(left), float64(entry.Qty))
		cost += entry.Price * qty
		left -= uint(qty)

		if left <= 0 {
			break
		}
	}

	return cost
}

type StockEntry struct {
	gorm.Model
	Qty         uint
	Price       float64
	ProductID   uint
	Product     *Product
	StockUsages []*StockUsage `gorm:"constraint:OnDelete:CASCADE"`
}

func (e *StockEntry) Stock() uint {
	used := uint(0)
	for _, usage := range e.StockUsages {
		used += usage.Qty
	}
	return e.Qty - used
}

type StockUsage struct {
	gorm.Model
	Qty          uint
	SourceID     uint
	SourceType   string
	StockEntryID uint
	StockEntry   *StockEntry `gorm:"constraint:OnDelete:CASCADE"`
}

type Vendor struct {
	gorm.Model
	Name      string `binding:"required"`
	Cnpj      string `binding:"required,cpf_cnpj,unique"`
	CompanyID uint
	Company   *Company
	Address   *Address `gorm:"embedded"`
}
