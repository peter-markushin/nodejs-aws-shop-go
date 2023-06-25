package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	sdkConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/db"
	"github.com/peterm-itr/nodejs-aws-shop-go/repositories"
	"log"
)

func main() {
	appConfig, err := config.GetConfig()

	if err != nil {
		log.Printf("Error loading configuration: %+v", err)

		return
	}

	db.Init(appConfig)
	log.Println("DB Init complete")

	repositories.Init()
	log.Println("Repo Init complete")

	sdkCfg, err := sdkConfig.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Printf("failed to load default SDK config: %+v", err)

		return
	}

	snsClient := sns.NewFromConfig(sdkCfg)

	handler := &CatalogBatchProcessHandler{
		config:            appConfig,
		snsClient:         snsClient,
		productRepository: repositories.ProductRepositoryImpl,
	}

	lambda.Start(handler.Handler)
}
