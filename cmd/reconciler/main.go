package main

import (
	"context"
	"log"
	"os"
	"time"

	"adnet/internal/store/postgres"
	"adnet/internal/store/redis"
)

type Config struct {
	RedisURL    string
	PostgresURL string
	Interval    time.Duration
}

func main() {
	config := Config{
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/adnet?sslmode=disable"),
		Interval:    1 * time.Minute,
	}

	ctx := context.Background()

	redisClient, err := redis.NewClient(config.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	db, err := postgres.NewDB(ctx, config.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	reconciler := &Reconciler{
		redis: redisClient,
		db:    db,
	}

	log.Printf("Reconciler starting with interval: %v", config.Interval)

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Run initial reconciliation
	if err := reconciler.ReconcileAll(ctx); err != nil {
		log.Printf("Initial reconciliation error: %v", err)
	}

	// Also run daily budget reset check every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := reconciler.CheckDailyBudgetResets(ctx); err != nil {
					log.Printf("Daily budget reset error: %v", err)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := reconciler.ReconcileAll(ctx); err != nil {
				log.Printf("Reconciliation error: %v", err)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type Reconciler struct {
	redis *redis.Client
	db    *postgres.DB
}

func (r *Reconciler) ReconcileAll(ctx context.Context) error {
	// Get all active campaigns from Postgres
	campaigns, err := r.getActiveCampaigns(ctx)
	if err != nil {
		return err
	}

	for _, campaign := range campaigns {
		if err := r.reconcileCampaign(ctx, campaign); err != nil {
			log.Printf("Error reconciling campaign %s: %v", campaign.ID, err)
		}
	}

	return nil
}

func (r *Reconciler) getActiveCampaigns(ctx context.Context) ([]*postgres.Campaign, error) {
	const query = `
		SELECT id, advertiser_id, name, status, pricing_model, bid_amount_cents, daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at, created_at
		FROM campaigns
		WHERE status IN ('active', 'paused')
	`

	rows, err := r.db.Pool().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []*postgres.Campaign
	for rows.Next() {
		var campaign postgres.Campaign
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

func (r *Reconciler) reconcileCampaign(ctx context.Context, campaign *postgres.Campaign) error {
	campaignID := campaign.ID.String()

	// Sync daily budget to Redis
	if err := r.redis.SetDailyBudget(ctx, campaignID, int64(campaign.DailyBudgetCents)); err != nil {
		return err
	}

	// Calculate total remaining budget from Postgres
	totalSpent, err := r.getTotalSpend(ctx, campaignID)
	if err != nil {
		return err
	}

	totalRemaining := int64(campaign.TotalBudgetCents) - totalSpent
	if totalRemaining < 0 {
		totalRemaining = 0
	}

	if err := r.redis.SetTotalBudgetRemaining(ctx, campaignID, totalRemaining); err != nil {
		return err
	}

	// Check if campaign should be paused due to budget exhaustion
	if totalRemaining == 0 && campaign.Status == "active" {
		log.Printf("Pausing campaign %s due to budget exhaustion", campaignID)
		if err := r.db.UpdateCampaignStatus(ctx, campaign.ID, "exhausted"); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) getTotalSpend(ctx context.Context, campaignID string) (int64, error) {
	const query = `
		SELECT COALESCE(SUM(cost_cents), 0)
		FROM impressions
		WHERE campaign_id = $1
	`

	var totalSpend int64
	if err := r.db.Pool().QueryRow(ctx, query, campaignID).Scan(&totalSpend); err != nil {
		return 0, err
	}

	return totalSpend, nil
}

func (r *Reconciler) CheckDailyBudgetResets(ctx context.Context) error {
	// Get all active campaigns
	campaigns, err := r.getActiveCampaigns(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, campaign := range campaigns {
		// Load campaign timezone
		loc, err := time.LoadLocation(campaign.Timezone)
		if err != nil {
			log.Printf("Error loading timezone for campaign %s: %v", campaign.ID, err)
			continue
		}

		campaignTime := now.In(loc)
		midnight := time.Date(campaignTime.Year(), campaignTime.Month(), campaignTime.Day(), 0, 0, 0, 0, loc)

		// Check if we just passed midnight (within the last minute)
		if campaignTime.Sub(midnight) < 1*time.Minute {
			log.Printf("Resetting daily budget for campaign %s", campaign.ID)
			if err := r.redis.SetDailyBudget(ctx, campaign.ID.String(), int64(campaign.DailyBudgetCents)); err != nil {
				log.Printf("Error resetting daily budget for campaign %s: %v", campaign.ID, err)
			}
			// Reset daily spend counter
			if err := r.redis.SetDailyBudget(ctx, "spend:"+campaign.ID.String(), 0); err != nil {
				log.Printf("Error resetting daily spend for campaign %s: %v", campaign.ID, err)
			}
		}
	}

	return nil
}
