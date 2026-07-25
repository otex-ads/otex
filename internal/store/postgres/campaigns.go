package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Campaign struct {
	ID               uuid.UUID  `json:"id"`
	AdvertiserID     uuid.UUID  `json:"advertiser_id"`
	Name             string     `json:"name"`
	Status           string     `json:"status"`
	PricingModel     string     `json:"pricing_model"`
	BidAmountCents   int        `json:"bid_amount_cents"`
	DailyBudgetCents int        `json:"daily_budget_cents"`
	TotalBudgetCents int        `json:"total_budget_cents"`
	Timezone         string     `json:"timezone"`
	StartsAt         *time.Time `json:"starts_at,omitempty"`
	EndsAt           *time.Time `json:"ends_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (db *DB) CreateCampaign(ctx context.Context, advertiserID uuid.UUID, name, status, pricingModel string, bidAmountCents, dailyBudgetCents, totalBudgetCents int, timezone string, startsAt, endsAt *time.Time) (*Campaign, error) {
	const query = `
		INSERT INTO campaigns (advertiser_id, name, status, pricing_model, bid_amount_cents, daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, advertiser_id, name, status, pricing_model, bid_amount_cents, daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at, created_at
	`

	var campaign Campaign
	err := db.pool.QueryRow(ctx, query, advertiserID, name, status, pricingModel, bidAmountCents, dailyBudgetCents, totalBudgetCents, timezone, startsAt, endsAt).Scan(
		&campaign.ID,
		&campaign.AdvertiserID,
		&campaign.Name,
		&campaign.Status,
		&campaign.PricingModel,
		&campaign.BidAmountCents,
		&campaign.DailyBudgetCents,
		&campaign.TotalBudgetCents,
		&campaign.Timezone,
		&campaign.StartsAt,
		&campaign.EndsAt,
		&campaign.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (db *DB) GetCampaignByID(ctx context.Context, id uuid.UUID) (*Campaign, error) {
	const query = `
		SELECT id, advertiser_id, name, status, pricing_model, bid_amount_cents, daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at, created_at
		FROM campaigns
		WHERE id = $1
	`

	var campaign Campaign
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&campaign.ID,
		&campaign.AdvertiserID,
		&campaign.Name,
		&campaign.Status,
		&campaign.PricingModel,
		&campaign.BidAmountCents,
		&campaign.DailyBudgetCents,
		&campaign.TotalBudgetCents,
		&campaign.Timezone,
		&campaign.StartsAt,
		&campaign.EndsAt,
		&campaign.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &campaign, nil
}

func (db *DB) ListCampaignsByAdvertiser(ctx context.Context, advertiserID uuid.UUID) ([]*Campaign, error) {
	const query = `
		SELECT id, advertiser_id, name, status, pricing_model, bid_amount_cents, daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at, created_at
		FROM campaigns
		WHERE advertiser_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.pool.Query(ctx, query, advertiserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []*Campaign
	for rows.Next() {
		var campaign Campaign
		if err := rows.Scan(
			&campaign.ID,
			&campaign.AdvertiserID,
			&campaign.Name,
			&campaign.Status,
			&campaign.PricingModel,
			&campaign.BidAmountCents,
			&campaign.DailyBudgetCents,
			&campaign.TotalBudgetCents,
			&campaign.Timezone,
			&campaign.StartsAt,
			&campaign.EndsAt,
			&campaign.CreatedAt,
		); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, &campaign)
	}

	return campaigns, nil
}

func (db *DB) UpdateCampaign(ctx context.Context, id uuid.UUID, name, status, pricingModel string, bidAmountCents, dailyBudgetCents, totalBudgetCents int, timezone string, startsAt, endsAt *time.Time) (*Campaign, error) {
	const query = `
		UPDATE campaigns
		SET name = $2, status = $3, pricing_model = $4, bid_amount_cents = $5, daily_budget_cents = $6, total_budget_cents = $7, timezone = $8, starts_at = $9, ends_at = $10
		WHERE id = $1
		RETURNING id, advertiser_id, name, status, pricing_model, bid_amount_cents, daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at, created_at
	`

	var campaign Campaign
	err := db.pool.QueryRow(ctx, query, id, name, status, pricingModel, bidAmountCents, dailyBudgetCents, totalBudgetCents, timezone, startsAt, endsAt).Scan(
		&campaign.ID,
		&campaign.AdvertiserID,
		&campaign.Name,
		&campaign.Status,
		&campaign.PricingModel,
		&campaign.BidAmountCents,
		&campaign.DailyBudgetCents,
		&campaign.TotalBudgetCents,
		&campaign.Timezone,
		&campaign.StartsAt,
		&campaign.EndsAt,
		&campaign.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (db *DB) UpdateCampaignStatus(ctx context.Context, id uuid.UUID, status string) error {
	const query = `UPDATE campaigns SET status = $2 WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, id, status)
	return err
}

func (db *DB) CreateTargetingRule(ctx context.Context, campaignID uuid.UUID, countries, devices, os []string) (*struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
}, error) {
	const query = `
		INSERT INTO targeting_rules (campaign_id, countries, device_types, os)
		VALUES ($1, $2, $3, $4)
		RETURNING id, campaign_id
	`

	var result struct {
		ID         uuid.UUID
		CampaignID uuid.UUID
	}
	err := db.pool.QueryRow(ctx, query, campaignID, countries, devices, os).Scan(&result.ID, &result.CampaignID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
