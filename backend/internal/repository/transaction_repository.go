package repository

import (
	"context"
	"errors"

	"github.com/edberthansel/BtechDevCases/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrRecipientNotFound   = errors.New("recipient not found")
	ErrSelfTransfer        = errors.New("cannot transfer to yourself")
	ErrIdempotencyConflict = errors.New("idempotency key already used with different data")
)

type TransferInput struct {
	FromUserID     uuid.UUID
	ToUserID       uuid.UUID
	Amount         int64
	Notes          string
	IdempotencyKey string
}

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

// Transfer memindahkan uang dalam satu transaksi database.
// Mengembalikan transaksi, dan true kalau itu hasil ulang dari idempotency key yang sama.
func (r *TransactionRepository) Transfer(ctx context.Context, in TransferInput) (*model.Transaction, bool, error) {
	if in.FromUserID == in.ToUserID {
		return nil, false, ErrSelfTransfer
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx) // no-op kalau sudah commit

	// 1. Kunci kedua wallet, diurutkan berdasarkan user_id (mencegah deadlock).
	rows, err := tx.Query(ctx,
		`SELECT user_id, balance FROM wallets
		 WHERE user_id IN ($1, $2)
		 ORDER BY user_id
		 FOR UPDATE`,
		in.FromUserID, in.ToUserID,
	)
	if err != nil {
		return nil, false, err
	}

	balances := make(map[uuid.UUID]int64, 2)
	for rows.Next() {
		var id uuid.UUID
		var balance int64
		if err := rows.Scan(&id, &balance); err != nil {
			rows.Close()
			return nil, false, err
		}
		balances[id] = balance
	}
	rows.Close() // wajib sebelum query berikutnya di transaksi yang sama
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	senderBalance, ok := balances[in.FromUserID]
	if !ok {
		return nil, false, ErrWalletNotFound
	}
	if _, ok := balances[in.ToUserID]; !ok {
		return nil, false, ErrRecipientNotFound
	}

	// 2. Cek idempotency (aman dari race karena wallet pengirim sudah terkunci).
	var existing model.Transaction
	err = tx.QueryRow(ctx,
		`SELECT id, from_user_id, to_user_id, amount, notes, idempotency_key, created_at
		 FROM transactions
		 WHERE from_user_id = $1 AND idempotency_key = $2`,
		in.FromUserID, in.IdempotencyKey,
	).Scan(&existing.ID, &existing.FromUserID, &existing.ToUserID,
		&existing.Amount, &existing.Notes, &existing.IdempotencyKey, &existing.CreatedAt)

	switch {
	case err == nil:
		if existing.ToUserID != in.ToUserID || existing.Amount != in.Amount || existing.Notes != in.Notes {
			return nil, false, ErrIdempotencyConflict
		}
		return &existing, true, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, false, err
	}

	// 3. Cek saldo.
	if senderBalance < in.Amount {
		return nil, false, ErrInsufficientFunds
	}

	// 4. Pindahkan uang.
	if _, err := tx.Exec(ctx,
		`UPDATE wallets SET balance = balance - $2, updated_at = now() WHERE user_id = $1`,
		in.FromUserID, in.Amount,
	); err != nil {
		return nil, false, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE wallets SET balance = balance + $2, updated_at = now() WHERE user_id = $1`,
		in.ToUserID, in.Amount,
	); err != nil {
		return nil, false, err
	}

	// 5. Catat transaksi.
	rec := model.Transaction{
		ID:             uuid.New(),
		FromUserID:     in.FromUserID,
		ToUserID:       in.ToUserID,
		Amount:         in.Amount,
		Notes:          in.Notes,
		IdempotencyKey: in.IdempotencyKey,
	}
	if err := tx.QueryRow(ctx,
		`INSERT INTO transactions (id, from_user_id, to_user_id, amount, notes, idempotency_key)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at`,
		rec.ID, rec.FromUserID, rec.ToUserID, rec.Amount, rec.Notes, rec.IdempotencyKey,
	).Scan(&rec.CreatedAt); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return &rec, false, nil
}

// ListForUser mengembalikan transaksi masuk dan keluar milik user, terbaru dulu.
func (r *TransactionRepository) ListForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.HistoryItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.from_user_id, t.amount, t.notes, t.created_at, fu.email, tu.email
		 FROM transactions t
		 JOIN users fu ON fu.id = t.from_user_id
		 JOIN users tu ON tu.id = t.to_user_id
		 WHERE t.from_user_id = $1 OR t.to_user_id = $1
		 ORDER BY t.created_at DESC, t.id DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.HistoryItem, 0, limit) // slice kosong, bukan nil, agar JSON jadi []
	for rows.Next() {
		var (
			item                 model.HistoryItem
			fromID               uuid.UUID
			fromEmail, toEmail   string
		)
		if err := rows.Scan(&item.ID, &fromID, &item.Amount, &item.Notes, &item.CreatedAt, &fromEmail, &toEmail); err != nil {
			return nil, err
		}
		if fromID == userID {
			item.Direction = "out"
			item.CounterpartyEmail = toEmail
		} else {
			item.Direction = "in"
			item.CounterpartyEmail = fromEmail
		}
		items = append(items, item)
	}
	return items, rows.Err()
}