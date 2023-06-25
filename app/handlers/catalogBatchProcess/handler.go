package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/uuid"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/controllers/DTO"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"github.com/peterm-itr/nodejs-aws-shop-go/repositories"
)

type SnsClient interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type CatalogBatchProcessHandler struct {
	config            *config.Configuration
	snsClient         SnsClient
	productRepository repositories.IProductRepository
}

func (c *CatalogBatchProcessHandler) Handler(ctx context.Context, event events.SQSEvent) error {
	for _, message := range event.Records {
		var dto DTO.ProductRequest

		if err := json.Unmarshal([]byte(message.Body), &dto); err != nil {
			log.Printf("Unmarshalling failed: %+v", err)

			continue
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

		newProduct, err := c.productRepository.Add(newProduct)

		if err != nil {
			log.Printf("Adding product failed: %+v", err)

			continue
		}

		productJson, err := json.Marshal(newProduct)

		if err != nil {
			log.Printf("Json marshalling failed: %+v", err)

			continue
		}

		_, err = c.snsClient.Publish(ctx, &sns.PublishInput{
			Message:  jsii.String(string(productJson)),
			TopicArn: &c.config.ImportNotificationTopic,
			MessageAttributes: map[string]types.MessageAttributeValue{"price": {
				DataType:    jsii.String("Number"),
				StringValue: jsii.String(newProduct.Price.String()),
			}},
		})

		if err != nil {
			log.Printf("Sending notification failed: %+v", err)
		}
	}

	return nil
}
