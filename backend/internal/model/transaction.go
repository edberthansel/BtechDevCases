package model

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID             uuid.UUID
	FromUserID     uuid.UUID
	ToUserID       uuid.UUID
	Amount         int64
	Notes          string
	IdempotencyKey string
	CreatedAt      time.Time
}

// HistoryItem adalah transaksi dari sudut pandang satu user.
type HistoryItem struct {
	ID                uuid.UUID `json:"id"`
	Direction         string    `json:"direction"` // "in" atau "out"
	CounterpartyEmail string    `json:"counterparty"`
	Amount            int64     `json:"amount"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"createdAt"`
}