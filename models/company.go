package models

import "example.com/accounting/database"

type StockOption int

const (
	FIFO StockOption = iota
	LIFO
)

type Company struct {
	database.Model
	Name  string
	Stock StockOption
}
