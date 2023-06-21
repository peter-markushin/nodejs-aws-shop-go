package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"io"
	"log"
	"strings"
)

type S3Client interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	CopyObject(ctx context.Context, input *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type SqsClient interface {
	SendMessageBatch(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
}

type UploadedCsvFileHandler struct {
	config    *config.Configuration
	s3Client  S3Client
	sqsClient SqsClient
}

func (h UploadedCsvFileHandler) HandleUploadedCsvFile(ctx context.Context, event events.S3Event) error {
	messagesToSend := []types.SendMessageBatchRequestEntry{}

	for _, record := range event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		getOutput, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
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

			log.Printf("CSV row: %+v", row)

			product, err := models.NewProductFromCsvRow(row)

			if err != nil {
				log.Printf("error creating product from csv row: %s", err)

				continue
			}

			productJson, err := json.Marshal(product)
			messagesToSend = append(
				messagesToSend,
				types.SendMessageBatchRequestEntry{MessageBody: jsii.String(string(productJson))},
			)

			if err != nil {
				log.Printf("error marshalling object ot json: %s", err)

				continue
			}
		}

		_, err = h.sqsClient.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
			Entries:  messagesToSend,
			QueueUrl: &h.config.ImportQueueUrl,
		})

		if err != nil {
			log.Printf("error sending message to queue: %s", err)
		}

		src := fmt.Sprintf("%s/%s", bucket, key)
		dst := strings.Replace(key, "uploaded/", "parsed/", 1)
		log.Println(src, bucket, dst)

		_, err = h.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
			CopySource: &src,
			Bucket:     &bucket,
			Key:        &dst,
		})
		if err != nil {
			log.Printf("Error copying file: %+v", err)

			continue
		}

		_, err = h.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			log.Printf("Error removing file: %s", err)
		}
	}

	return nil
}
