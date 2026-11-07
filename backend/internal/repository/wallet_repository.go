package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrWalletNotFound = errors.New("wallet not found")

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) GetBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx,
		`SELECT balance FROM wallets WHERE user_id = $1`,
		userID,
	).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrWalletNotFound
	}
	if err != nil {
		return 0, err
	}
	return balance, nil
}