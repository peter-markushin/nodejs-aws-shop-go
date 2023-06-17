package repositories

import (
	"github.com/peterm-itr/nodejs-aws-shop-go/db"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"log"
)

var StockReposityryImpl IStockRepository

type IStockRepository interface {
	GetByID(id string) (*models.Stock, error)
}

type StockRepository struct {
}

func (h StockRepository) GetByID(id string) (*models.Stock, error) {
	var stock *models.Stock

	db := db.GetDB()
	result := db.First(&stock, "id = ?", id)

	if result.Error != nil {
		log.Println(result.Error)

		return nil, result.Error
	}

	return stock, nil
}
