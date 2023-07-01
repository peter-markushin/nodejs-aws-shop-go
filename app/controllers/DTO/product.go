package DTO

import (
	"github.com/shopspring/decimal"
)

type ProductRequest struct {
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description" binding:"omitempty"`
	Price       decimal.Decimal `json:"price" binding:"required,gte=0"`
	Count       uint8           `json:"count" binding:"required,gte=0"`
}
