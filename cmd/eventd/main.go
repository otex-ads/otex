package main

import (
	"context"
	"log"
	"os"
	"time"

	"adnet/internal/store/postgres"
	"adnet/internal/store/redis"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Config struct {
	RedisURL  string
	PostgresURL string
	BatchSize int
	FlushInterval time.Duration
}

func main() {
	config := Config{
		RedisURL:  getEnv("REDIS_URL", "localhost:6379"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/adnet?sslmode=disable"),
		BatchSize: 100,
		FlushInterval: 5 * time.Second,
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

	// Create consumer groups
	if err := redisClient.CreateConsumerGroup(ctx, "events:impressions", "eventd"); err != nil {
		log.Printf("Warning: failed to create consumer group for impressions: %v", err)
	}
	if err := redisClient.CreateConsumerGroup(ctx, "events:clicks", "eventd"); err != nil {
		log.Printf("Warning: failed to create consumer group for clicks: %v", err)
	}

	consumerID := uuid.New().String()
	log.Printf("eventd starting with consumer ID: %s", consumerID)

	// Start impression processor
	go processStream(ctx, redisClient, db, "events:impressions", "eventd", consumerID, config.BatchSize, config.FlushInterval, processImpression)

	// Start click processor
	go processStream(ctx, redisClient, db, "events:clicks", "eventd", consumerID, config.BatchSize, config.FlushInterval, processClick)

	// Keep running
	select {}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type EventProcessor func(ctx context.Context, tx pgx.Tx, event map[string]interface{}) error

func processStream(ctx context.Context, redisClient *redis.Client, db *postgres.DB, stream, group, consumerID string, batchSize int, flushInterval time.Duration, processor EventProcessor) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	batch := make([]map[string]interface{}, 0, batchSize)
	ids := make([]string, 0, batchSize)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if len(batch) > 0 {
				if err := flushBatch(ctx, redisClient, db, stream, group, batch, ids, processor); err != nil {
					log.Printf("Error flushing batch: %v", err)
				}
				batch = batch[:0]
				ids = ids[:0]
			}
		default:
			streams, err := redisClient.ReadEventsGroup(ctx, stream, group, consumerID, int64(batchSize))
			if err != nil {
				log.Printf("Error reading from stream %s: %v", stream, err)
				time.Sleep(1 * time.Second)
				continue
			}

			for _, streamData := range streams {
				for _, message := range streamData.Messages {
					batch = append(batch, message.Values)
					ids = append(ids, message.ID)

					if len(batch) >= batchSize {
						if err := flushBatch(ctx, redisClient, db, stream, group, batch, ids, processor); err != nil {
							log.Printf("Error flushing batch: %v", err)
						}
						batch = batch[:0]
						ids = ids[:0]
					}
				}
			}
		}
	}
}

func flushBatch(ctx context.Context, redisClient *redis.Client, db *postgres.DB, stream, group string, batch []map[string]interface{}, ids []string, processor EventProcessor) error {
	tx, err := db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, event := range batch {
		if err := processor(ctx, tx, event); err != nil {
			log.Printf("Error processing event: %v", err)
			continue
		}
		// Acknowledge event
		if err := redisClient.AckEvent(ctx, stream, group, ids[i]); err != nil {
			log.Printf("Error acknowledging event: %v", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Printf("Flushed %d events to Postgres", len(batch))
	return nil
}

func processImpression(ctx context.Context, tx pgx.Tx, event map[string]interface{}) error {
	const query = `
		INSERT INTO impressions (campaign_id, zone_id, creative_id, user_hash, country, device_type, cost_cents, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	campaignID := event["campaign_id"].(string)
	zoneID := event["zone_id"].(string)
	creativeID := event["creative_id"].(string)
	userHash := event["user_hash"].(string)
	country := event["country"].(string)
	deviceType := event["device_type"].(string)
	costCents := int64(event["cost_cents"].(float64))
	timestamp := time.Unix(int64(event["timestamp"].(float64)), 0)

	_, err := tx.Exec(ctx, query, campaignID, zoneID, creativeID, userHash, country, deviceType, costCents, timestamp)
	if err != nil {
		return err
	}

	// Deduct from advertiser wallet
	if err := deductSpend(ctx, tx, campaignID, costCents); err != nil {
		return err
	}

	// Credit publisher wallet (75% revenue share)
	if err := creditPublisher(ctx, tx, zoneID, costCents*75/100); err != nil {
		return err
	}

	return nil
}

func processClick(ctx context.Context, tx pgx.Tx, event map[string]interface{}) error {
	const query = `
		INSERT INTO clicks (campaign_id, zone_id, creative_id, user_hash, cost_cents, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	campaignID := event["campaign_id"].(string)
	zoneID := event["zone_id"].(string)
	creativeID := event["creative_id"].(string)
	userHash := event["user_hash"].(string)
	costCents := int64(event["cost_cents"].(float64))
	timestamp := time.Unix(int64(event["timestamp"].(float64)), 0)

	_, err := tx.Exec(ctx, query, campaignID, zoneID, creativeID, userHash, costCents, timestamp)
	if err != nil {
		return err
	}

	// Deduct from advertiser wallet
	if err := deductSpend(ctx, tx, campaignID, costCents); err != nil {
		return err
	}

	// Credit publisher wallet (75% revenue share)
	if err := creditPublisher(ctx, tx, zoneID, costCents*75/100); err != nil {
		return err
	}

	return nil
}

func deductSpend(ctx context.Context, tx pgx.Tx, campaignID string, costCents int64) error {
	// Get advertiser ID from campaign
	const getCampaignQuery = `
		SELECT advertiser_id FROM campaigns WHERE id = $1
	`
	var advertiserID string
	if err := tx.QueryRow(ctx, getCampaignQuery, campaignID).Scan(&advertiserID); err != nil {
		return err
	}

	// Update wallet balance
	const updateWalletQuery = `
		UPDATE wallets
		SET balance_cents = balance_cents - $1
		WHERE account_id = $2
	`
	_, err := tx.Exec(ctx, updateWalletQuery, costCents, advertiserID)
	if err != nil {
		return err
	}

	// Record transaction
	const insertTxQuery = `
		INSERT INTO transactions (account_id, type, amount_cents)
		VALUES ($1, 'spend', $2)
	`
	_, err = tx.Exec(ctx, insertTxQuery, advertiserID, costCents)
	return err
}

func creditPublisher(ctx context.Context, tx pgx.Tx, zoneID string, revenueCents int64) error {
	// Get publisher ID from zone
	const getPublisherQuery = `
		SELECT s.publisher_id
		FROM zones z
		INNER JOIN sites s ON z.site_id = s.id
		WHERE z.id = $1
	`
	var publisherID string
	if err := tx.QueryRow(ctx, getPublisherQuery, zoneID).Scan(&publisherID); err != nil {
		return err
	}

	// Update publisher wallet balance
	const updateWalletQuery = `
		UPDATE wallets
		SET balance_cents = balance_cents + $1
		WHERE account_id = $2
	`
	_, err := tx.Exec(ctx, updateWalletQuery, revenueCents, publisherID)
	if err != nil {
		return err
	}

	// Record transaction
	const insertTxQuery = `
		INSERT INTO transactions (account_id, type, amount_cents)
		VALUES ($1, 'payout', $2)
	`
	_, err = tx.Exec(ctx, insertTxQuery, publisherID, revenueCents)
	return err
}
