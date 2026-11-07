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