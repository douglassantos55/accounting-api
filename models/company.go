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
