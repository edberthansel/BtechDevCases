package handler

import (
	"fmt"
	"net/http"

	"github.com/edberthansel/BtechDevCases/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	email := middleware.Email(c)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Hello %s, welcome back", email),
		"id":      middleware.UserID(c),
		"email":   email,
	})
}