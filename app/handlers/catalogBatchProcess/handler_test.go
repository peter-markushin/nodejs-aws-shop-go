package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/jsii-runtime-go"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	mock_main "github.com/peterm-itr/nodejs-aws-shop-go/tests/mocks/handlers/catalogBatchProcess"
	mock_repositories "github.com/peterm-itr/nodejs-aws-shop-go/tests/mocks/repositories"
	"github.com/shopspring/decimal"
)

func TestCatalogBatchProcess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	event := events.SQSEvent{Records: []events.SQSMessage{{
		Body: "{\"title\":\"Test\",\"description\":\"Description test\",\"price\":\"22.99\",\"count\":22}",
	}}}

	appConfig := &config.Configuration{
		ImportNotificationTopic: "TEST",
	}

	productId := uuid.NewString()
	newProduct := models.Product{
		ID:          productId,
		Title:       "Test",
		Description: "Description test",
		Price:       decimal.RequireFromString("22.99"),
		ProductStock: &models.Stock{
			ProductID: productId,
			Count:     22,
		},
	}

	productRepositoryMock := mock_repositories.NewMockIProductRepository(ctrl)
	productRepositoryMock.
		EXPECT().
		Add(gomock.Any()).
		Return(&newProduct, nil).
		Times(1)

	snsClientMock := mock_main.NewMockSnsClient(ctrl)
	snsClientMock.
		EXPECT().
		Publish(gomock.Eq(ctx), gomock.Eq(&sns.PublishInput{
			Message:  jsii.String(fmt.Sprintf("{\"id\":\"%s\",\"title\":\"Test\",\"description\":\"Description test\",\"price\":\"22.99\",\"count\":22}", productId)),
			TopicArn: jsii.String("TEST"),
			MessageAttributes: map[string]types.MessageAttributeValue{"price": {
				DataType:    jsii.String("Number"),
				StringValue: jsii.String("22.99"),
			}},
		})).
		Return(&sns.PublishOutput{}, nil).
		Times(1)

	handler := CatalogBatchProcessHandler{
		config:            appConfig,
		snsClient:         snsClientMock,
		productRepository: productRepositoryMock,
	}

	err := handler.Handler(ctx, event)

	if err != nil {
		panic(err)
	}
}
