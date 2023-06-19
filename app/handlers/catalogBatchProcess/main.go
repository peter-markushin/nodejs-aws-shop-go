package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	sdkConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/controllers/DTO"
	"github.com/peterm-itr/nodejs-aws-shop-go/db"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"github.com/peterm-itr/nodejs-aws-shop-go/repositories"
	"log"
)

var configuration *config.Configuration

func main() {
	var err error
	configuration, err = config.GetConfig()

	if err != nil {
		log.Println(err.Error())
	}

	db.Init(configuration)
	log.Println("DB Init complete")

	repositories.Init()
	log.Println("Repo Init complete")

	lambda.Start(Handler)
}

func Handler(ctx context.Context, event events.SQSEvent) error {
	importedProducts := 0
	for _, message := range event.Records {
		var dto DTO.ProductRequest

		if err := json.Unmarshal([]byte(message.Body), &dto); err != nil {
			return err
		}

		productId := uuid.NewString()
		newProduct := &models.Product{
			ID:          productId,
			Title:       dto.Title,
			Description: dto.Description,
			Price:       dto.Price,
			ProductStock: &models.Stock{
				ProductID: productId,
				Count:     dto.Count,
			},
		}

		_, err := repositories.ProductRepositoryImpl.Add(newProduct)

		if err != nil {
			return err
		}

		importedProducts++
	}

	msg := fmt.Sprintf("Imported %d products", importedProducts)
	cfg, err := sdkConfig.LoadDefaultConfig(ctx)

	if err != nil {
		return err
	}

	client := sns.NewFromConfig(cfg)
	snsInput := &sns.PublishInput{
		Message:  &msg,
		TopicArn: &configuration.ImportNotificationTopic,
	}

	_, err = client.Publish(ctx, snsInput)

	if err != nil {
		return err
	}

	return nil
}
