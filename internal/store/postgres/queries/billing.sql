-- name: GetWallet :one
SELECT * FROM wallets
WHERE account_id = $1;

-- name: CreateWallet :one
INSERT INTO wallets (account_id, balance_cents, currency)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateWalletBalance :one
UPDATE wallets
SET balance_cents = balance_cents + $2
WHERE account_id = $1
RETURNING *;

-- name: CreateTransaction :one
INSERT INTO transactions (account_id, type, amount_cents, reference)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListTransactionsByAccount :many
SELECT * FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: CreatePayoutRequest :one
INSERT INTO payout_requests (publisher_id, amount_cents, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdatePayoutRequestStatus :one
UPDATE payout_requests
SET status = $2,
    mpesa_receipt = $3,
    processed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListPayoutRequestsByPublisher :many
SELECT * FROM payout_requests
WHERE publisher_id = $1
ORDER BY requested_at DESC;
