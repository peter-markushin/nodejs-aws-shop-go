package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang/mock/gomock"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	mock_main "github.com/peterm-itr/nodejs-aws-shop-go/tests/mocks/handlers/importFileParser"
)

func TestHandleUploadedCsvFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	event := events.S3Event{
		Records: []events.S3EventRecord{{
			S3: events.S3Entity{
				Bucket: events.S3Bucket{Name: "test-bucket"},
				Object: events.S3Object{Key: "uploaded/test-object-key"},
			},
		}},
	}

	appConfig := &config.Configuration{
		ImportQueueUrl: "sqs://test",
	}
	s3ClientMock := mock_main.NewMockS3Client(ctrl)
	s3ClientMock.
		EXPECT().
		GetObject(
			gomock.Eq(ctx),
			gomock.Eq(&s3.GetObjectInput{
				Bucket: jsii.String("test-bucket"),
				Key:    jsii.String("uploaded/test-object-key"),
			}),
		).
		Return(
			&s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader("\"aaad9c29-a269-4ce9-9d53-0dd7bd55ea6c\",P001,Test descr,29.90,65")),
			},
			nil,
		).
		Times(1)
	s3ClientMock.
		EXPECT().
		CopyObject(
			gomock.Eq(ctx),
			gomock.Eq(&s3.CopyObjectInput{
				CopySource: jsii.String("test-bucket/uploaded/test-object-key"),
				Bucket:     jsii.String("test-bucket"),
				Key:        jsii.String("parsed/test-object-key"),
			}),
		).
		Return(&s3.CopyObjectOutput{}, nil).
		Times(1)
	s3ClientMock.
		EXPECT().
		DeleteObject(
			gomock.Eq(ctx),
			gomock.Eq(&s3.DeleteObjectInput{
				Bucket: jsii.String("test-bucket"),
				Key:    jsii.String("uploaded/test-object-key"),
			}),
		).
		Return(&s3.DeleteObjectOutput{}, nil).
		Times(1)

	sqsClientMock := mock_main.NewMockSqsClient(ctrl)
	sqsClientMock.
		EXPECT().
		SendMessageBatch(
			gomock.Eq(ctx),
			gomock.Eq(&sqs.SendMessageBatchInput{
				QueueUrl: &appConfig.ImportQueueUrl,
				Entries: []types.SendMessageBatchRequestEntry{
					{MessageBody: jsii.String("{\"id\":\"aaad9c29-a269-4ce9-9d53-0dd7bd55ea6c\",\"title\":\"P001\",\"description\":\"Test descr\",\"price\":\"29.9\",\"count\":65}")},
				},
			}),
		).
		Return(&sqs.SendMessageBatchOutput{}, nil).
		Times(1)

	handler := &UploadedCsvFileHandler{
		config:    appConfig,
		s3Client:  s3ClientMock,
		sqsClient: sqsClientMock,
	}

	err := handler.HandleUploadedCsvFile(ctx, event)

	if err != nil {
		t.Error(err)
	}
}
