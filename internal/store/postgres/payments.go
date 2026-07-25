package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TransferRecipient represents a publisher's payout details stored from Paystack
type TransferRecipient struct {
	ID            uuid.UUID `json:"id"`
	AccountID     uuid.UUID `json:"account_id"`
	RecipientCode string    `json:"recipient_code"`
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	Phone         *string   `json:"phone,omitempty"`
	Email         *string   `json:"email,omitempty"`
	BankCode      *string   `json:"bank_code,omitempty"`
	Currency      string    `json:"currency"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
}

// PaystackEvent represents a webhook event from Paystack
type PaystackEvent struct {
	ID          uuid.UUID `json:"id"`
	EventType   string    `json:"event_type"`
	PaystackID  *string   `json:"paystack_id,omitempty"`
	Reference   *string   `json:"reference,omitempty"`
	AmountCents *int64    `json:"amount_cents,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	Status      *string   `json:"status,omitempty"`
	Processed   bool      `json:"processed"`
	CreatedAt   time.Time `json:"created_at"`
}

// --- Transfer Recipients ---

func (db *DB) CreateTransferRecipient(ctx context.Context, accountID uuid.UUID, recipientCode, recipientType, name string, phone, email, bankCode *string, currency string) (*TransferRecipient, error) {
	// Set all other recipients for this account to non-default
	_, _ = db.pool.Exec(ctx, `UPDATE transfer_recipients SET is_default = false WHERE account_id = $1`, accountID)

	const query = `
		INSERT INTO transfer_recipients (account_id, recipient_code, type, name, phone, email, bank_code, currency, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
		RETURNING id, account_id, recipient_code, type, name, phone, email, bank_code, currency, is_default, created_at
	`

	var r TransferRecipient
	err := db.pool.QueryRow(ctx, query, accountID, recipientCode, recipientType, name, phone, email, bankCode, currency).Scan(
		&r.ID, &r.AccountID, &r.RecipientCode, &r.Type, &r.Name, &r.Phone, &r.Email, &r.BankCode, &r.Currency, &r.IsDefault, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) GetDefaultTransferRecipient(ctx context.Context, accountID uuid.UUID) (*TransferRecipient, error) {
	const query = `
		SELECT id, account_id, recipient_code, type, name, phone, email, bank_code, currency, is_default, created_at
		FROM transfer_recipients
		WHERE account_id = $1 AND is_default = true
		LIMIT 1
	`

	var r TransferRecipient
	err := db.pool.QueryRow(ctx, query, accountID).Scan(
		&r.ID, &r.AccountID, &r.RecipientCode, &r.Type, &r.Name, &r.Phone, &r.Email, &r.BankCode, &r.Currency, &r.IsDefault, &r.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (db *DB) ListTransferRecipients(ctx context.Context, accountID uuid.UUID) ([]*TransferRecipient, error) {
	const query = `
		SELECT id, account_id, recipient_code, type, name, phone, email, bank_code, currency, is_default, created_at
		FROM transfer_recipients
		WHERE account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []*TransferRecipient
	for rows.Next() {
		var r TransferRecipient
		if err := rows.Scan(&r.ID, &r.AccountID, &r.RecipientCode, &r.Type, &r.Name, &r.Phone, &r.Email, &r.BankCode, &r.Currency, &r.IsDefault, &r.CreatedAt); err != nil {
			return nil, err
		}
		recipients = append(recipients, &r)
	}
	return recipients, nil
}

// --- Paystack Events ---

func (db *DB) CreatePaystackEvent(ctx context.Context, eventType string, paystackID, reference *string, amountCents *int64, currency, status *string, rawPayload []byte) (*PaystackEvent, error) {
	const query = `
		INSERT INTO paystack_events (event_type, paystack_id, reference, amount_cents, currency, status, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (paystack_id) WHERE paystack_id IS NOT NULL DO NOTHING
		RETURNING id, event_type, paystack_id, reference, amount_cents, currency, status, processed, created_at
	`

	var e PaystackEvent
	err := db.pool.QueryRow(ctx, query, eventType, paystackID, reference, amountCents, currency, status, rawPayload).Scan(
		&e.ID, &e.EventType, &e.PaystackID, &e.Reference, &e.AmountCents, &e.Currency, &e.Status, &e.Processed, &e.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // duplicate event
		}
		return nil, err
	}
	return &e, nil
}

func (db *DB) MarkPaystackEventProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := db.pool.Exec(ctx, `UPDATE paystack_events SET processed = true WHERE id = $1`, id)
	return err
}

// --- Payout Request Updates ---

func (db *DB) UpdatePayoutWithPaystack(ctx context.Context, id uuid.UUID, status, paystackTransferCode, paystackReference, recipientCode string) error {
	const query = `
		UPDATE payout_requests
		SET status = $2, paystack_transfer_code = $3, paystack_reference = $4, recipient_code = $5
		WHERE id = $1
	`
	_, err := db.pool.Exec(ctx, query, id, status, paystackTransferCode, paystackReference, recipientCode)
	return err
}

func (db *DB) GetPayoutByPaystackReference(ctx context.Context, reference string) (*PayoutRequest, error) {
	const query = `
		SELECT id, publisher_id, amount_cents, status, mpesa_receipt, paystack_transfer_code, paystack_reference, recipient_code, failure_reason, requested_at, processed_at
		FROM payout_requests
		WHERE paystack_reference = $1
	`

	var req PayoutRequest
	err := db.pool.QueryRow(ctx, query, reference).Scan(
		&req.ID, &req.PublisherID, &req.AmountCents, &req.Status, &req.MpesaReceipt, &req.PaystackTransferCode, &req.PaystackReference, &req.RecipientCode, &req.FailureReason, &req.RequestedAt, &req.ProcessedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (db *DB) GetTransactionByReference(ctx context.Context, reference string) (*Transaction, error) {
	const query = `
		SELECT id, account_id, type, amount_cents, reference, status, paystack_reference, metadata, created_at
		FROM transactions
		WHERE reference = $1
		LIMIT 1
	`

	var tx Transaction
	err := db.pool.QueryRow(ctx, query, reference).Scan(
		&tx.ID, &tx.AccountID, &tx.Type, &tx.AmountCents, &tx.Reference, &tx.Status, &tx.PaystackReference, &tx.Metadata, &tx.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

// --- Admin Financial Queries ---

type FinancialSummary struct {
	TotalDeposits         int64 `json:"total_deposits"`
	TotalPublisherPayouts int64 `json:"total_publisher_payouts"`
	PendingPayouts        int64 `json:"pending_payouts"`
	TotalPlatformFees     int64 `json:"total_platform_fees"`
	TotalAdSpend          int64 `json:"total_ad_spend"`
}

func (db *DB) GetFinancialSummary(ctx context.Context) (*FinancialSummary, error) {
	var s FinancialSummary

	db.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_cents), 0) FROM transactions WHERE type = 'topup'`).Scan(&s.TotalDeposits)
	db.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_cents), 0) FROM payout_requests WHERE status = 'paid'`).Scan(&s.TotalPublisherPayouts)
	db.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_cents), 0) FROM payout_requests WHERE status IN ('pending', 'processing')`).Scan(&s.PendingPayouts)
	db.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_cents), 0) FROM transactions WHERE type = 'spend'`).Scan(&s.TotalAdSpend)

	// Platform fees = total ad spend - total publisher payouts (or from revenue_ledger if available)
	db.pool.QueryRow(ctx, `SELECT COALESCE(SUM(platform_fee_cents), 0) FROM revenue_ledger`).Scan(&s.TotalPlatformFees)

	return &s, nil
}

func (db *DB) ListAllDeposits(ctx context.Context, limit int) ([]*Transaction, error) {
	const query = `
		SELECT t.id, t.account_id, t.type, t.amount_cents, t.reference, t.status, t.paystack_reference, t.metadata, t.created_at
		FROM transactions t
		WHERE t.type = 'topup'
		ORDER BY t.created_at DESC
		LIMIT $1
	`

	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.AccountID, &tx.Type, &tx.AmountCents, &tx.Reference, &tx.Status, &tx.PaystackReference, &tx.Metadata, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, &tx)
	}
	return txs, nil
}

func (db *DB) ListAllPayouts(ctx context.Context, limit int) ([]*PayoutRequest, error) {
	const query = `
		SELECT id, publisher_id, amount_cents, status, mpesa_receipt, paystack_transfer_code, paystack_reference, recipient_code, failure_reason, requested_at, processed_at
		FROM payout_requests
		ORDER BY requested_at DESC
		LIMIT $1
	`

	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []*PayoutRequest
	for rows.Next() {
		var p PayoutRequest
		if err := rows.Scan(&p.ID, &p.PublisherID, &p.AmountCents, &p.Status, &p.MpesaReceipt, &p.PaystackTransferCode, &p.PaystackReference, &p.RecipientCode, &p.FailureReason, &p.RequestedAt, &p.ProcessedAt); err != nil {
			return nil, err
		}
		payouts = append(payouts, &p)
	}
	return payouts, nil
}
