package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/edberthansel/BtechDevCases/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	user, err := h.auth.Register(c.Request.Context(), req.Email, req.Password, req.ConfirmPassword)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": ve.Fields})
			return
		}
		if errors.Is(err, repository.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		log.Printf("register error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	result, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": ve.Fields})
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		log.Printf("login error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     result.Token,
		"tokenType": "Bearer",
		"expiresAt": result.ExpiresAt,
		"user": gin.H{
			"id":    result.User.ID,
			"email": result.User.Email,
		},
	})
}