package products

import (
	"errors"

	"example.com/accounting/accounts"
	"example.com/accounting/database"
	"example.com/accounting/vendors"
)

var (
	ErrRevenueAccountMissing    = errors.New("Revenue account is required")
	ErrReceivableAccountMissing = errors.New("Receivable account is required")
	ErrInventoryAccountMissing  = errors.New("Inventory account is required")
)

type StockEntry struct {
	Qty       uint
	Price     float64
	ProductID uint
	Product   *Product
}

type Product struct {
	database.Model
	Name                string
	Price               float64
	Purchasable         bool
	ManageStock         bool
	RevenueAccountID    *uint
	RevenueAccount      *accounts.Account
	ReceivableAccountID *uint
	ReceivableAccount   *accounts.Account
	InventoryAccountID  *uint
	InventoryAccount    *accounts.Account
	VendorID            *uint
	Vendor              *vendors.Vendor
	StockEntries        []*StockEntry
}

func Create(product *Product) error {
	db, err := database.GetConnection()

	if err != nil {
		return err
	}

	if err := validateAccounts(product); err != nil {
		return err
	}

	if err := db.Create(&product); err != nil {
		return err
	}

	return nil
}

func List() database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Product{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&Product{}).Where("ID", id)
}

func Update(product *Product) error {
	db, err := database.GetConnection()

	if err != nil {
		return nil
	}

	if err := validateAccounts(product); err != nil {
		return err
	}

	return db.Update(product)
}

func Delete(id uint) error {
	db, err := database.GetConnection()

	if err != nil {
		return err
	}

	if err := Find(id).First(&Product{}); err != nil {
		return err
	}

	return db.Delete(&Product{}, id)
}

func validateAccounts(product *Product) error {
	if product.Purchasable {
		if product.RevenueAccountID == nil {
			return ErrRevenueAccountMissing
		}

		if product.ReceivableAccountID == nil {
			return ErrReceivableAccountMissing
		}
	}

	if product.ManageStock && product.InventoryAccountID == nil {
		return ErrInventoryAccountMissing
	}

	return nil
}
