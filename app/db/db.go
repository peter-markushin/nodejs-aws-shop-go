package db

import (
	"fmt"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

func Init(c *config.Configuration) {
	var err error

	dsn := fmt.Sprintf("user=%s password='%s' dbname=%s host=%s port=%s sslmode=require connect_timeout=2", c.DbUser, c.DbPassword, c.DbName, c.DbHost, c.DbPort)
	db, err = gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)

	if err != nil {
		log.Fatalln(err.Error())
	}

	db.AutoMigrate(&models.Product{}, &models.Stock{})
}

func GetDB() *gorm.DB {
	return db
}
