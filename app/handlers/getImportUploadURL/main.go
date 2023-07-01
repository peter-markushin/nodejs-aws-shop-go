package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/controllers"
	"log"
)

var ginLambda *ginadapter.GinLambda

func main() {
	configuration, err := config.GetConfig()

	if err != nil {
		log.Println(err.Error())
	}

	importController := controllers.NewImportController(configuration)

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	router.GET("*any", importController.GetSignedImportUrl)

	ginLambda = ginadapter.New(router)
	lambda.Start(Handler)
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, request)
}
