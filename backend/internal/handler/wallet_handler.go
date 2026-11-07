package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/edberthansel/BtechDevCases/backend/internal/middleware"
	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/edberthansel/BtechDevCases/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	wallets *service.WalletService
}

func NewWalletHandler(wallets *service.WalletService) *WalletHandler {
	return &WalletHandler{wallets: wallets}
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	balance, err := h.wallets.GetBalance(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		if errors.Is(err, repository.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
			return
		}
		log.Printf("get balance error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":  balance,
		"currency": "IDR",
	})
}