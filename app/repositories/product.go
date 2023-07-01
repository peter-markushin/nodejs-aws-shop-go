package repositories

import (
	"github.com/peterm-itr/nodejs-aws-shop-go/db"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"log"
)

var ProductRepositoryImpl IProductRepository

type IProductRepository interface {
	GetAll() ([]models.Product, error)
	GetAvailable() ([]models.Product, error)
	GetByID(id string) (*models.Product, error)
	Add(product *models.Product) (*models.Product, error)
}

type ProductRepository struct {
}

func (h ProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product

	db := db.GetDB()
	result := db.Model(&models.Product{}).Preload("ProductStock").Find(&products).Limit(1000)

	if result.Error != nil {
		log.Println(result.Error)

		return nil, result.Error
	}

	return products, nil
}

func (h ProductRepository) GetAvailable() ([]models.Product, error) {
	var products []models.Product

	db := db.GetDB()
	result := db.Model(&models.Product{}).InnerJoins("ProductStock").Find(&products, "count > 0")

	if result.Error != nil {
		log.Println(result.Error)

		return nil, result.Error
	}

	return products, nil
}

func (h ProductRepository) GetByID(id string) (*models.Product, error) {
	var product *models.Product

	db := db.GetDB()
	result := db.First(&product, "id = ?", id)

	if result.Error != nil {
		log.Println(result.Error)

		return nil, result.Error
	}

	return product, nil
}

func (h ProductRepository) Add(product *models.Product) (*models.Product, error) {
	db := db.GetDB()
	result := db.Create(&product)

	if result.Error != nil {
		log.Println(result.Error)

		return nil, result.Error
	}

	return product, nil
}
