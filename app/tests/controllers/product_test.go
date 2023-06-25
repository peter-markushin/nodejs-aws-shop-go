package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/peterm-itr/nodejs-aws-shop-go/models"
	"github.com/peterm-itr/nodejs-aws-shop-go/repositories"
	"github.com/peterm-itr/nodejs-aws-shop-go/server"
	mock_repositories "github.com/peterm-itr/nodejs-aws-shop-go/tests/mocks/repositories"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	products := []models.Product{
		{
			ID:          uuid.NewString(),
			Title:       "P1",
			Description: "D1",
			Price:       decimal.New(22249, -2),
		},
		{
			ID:          uuid.NewString(),
			Title:       "P2",
			Description: "D2",
			Price:       decimal.New(1199, -2),
		},
	}

	repositoryMock := mock_repositories.NewMockIProductRepository(ctrl)
	repositoryMock.
		EXPECT().
		GetAll().
		Return(products, nil).
		Times(1)

	repositories.ProductRepositoryImpl = repositoryMock

	router := server.NewRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"GET",
		"/v1/products",
		nil,
	)
	router.ServeHTTP(w, req)

	productsJson, _ := json.Marshal(products)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t,
		string(productsJson),
		w.Body.String(),
	)
}

func TestGetProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productId := uuid.New().String()
	product := &models.Product{
		ID:          productId,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       decimal.New(22249, -2),
	}

	repositoryMock := mock_repositories.NewMockIProductRepository(ctrl)
	repositoryMock.
		EXPECT().
		GetByID(gomock.Eq(productId)).
		Return(product, nil).
		Times(1)

	repositories.ProductRepositoryImpl = repositoryMock

	router := server.NewRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("/v1/products/%s", productId),
		nil,
	)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t,
		fmt.Sprintf("{\"id\":\"%s\",\"title\":\"Test Product\",\"description\":\"Test Description\",\"price\":\"222.49\",\"count\":0}", productId),
		w.Body.String(),
	)
}

func TestProductNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productId := uuid.New().String()

	repositoryMock := mock_repositories.NewMockIProductRepository(ctrl)
	repositoryMock.
		EXPECT().
		GetByID(gomock.Eq(productId)).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	repositories.ProductRepositoryImpl = repositoryMock

	router := server.NewRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("/v1/products/%s", productId),
		nil,
	)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Equal(t,
		"{\"code\":404,\"message\":\"record not found\"}",
		w.Body.String(),
	)
}
