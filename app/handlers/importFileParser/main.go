package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	sdkConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
)

var configuration *config.Configuration

func main() {
	var err error
	configuration, err = config.GetConfig()

	if err != nil {
		log.Println(err.Error())
	}

	lambda.Start(Handler)
}

func Handler(ctx context.Context, event events.S3Event) error {
	cfg, err := sdkConfig.LoadDefaultConfig(ctx)

	if err != nil {
		log.Printf("failed to load default config: %s", err)
		return err
	}

	s3Client := s3.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	for _, record := range event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		getOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})

		if err != nil {
			log.Printf("error getting object %s/%s: %s", bucket, key, err)
			return err
		}

		csvReader := csv.NewReader(getOutput.Body)

		for {
			row, err := csvReader.Read()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("error reading csv row: %s", err)

				return err
			}

			log.Printf("CSV row: %s", row)

			product, err := models.NewProductFromCsvRow(row)

			if err != nil {
				log.Printf("error creating product from csv row: %s", err)

				continue
			}

			productJson, err := json.Marshal(product)
			messageBody := string(productJson)

			if err != nil {
				log.Printf("error marshalling object ot json: %s", err)

				continue
			}

			_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				QueueUrl:               &configuration.ImportQueueUrl,
				MessageBody:            &messageBody,
				DelaySeconds:           0,
				MessageDeduplicationId: nil,
			})

			if err != nil {
				log.Printf("error sending message to queue: %s", err)
			}
		}

		src := fmt.Sprintf("%s/%s", bucket, key)
		dst := strings.Replace(key, "uploaded/", "parsed/", 1)
		log.Println(src, bucket, dst)

		_, err = s3Client.CopyObject(ctx, &s3.CopyObjectInput{
			CopySource: &src,
			Bucket:     &bucket,
			Key:        &dst,
		})
		if err != nil {
			log.Printf("Error copying file: %+v", err)

			continue
		}

		_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			log.Printf("Error removing file: %s", err)
		}
	}

	return nil
}
