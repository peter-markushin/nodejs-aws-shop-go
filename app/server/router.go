package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/peterm-itr/nodejs-aws-shop-go/controllers"
	"github.com/peterm-itr/nodejs-aws-shop-go/docs"
	"github.com/peterm-itr/nodejs-aws-shop-go/repositories"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	docs.SwaggerInfo.BasePath = "/"

	ping := new(controllers.PingController)

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/ping", ping.Status)

	v1 := router.Group("v1")
	{
		productGroup := v1.Group("products")
		{
			productController := controllers.NewProductController(repositories.ProductRepositoryImpl)

			productGroup.GET("", productController.ListProducts)
			productGroup.POST("", productController.AddProduct)
			productGroup.GET("/available", productController.ListAvailableProducts)
			productGroup.GET("/:id", productController.GetProduct)
		}
	}

	return router
}
