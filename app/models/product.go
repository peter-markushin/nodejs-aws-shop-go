package models

import "github.com/shopspring/decimal"

type Product struct {
	ID           string          `json:"id" gorm:"primaryKey" binding:"required,uuid"`
	Title        string          `json:"title" binding:"required"`
	Description  string          `json:"description" binding:"omitempty"`
	Price        decimal.Decimal `json:"price" binding:"required,gte=0"`
	ProductStock Stock           `json:"-" binding:"-"`
}
