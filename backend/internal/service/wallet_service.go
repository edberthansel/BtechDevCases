package service

import (
	"context"

	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/google/uuid"
)

type WalletService struct {
	wallets *repository.WalletRepository
}

func NewWalletService(wallets *repository.WalletRepository) *WalletService {
	return &WalletService{wallets: wallets}
}

func (s *WalletService) GetBalance(ctx context.Context, userID string) (int64, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return 0, err
	}
	return s.wallets.GetBalance(ctx, id)
}