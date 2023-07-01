package controllers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/httputil"
	"log"
	"net/http"
	"time"
)

type ImportController struct {
	config *config.Configuration
}

func NewImportController(c *config.Configuration) *ImportController {
	return &ImportController{
		config: c,
	}
}

// @BasePath /prod

// GetSignedImportUrl godoc
// @Summary 	Get pre-signed S3 URL
// @Description Get pre-signed S3 URL that can be used to upload product list for import
// @ID 			api.v1.import.get-url
// @Tags 		Products
// @Accept 		json-api
// @Produce 	json-api
// @Param 		name	query	string	true	"File name that is uploaded"
// @Success 	200	{object}	controllers.GetSignedImportUrl.response
// @Failure		400	{object}	httputil.HTTPError
// @Failure		500	{object}	httputil.HTTPError
// @Router		/import	[get]
func (i ImportController) GetSignedImportUrl(c *gin.Context) {
	fileName := c.Query("name")

	if fileName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, &httputil.HTTPError{
			Code:    400,
			Message: "Name should not be empty",
		})
		return
	}

	bucketKey := fmt.Sprintf("uploaded/%s", fileName)
	awsSession, _ := session.NewSession()
	s3svc := s3.New(awsSession)

	resp, _ := s3svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(i.config.S3BucketName),
		Key:    aws.String(bucketKey),
	})

	url, err := resp.Presign(10 * time.Minute)

	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		c.Abort()
		return
	}

	log.Println(url)

	c.JSON(http.StatusCreated, struct {
		Link string `json:"link"`
	}{Link: url})

	return
}
