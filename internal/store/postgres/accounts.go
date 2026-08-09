package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Account struct {
	ID           uuid.UUID `json:"id"`
	Type         string    `json:"type"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CompanyName  *string   `json:"company_name,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type APIKey struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	KeyHash   string    `json:"-"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

func (db *DB) CreateAccount(ctx context.Context, accountType, email, passwordHash string, companyName *string) (*Account, error) {
	const query = `
		INSERT INTO accounts (type, email, password_hash, company_name, status)
		VALUES ($1, $2, $3, $4, 'active')
		RETURNING id, type, email, password_hash, company_name, status, created_at
	`

	var account Account
	err := db.pool.QueryRow(ctx, query, accountType, email, passwordHash, companyName).Scan(
		&account.ID,
		&account.Type,
		&account.Email,
		&account.PasswordHash,
		&account.CompanyName,
		&account.Status,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (db *DB) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	const query = `
		SELECT id, type, email, password_hash, company_name, status, created_at
		FROM accounts
		WHERE email = $1
	`

	var account Account
	err := db.pool.QueryRow(ctx, query, email).Scan(
		&account.ID,
		&account.Type,
		&account.Email,
		&account.PasswordHash,
		&account.CompanyName,
		&account.Status,
		&account.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}

func (db *DB) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	const query = `
		SELECT id, type, email, password_hash, company_name, status, created_at
		FROM accounts
		WHERE id = $1
	`

	var account Account
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.Type,
		&account.Email,
		&account.PasswordHash,
		&account.CompanyName,
		&account.Status,
		&account.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}

func (db *DB) UpdateAccountStatus(ctx context.Context, id uuid.UUID, status string) error {
	const query = `
		UPDATE accounts
		SET status = $2
		WHERE id = $1
	`

	_, err := db.pool.Exec(ctx, query, id, status)
	return err
}

func (db *DB) ListAccountsByType(ctx context.Context, accountType string) ([]*Account, error) {
	const query = `
		SELECT id, type, email, password_hash, company_name, status, created_at
		FROM accounts
		WHERE type = $1
		ORDER BY created_at DESC
	`

	rows, err := db.pool.Query(ctx, query, accountType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		var account Account
		err := rows.Scan(
			&account.ID,
			&account.Type,
			&account.Email,
			&account.PasswordHash,
			&account.CompanyName,
			&account.Status,
			&account.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	return accounts, nil
}
