-- name: CreateAccount :one
INSERT INTO accounts (type, email, password_hash, company_name, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = $1;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = $1;

-- name: UpdateAccountStatus :one
UPDATE accounts
SET status = $2
WHERE id = $1
RETURNING *;
