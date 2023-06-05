package server

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
)

var ginLambda *ginadapter.GinLambda

func Init(c *config.Configuration) {
	r := NewRouter()

	if isInLambda() {
		ginLambda = ginadapter.New(r)
		lambda.Start(Handler)
	} else {
		r.Run(c.AppPort)
	}
}

func isInLambda() bool {
	lambdaTaskRoot := os.Getenv("LAMBDA_TASK_ROOT")

	return lambdaTaskRoot != ""
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, request)
}
