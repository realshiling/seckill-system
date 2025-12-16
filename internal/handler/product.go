package handler

import (
	"net/http"
	"seckill-system/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	ProductService *service.ProductService
}

// 创建产品
func (h *ProductHandler) Create(c *gin.Context) {
	var req struct {
		Name  string  `json:"name"`
		Stock int     `json:"stock"`
		Price float64 `json:"price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err := h.ProductService.Create(req.Name, req.Stock, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "product created successfully"})
}

// 列出所有产品
func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.ProductService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, products)
}
