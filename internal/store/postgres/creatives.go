package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Creative struct {
	ID         uuid.UUID `json:"id"`
	CampaignID uuid.UUID `json:"campaign_id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	URL        string    `json:"url"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	CreatedAt  time.Time `json:"created_at"`
}

func (db *DB) CreateCreative(ctx context.Context, campaignID uuid.UUID, creativeType, status, url string, width, height int) (*Creative, error) {
	const query = `
		INSERT INTO creatives (campaign_id, type, status, url, width, height)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, campaign_id, type, status, url, width, height, created_at
	`

	var creative Creative
	err := db.pool.QueryRow(ctx, query, campaignID, creativeType, status, url, width, height).Scan(
		&creative.ID,
		&creative.CampaignID,
		&creative.Type,
		&creative.Status,
		&creative.URL,
		&creative.Width,
		&creative.Height,
		&creative.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &creative, nil
}

func (db *DB) GetCreativeByID(ctx context.Context, id uuid.UUID) (*Creative, error) {
	const query = `
		SELECT id, campaign_id, type, status, url, width, height, created_at
		FROM creatives
		WHERE id = $1
	`

	var creative Creative
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&creative.ID,
		&creative.CampaignID,
		&creative.Type,
		&creative.Status,
		&creative.URL,
		&creative.Width,
		&creative.Height,
		&creative.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &creative, nil
}

func (db *DB) ListCreativesByCampaign(ctx context.Context, campaignID uuid.UUID) ([]*Creative, error) {
	const query = `
		SELECT id, campaign_id, type, status, url, width, height, created_at
		FROM creatives
		WHERE campaign_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.pool.Query(ctx, query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creatives []*Creative
	for rows.Next() {
		var creative Creative
		if err := rows.Scan(
			&creative.ID,
			&creative.CampaignID,
			&creative.Type,
			&creative.Status,
			&creative.URL,
			&creative.Width,
			&creative.Height,
			&creative.CreatedAt,
		); err != nil {
			return nil, err
		}
		creatives = append(creatives, &creative)
	}

	return creatives, nil
}

func (db *DB) UpdateCreativeStatus(ctx context.Context, id uuid.UUID, status string) error {
	const query = `
		UPDATE creatives
		SET status = $2
		WHERE id = $1
	`
	_, err := db.pool.Exec(ctx, query, id, status)
	return err
}

func (db *DB) ListPendingCreatives(ctx context.Context) ([]*Creative, error) {
	const query = `
		SELECT id, campaign_id, type, status, url, width, height, created_at
		FROM creatives
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creatives []*Creative
	for rows.Next() {
		var creative Creative
		if err := rows.Scan(
			&creative.ID,
			&creative.CampaignID,
			&creative.Type,
			&creative.Status,
			&creative.URL,
			&creative.Width,
			&creative.Height,
			&creative.CreatedAt,
		); err != nil {
			return nil, err
		}
		creatives = append(creatives, &creative)
	}

	return creatives, nil
}
