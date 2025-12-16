package handler

import (
	"fmt"
	"net/http"
	"seckill-system/internal/model"
	"seckill-system/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB // 假设你使用 GORM 作为 ORM
}

func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "hash error"})
		return
	}

	user := model.User{
		Username: req.Username,
		Password: hashed,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Database error,用户已存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "register success"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	c.Bind(&req)

	var user model.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	if !utils.ComparePassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "token generate failed", "detail": err.Error()})
		return
	}

	fmt.Println("generated token:", token)
	c.JSON(http.StatusOK, gin.H{"token": token})
}
