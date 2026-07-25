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
	Format     string    `json:"format"`
	Title      *string   `json:"title,omitempty"`
	Body       *string   `json:"body,omitempty"`
	IconURL    *string   `json:"icon_url,omitempty"`
	ImageURL   *string   `json:"image_url,omitempty"`
	ClickURL   string    `json:"click_url"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func (db *DB) CreateCreative(ctx context.Context, campaignID uuid.UUID, format, title, body, iconURL, imageURL, clickURL, status string) (*Creative, error) {
	const query = `
		INSERT INTO creatives (campaign_id, format, title, body, icon_url, image_url, click_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, campaign_id, format, title, body, icon_url, image_url, click_url, status, created_at
	`

	var creative Creative
	err := db.pool.QueryRow(ctx, query, campaignID, format, &title, &body, &iconURL, &imageURL, clickURL, status).Scan(
		&creative.ID,
		&creative.CampaignID,
		&creative.Format,
		&creative.Title,
		&creative.Body,
		&creative.IconURL,
		&creative.ImageURL,
		&creative.ClickURL,
		&creative.Status,
		&creative.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &creative, nil
}

func (db *DB) GetCreativeByID(ctx context.Context, id uuid.UUID) (*Creative, error) {
	const query = `
		SELECT id, campaign_id, format, title, body, icon_url, image_url, click_url, status, created_at
		FROM creatives
		WHERE id = $1
	`

	var creative Creative
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&creative.ID,
		&creative.CampaignID,
		&creative.Format,
		&creative.Title,
		&creative.Body,
		&creative.IconURL,
		&creative.ImageURL,
		&creative.ClickURL,
		&creative.Status,
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
		SELECT id, campaign_id, format, title, body, icon_url, image_url, click_url, status, created_at
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
			&creative.Format,
			&creative.Title,
			&creative.Body,
			&creative.IconURL,
			&creative.ImageURL,
			&creative.ClickURL,
			&creative.Status,
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
		SELECT id, campaign_id, format, title, body, icon_url, image_url, click_url, status, created_at
		FROM creatives
		WHERE status = 'pending_review'
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
			&creative.Format,
			&creative.Title,
			&creative.Body,
			&creative.IconURL,
			&creative.ImageURL,
			&creative.ClickURL,
			&creative.Status,
			&creative.CreatedAt,
		); err != nil {
			return nil, err
		}
		creatives = append(creatives, &creative)
	}

	return creatives, nil
}
