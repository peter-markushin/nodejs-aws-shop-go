package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @BasePath /prod

type PingController struct{}

// Ping godoc
// @Summary ping
// @Schemes
// @Description do ping
// @Tags Ping
// @Produce plain
// @Success 200 {string} pong
// @Router /ping [get]
func (p PingController) Status(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
