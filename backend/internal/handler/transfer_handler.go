package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/edberthansel/BtechDevCases/backend/internal/middleware"
	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/edberthansel/BtechDevCases/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	transfers *service.TransferService
}

func NewTransferHandler(transfers *service.TransferService) *TransferHandler {
	return &TransferHandler{transfers: transfers}
}

type transferRequest struct {
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
	Notes     string `json:"notes"`
}

func (h *TransferHandler) Transfer(c *gin.Context) {
	var body transferRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body (amount must be a whole number)"})
		return
	}

	result, err := h.transfers.Transfer(c.Request.Context(), service.TransferRequest{
		FromUserID:     middleware.UserID(c),
		Recipient:      body.Recipient,
		Amount:         body.Amount,
		Notes:          body.Notes,
		IdempotencyKey: c.GetHeader("Idempotency-Key"),
	})
	if err != nil {
		var ve *service.ValidationError
		switch {
		case errors.As(err, &ve):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": ve.Fields})
		case errors.Is(err, repository.ErrSelfTransfer):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot transfer to yourself"})
		case errors.Is(err, repository.ErrRecipientNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "recipient not found"})
		case errors.Is(err, repository.ErrInsufficientFunds):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "insufficient funds"})
		case errors.Is(err, repository.ErrIdempotencyConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "idempotency key already used with different data"})
		default:
			log.Printf("transfer error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
		c.Header("Idempotent-Replayed", "true")
	}

	tx := result.Transaction
	c.JSON(status, gin.H{
		"id":        tx.ID,
		"recipient": result.RecipientEmail,
		"amount":    tx.Amount,
		"notes":     tx.Notes,
		"createdAt": tx.CreatedAt,
		"replayed":  result.Replayed,
	})
}

func (h *TransferHandler) History(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := h.transfers.History(c.Request.Context(), middleware.UserID(c), limit, offset)
	if err != nil {
		log.Printf("history error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}