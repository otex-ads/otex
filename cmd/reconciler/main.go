package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"adnet/internal/email"
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

	// Run low balance check every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := reconciler.CheckLowBalance(ctx); err != nil {
					log.Printf("Low balance check error: %v", err)
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

	// Sync serving metadata (zones, campaign creatives, and candidates) so
	// that zones/campaigns created via the portals become servable.
	if err := r.reconcileServingMetadata(ctx, campaigns); err != nil {
		log.Printf("Error reconciling serving metadata: %v", err)
	}

	return nil
}

// zoneRow holds the zone fields needed for ad serving.
type zoneRow struct {
	ID               string
	SiteID           string
	Name             string
	Format           string
	FloorPriceCents  int
	Status           string
	Countries        string
	DeviceTypes      string
	OS               string
	Browsers         string
	Carriers         string
	ConnectionTypes  string
}

// campaignCreative holds a campaign joined with its best approved creative.
type campaignCreative struct {
	CampaignID string
	BidCents   int
	Format     string
	CreativeID string
	Title      string
	Body       string
	IconURL    string
	ImageURL   string
	ClickURL   string
	Status     string
}

// reconcileServingMetadata pushes zone metadata, campaign creative metadata,
// and per-zone candidate lists into Redis so the adserve service can match
// ads to zones by format.
func (r *Reconciler) reconcileServingMetadata(ctx context.Context, campaigns []*postgres.Campaign) error {
	zones, err := r.getActiveZones(ctx)
	if err != nil {
		return err
	}

	// Write zone metadata.
	for _, z := range zones {
		meta := map[string]string{
			"site_id":           z.SiteID,
			"name":              z.Name,
			"format":            z.Format,
			"floor_price_cents": strconv.Itoa(z.FloorPriceCents),
			"status":            z.Status,
			"countries":         z.Countries,
			"device_types":      z.DeviceTypes,
			"os":                z.OS,
			"browsers":          z.Browsers,
			"carriers":          z.Carriers,
			"connection_types":  z.ConnectionTypes,
		}
		if err := r.redis.SetZoneMeta(ctx, z.ID, meta); err != nil {
			log.Printf("Error setting zone meta %s: %v", z.ID, err)
		}
	}

	// Build campaign creative metadata for active campaigns that have an
	// approved creative, grouped by format for candidate matching.
	byFormat := make(map[string]map[string]float64)
	for _, c := range campaigns {
		if c.Status != "active" {
			continue
		}
		cc, err := r.getApprovedCreative(ctx, c.ID.String())
		if err != nil || cc == nil {
			continue
		}
		cc.BidCents = c.BidAmountCents
		cc.Status = c.Status

		meta := map[string]string{
			"bid_cents":       strconv.Itoa(cc.BidCents),
			"bid":             strconv.FormatFloat(float64(cc.BidCents)/100.0, 'f', 2, 64),
			"format":          cc.Format,
			"creative_id":     cc.CreativeID,
			"title":           cc.Title,
			"body":            cc.Body,
			"icon_url":        cc.IconURL,
			"image_url":       cc.ImageURL,
			"click_url":       cc.ClickURL,
			"status":          "active",
			"creative_status": "approved",
		}
		if err := r.redis.SetCampaignMeta(ctx, cc.CampaignID, meta); err != nil {
			log.Printf("Error setting campaign meta %s: %v", cc.CampaignID, err)
			continue
		}

		if byFormat[cc.Format] == nil {
			byFormat[cc.Format] = make(map[string]float64)
		}
		// Rank candidates by bid (eCPM proxy).
		byFormat[cc.Format][cc.CampaignID] = float64(cc.BidCents)
	}

	// For each zone, set the candidate campaigns whose format matches.
	for _, z := range zones {
		candidates := byFormat[z.Format]
		if len(candidates) == 0 {
			continue
		}
		if err := r.redis.SetCampaignCandidates(ctx, z.ID, candidates); err != nil {
			log.Printf("Error setting candidates for zone %s: %v", z.ID, err)
		}
	}

	return nil
}

func (r *Reconciler) getActiveZones(ctx context.Context) ([]*zoneRow, error) {
	const query = `
		SELECT id, site_id, name, format, floor_price_cents, status,
		       array_to_string(countries, ',') as countries,
		       array_to_string(device_types, ',') as device_types,
		       array_to_string(os, ',') as os,
		       array_to_string(browsers, ',') as browsers,
		       array_to_string(carriers, ',') as carriers,
		       array_to_string(connection_types, ',') as connection_types
		FROM zones
		WHERE status = 'active'
	`
	rows, err := r.db.Pool().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*zoneRow
	for rows.Next() {
		var z zoneRow
		if err := rows.Scan(&z.ID, &z.SiteID, &z.Name, &z.Format, &z.FloorPriceCents, &z.Status,
			&z.Countries, &z.DeviceTypes, &z.OS, &z.Browsers, &z.Carriers, &z.ConnectionTypes); err != nil {
			return nil, err
		}
		zones = append(zones, &z)
	}
	return zones, nil
}

func (r *Reconciler) getApprovedCreative(ctx context.Context, campaignID string) (*campaignCreative, error) {
	const query = `
		SELECT id, format,
		       COALESCE(title, ''), COALESCE(body, ''),
		       COALESCE(icon_url, ''), COALESCE(image_url, ''), click_url
		FROM creatives
		WHERE campaign_id = $1 AND status = 'approved'
		ORDER BY created_at DESC
		LIMIT 1
	`
	var cc campaignCreative
	cc.CampaignID = campaignID
	err := r.db.Pool().QueryRow(ctx, query, campaignID).Scan(
		&cc.CreativeID, &cc.Format, &cc.Title, &cc.Body, &cc.IconURL, &cc.ImageURL, &cc.ClickURL,
	)
	if err != nil {
		return nil, err
	}
	return &cc, nil
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

func (r *Reconciler) CheckLowBalance(ctx context.Context) error {
	// Get all advertiser accounts
	advertisers, err := r.db.ListAccountsByType(ctx, "advertiser")
	if err != nil {
		return err
	}

	// Low balance threshold: KSh 3,000 (300,000 cents)
	const lowBalanceThreshold = 300000

	for _, advertiser := range advertisers {
		// Get wallet balance
		wallet, err := r.db.GetWallet(ctx, advertiser.ID)
		if err != nil || wallet == nil {
			continue
		}

		// Check if balance is below threshold
		if wallet.BalanceCents < lowBalanceThreshold {
			// Send low balance email
			go r.sendLowBalanceEmail(ctx, advertiser, wallet.BalanceCents, lowBalanceThreshold)
		}
	}

	return nil
}

func (r *Reconciler) sendLowBalanceEmail(ctx context.Context, account *postgres.Account, balanceCents, thresholdCents int64) {
	name := account.Email
	if account.CompanyName != nil && *account.CompanyName != "" {
		name = *account.CompanyName
	}

	amountKES := fmt.Sprintf("KSh %.2f", float64(balanceCents)/100)
	thresholdKES := fmt.Sprintf("KSh %.2f", float64(thresholdCents)/100)

	emailData := map[string]interface{}{
		"name":      name,
		"balance":   amountKES,
		"threshold": thresholdKES,
		"topupUrl":  "https://advertiser.otexads.com/wallet/topup",
	}

	emailReq := email.SendRequest{
		Template: email.TemplateLowBalance,
		To:       account.Email,
		Subject:  "Low balance warning — add funds to keep ads running",
		Data:     emailData,
	}

	reqBody, _ := json.Marshal(emailReq)
	httpReq, _ := http.NewRequest("POST", "http://emailer:8085/send", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Failed to send low balance email to %s: %v", account.Email, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Emailer returned status %d for low balance email to %s", resp.StatusCode, account.Email)
	} else {
		log.Printf("Low balance email sent to %s", account.Email)
	}
}
