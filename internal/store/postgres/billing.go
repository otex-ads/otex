package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Wallet struct {
	AccountID   uuid.UUID `json:"account_id"`
	BalanceCents int64    `json:"balance_cents"`
	Currency    string    `json:"currency"`
}

type Transaction struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"account_id"`
	Type        string    `json:"type"`
	AmountCents int64     `json:"amount_cents"`
	Reference   *string   `json:"reference,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type PayoutRequest struct {
	ID           uuid.UUID  `json:"id"`
	PublisherID  uuid.UUID  `json:"publisher_id"`
	AmountCents  int64      `json:"amount_cents"`
	Status       string     `json:"status"`
	MpesaReceipt *string    `json:"mpesa_receipt,omitempty"`
	RequestedAt  time.Time  `json:"requested_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
}

func (db *DB) GetWallet(ctx context.Context, accountID uuid.UUID) (*Wallet, error) {
	const query = `
		SELECT account_id, balance_cents, currency
		FROM wallets
		WHERE account_id = $1
	`

	var wallet Wallet
	err := db.pool.QueryRow(ctx, query, accountID).Scan(
		&wallet.AccountID,
		&wallet.BalanceCents,
		&wallet.Currency,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &wallet, nil
}

func (db *DB) CreateWallet(ctx context.Context, accountID uuid.UUID, balanceCents int64, currency string) (*Wallet, error) {
	const query = `
		INSERT INTO wallets (account_id, balance_cents, currency)
		VALUES ($1, $2, $3)
		RETURNING account_id, balance_cents, currency
	`

	var wallet Wallet
	err := db.pool.QueryRow(ctx, query, accountID, balanceCents, currency).Scan(
		&wallet.AccountID,
		&wallet.BalanceCents,
		&wallet.Currency,
	)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (db *DB) UpdateWalletBalance(ctx context.Context, accountID uuid.UUID, deltaCents int64) (*Wallet, error) {
	const query = `
		UPDATE wallets
		SET balance_cents = balance_cents + $2
		WHERE account_id = $1
		RETURNING account_id, balance_cents, currency
	`

	var wallet Wallet
	err := db.pool.QueryRow(ctx, query, accountID, deltaCents).Scan(
		&wallet.AccountID,
		&wallet.BalanceCents,
		&wallet.Currency,
	)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (db *DB) CreateTransaction(ctx context.Context, accountID uuid.UUID, txType string, amountCents int64, reference *string) (*Transaction, error) {
	const query = `
		INSERT INTO transactions (account_id, type, amount_cents, reference)
		VALUES ($1, $2, $3, $4)
		RETURNING id, account_id, type, amount_cents, reference, created_at
	`

	var transaction Transaction
	err := db.pool.QueryRow(ctx, query, accountID, txType, amountCents, reference).Scan(
		&transaction.ID,
		&transaction.AccountID,
		&transaction.Type,
		&transaction.AmountCents,
		&transaction.Reference,
		&transaction.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (db *DB) ListTransactions(ctx context.Context, accountID uuid.UUID, limit int) ([]*Transaction, error) {
	const query = `
		SELECT id, account_id, type, amount_cents, reference, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := db.pool.Query(ctx, query, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.AccountID,
			&tx.Type,
			&tx.AmountCents,
			&tx.Reference,
			&tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (db *DB) CreatePayoutRequest(ctx context.Context, publisherID uuid.UUID, amountCents int64) (*PayoutRequest, error) {
	const query = `
		INSERT INTO payout_requests (publisher_id, amount_cents, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, publisher_id, amount_cents, status, mpesa_receipt, requested_at, processed_at
	`

	var request PayoutRequest
	err := db.pool.QueryRow(ctx, query, publisherID, amountCents).Scan(
		&request.ID,
		&request.PublisherID,
		&request.AmountCents,
		&request.Status,
		&request.MpesaReceipt,
		&request.RequestedAt,
		&request.ProcessedAt,
	)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (db *DB) UpdatePayoutRequestStatus(ctx context.Context, id uuid.UUID, status string, mpesaReceipt *string) (*PayoutRequest, error) {
	const query = `
		UPDATE payout_requests
		SET status = $2, mpesa_receipt = $3, processed_at = NOW()
		WHERE id = $1
		RETURNING id, publisher_id, amount_cents, status, mpesa_receipt, requested_at, processed_at
	`

	var request PayoutRequest
	err := db.pool.QueryRow(ctx, query, id, status, mpesaReceipt).Scan(
		&request.ID,
		&request.PublisherID,
		&request.AmountCents,
		&request.Status,
		&request.MpesaReceipt,
		&request.RequestedAt,
		&request.ProcessedAt,
	)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (db *DB) ListPayoutRequests(ctx context.Context, publisherID uuid.UUID) ([]*PayoutRequest, error) {
	const query = `
		SELECT id, publisher_id, amount_cents, status, mpesa_receipt, requested_at, processed_at
		FROM payout_requests
		WHERE publisher_id = $1
		ORDER BY requested_at DESC
	`

	rows, err := db.pool.Query(ctx, query, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*PayoutRequest
	for rows.Next() {
		var req PayoutRequest
		if err := rows.Scan(
			&req.ID,
			&req.PublisherID,
			&req.AmountCents,
			&req.Status,
			&req.MpesaReceipt,
			&req.RequestedAt,
			&req.ProcessedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}

	return requests, nil
}
