package products

import (
	"errors"

	"example.com/accounting/database"
	"example.com/accounting/models"
)

var (
	ErrRevenueAccountMissing    = errors.New("Revenue account is required")
	ErrReceivableAccountMissing = errors.New("Receivable account is required")
	ErrInventoryAccountMissing  = errors.New("Inventory account is required")
)

func Create(product *models.Product) error {
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
	return db.Find(&models.Product{})
}

func Find(id uint) database.QueryResult {
	db, err := database.GetConnection()
	if err != nil {
		return nil
	}
	return db.Find(&models.Product{}).Where("ID", id)
}

func Update(product *models.Product) error {
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

	if err := Find(id).First(&models.Product{}); err != nil {
		return err
	}

	return db.Delete(&models.Product{}, id)
}

func validateAccounts(product *models.Product) error {
	if product.Purchasable {
		if product.RevenueAccountID == nil {
			return ErrRevenueAccountMissing
		}

		if product.ReceivableAccountID == nil {
			return ErrReceivableAccountMissing
		}
	}

	return nil
}
