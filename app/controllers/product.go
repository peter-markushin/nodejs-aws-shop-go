package controllers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/peterm-itr/nodejs-aws-shop-go/controllers/DTO"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"gorm.io/gorm"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/peterm-itr/nodejs-aws-shop-go/httputil"
	"github.com/peterm-itr/nodejs-aws-shop-go/repositories"
)

type ProductController struct {
	repository repositories.IProductRepository
}

func NewProductController(repository repositories.IProductRepository) *ProductController {
	return &ProductController{
		repository: repository,
	}
}

// @BasePath /prod

// ListProducts godoc
// @Summary		List products
// @Description	List all products
// @ID			api.v1.products.list
// @Tags		Products
// @Accept		json-api
// @Produce		json-api
// @Success		200	{object}	[]DTO.ProductResponse
// @Failure		500	{object}	httputil.HTTPError
// @Router		/v1/products	[get]
func (u ProductController) ListProducts(c *gin.Context) {
	log.Println("List Products")

	products, err := u.repository.GetAll()

	if err != nil {
		log.Println(err.Error())
		httputil.NewError(c, http.StatusInternalServerError, err)
		c.Abort()

		return
	}

	c.JSON(http.StatusOK, products)

	return
}

// ListAvailableProducts godoc
// @Summary		List available products
// @Description	List only available products
// @ID			api.v1.products.available
// @Tags		Products
// @Accept		json-api
// @Produce		json-api
// @Success		200	{object}	[]DTO.ProductResponse
// @Failure		500	{object}	httputil.HTTPError
// @Router		/v1/products	[get]
func (u ProductController) ListAvailableProducts(c *gin.Context) {
	log.Println("List Available Products")

	products, err := u.repository.GetAvailable()

	if err != nil {
		log.Println(err.Error())
		httputil.NewError(c, http.StatusInternalServerError, err)
		c.Abort()

		return
	}

	c.JSON(http.StatusOK, products)

	return
}

// GetProduct godoc
// @Summary 	Get product
// @Description Get product by ID
// @ID 			api.v1.products.get
// @Tags 		Products
// @Accept 		json-api
// @Produce 	json-api
// @Param 		id 	path	string	true	"Product ID"
// @Success 	200	{object}	DTO.ProductResponse
// @Failure		400	{object}	httputil.HTTPError
// @Failure		500	{object}	httputil.HTTPError
// @Router		/v1/products/{id}	[get]
func (u ProductController) GetProduct(c *gin.Context) {
	log.Println("Get Product", fmt.Sprintf("%+v", c.Params))

	if _, err := uuid.Parse(c.Param("id")); err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid product id"))
		c.Abort()

		return
	}

	product, err := u.repository.GetByID(c.Param("id"))

	if errors.Is(err, gorm.ErrRecordNotFound) {
		httputil.NewError(c, http.StatusNotFound, err)
		c.Abort()

		return
	}

	if err != nil {
		log.Println(err.Error())
		httputil.NewError(c, http.StatusInternalServerError, err)
		c.Abort()

		return
	}

	c.JSON(http.StatusOK, product)

	return
}

// AddProduct godoc
// @Summary 	Add product
// @Description Add new product
// @ID 			api.v1.products.add
// @Tags 		Products
// @Accept 		json-api
// @Produce 	json-api
// @Param 		request	body	DTO.ProductRequest true "product data"
// @Success 	200	{object}	DTO.ProductResponse
// @Failure		400	{object}	httputil.HTTPError
// @Failure		422	{object}	httputil.HTTPError
// @Failure		500	{object}	httputil.HTTPError
// @Router		/v1/products	[post]
func (u ProductController) AddProduct(c *gin.Context) {
	var productDto DTO.ProductRequest

	if err := c.ShouldBindJSON(&productDto); err != nil {
		httputil.NewError(c, http.StatusUnprocessableEntity, err)
		c.Abort()

		return
	}

	log.Println("Add Product", fmt.Sprintf("%+v", productDto))

	productId := uuid.NewString()
	newProduct := &models.Product{
		ID:          productId,
		Title:       productDto.Title,
		Description: productDto.Description,
		Price:       productDto.Price,
		ProductStock: &models.Stock{
			ProductID: productId,
			Count:     productDto.Count,
		},
	}

	product, err := u.repository.Add(newProduct)

	if err != nil {
		log.Println(err.Error())
		httputil.NewError(c, http.StatusInternalServerError, err)
		c.Abort()

		return
	}

	c.JSON(http.StatusCreated, product)

	return
}
