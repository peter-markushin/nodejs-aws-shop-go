package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	sdkConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"log"
)

func main() {
	appConfig, err := config.GetConfig()

	if err != nil {
		log.Println(err.Error())
	}

	awsConfig, err := sdkConfig.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Printf("failed to load default config: %+v", err)

		return
	}

	s3Client := s3.NewFromConfig(awsConfig)
	sqsClient := sqs.NewFromConfig(awsConfig)

	handler := &UploadedCsvFileHandler{
		config:    appConfig,
		s3Client:  s3Client,
		sqsClient: sqsClient,
	}

	lambda.Start(handler.HandleUploadedCsvFile)
}
