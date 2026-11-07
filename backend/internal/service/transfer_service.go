package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/edberthansel/BtechDevCases/backend/internal/model"
	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/google/uuid"
)

const (
	MaxTransferAmount = 1_000_000_000
	MaxNotesLength    = 200
)

var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

type TransferRequest struct {
	FromUserID     string
	Recipient      string // email penerima
	Amount         int64
	Notes          string
	IdempotencyKey string
}

type TransferResult struct {
	Transaction    *model.Transaction
	RecipientEmail string
	Replayed       bool
}

type TransferService struct {
	users *repository.UserRepository
	txs   *repository.TransactionRepository
}

func NewTransferService(users *repository.UserRepository, txs *repository.TransactionRepository) *TransferService {
	return &TransferService{users: users, txs: txs}
}

func (s *TransferService) Transfer(ctx context.Context, req TransferRequest) (*TransferResult, error) {
	recipient := strings.ToLower(strings.TrimSpace(req.Recipient))
	notes := strings.TrimSpace(req.Notes)
	key := strings.TrimSpace(req.IdempotencyKey)

	fields := map[string]string{}

	switch {
	case recipient == "":
		fields["recipient"] = "recipient is required"
	case !validEmail(recipient):
		fields["recipient"] = "recipient must be a valid email"
	}

	switch {
	case req.Amount <= 0:
		fields["amount"] = "amount must be a whole number greater than 0"
	case req.Amount > MaxTransferAmount:
		fields["amount"] = "amount exceeds the maximum per transfer"
	}

	if utf8.RuneCountInString(notes) > MaxNotesLength {
		fields["notes"] = "notes must be at most 200 characters"
	}

	if !idempotencyKeyPattern.MatchString(key) {
		fields["Idempotency-Key"] = "header is required: 8-64 characters (letters, digits, - or _)"
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	fromID, err := uuid.Parse(req.FromUserID)
	if err != nil {
		return nil, err
	}

	toUser, err := s.users.FindByEmail(ctx, recipient)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrRecipientNotFound
		}
		return nil, err
	}

	rec, replayed, err := s.txs.Transfer(ctx, repository.TransferInput{
		FromUserID:     fromID,
		ToUserID:       toUser.ID,
		Amount:         req.Amount,
		Notes:          notes,
		IdempotencyKey: key,
	})
	if err != nil {
		return nil, err
	}

	return &TransferResult{
		Transaction:    rec,
		RecipientEmail: toUser.Email,
		Replayed:       replayed,
	}, nil
}

const (
	DefaultHistoryLimit = 20
	MaxHistoryLimit     = 100
)

func (s *TransferService) History(ctx context.Context, userID string, limit, offset int) ([]model.HistoryItem, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}
	if limit > MaxHistoryLimit {
		limit = MaxHistoryLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.txs.ListForUser(ctx, id, limit, offset)
}