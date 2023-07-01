package models

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"strconv"
)

type Product struct {
	ID           string          `json:"id" gorm:"primaryKey" binding:"required,uuid"`
	Title        string          `json:"title" binding:"required"`
	Description  string          `json:"description" binding:"omitempty"`
	Price        decimal.Decimal `json:"price" binding:"required,gte=0"`
	ProductStock *Stock          `json:"-" binding:"-"`
}

func NewProductFromCsvRow(in []string) (*Product, error) {
	count, err := strconv.ParseUint(in[4], 10, 0)
	if err != nil {
		return nil, err
	}

	price, err := decimal.NewFromString(in[3])
	if err != nil {
		return nil, err
	}

	return &Product{
		ID:          in[0],
		Title:       in[1],
		Description: in[2],
		Price:       price,
		ProductStock: &Stock{
			ProductID: in[0],
			Count:     uint8(count),
		},
	}, nil
}

func (p *Product) MarshalJSON() ([]byte, error) {
	stockCount := uint8(0)

	if p.ProductStock != nil {
		stockCount = p.ProductStock.Count
	}

	return json.Marshal(&struct {
		ID          string          `json:"id"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Price       decimal.Decimal `json:"price"`
		Count       uint8           `json:"count"`
	}{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Price:       p.Price,
		Count:       stockCount,
	})
}
