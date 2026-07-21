package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"adnet/internal/store/redis"
)

func main() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	redisClient, err := redis.NewClient(redisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	ctx := context.Background()

	// Seed a test zone
	zoneID := "550e8400-e29b-41d4-a716-446655440000"
	zoneMeta := map[string]string{
		"site_id":         "550e8400-e29b-41d4-a716-446655440001",
		"name":            "Test Zone",
		"format":          "banner",
		"floor_price_cents": "100",
		"status":          "active",
	}
	if err := redisClient.SetZoneMeta(ctx, zoneID, zoneMeta); err != nil {
		log.Fatalf("Failed to set zone meta: %v", err)
	}

	// Seed a test campaign
	campaignID := "660e8400-e29b-41d4-a716-446655440000"
	campaignMeta := map[string]string{
		"bid_cents":      "500",
		"format":         "banner",
		"creative_id":    "770e8400-e29b-41d4-a716-446655440000",
		"title":          "Test Ad",
		"body":           "This is a test advertisement",
		"image_url":      "https://example.com/ad.jpg",
		"click_url":      "https://example.com/landing",
		"status":         "active",
		"creative_status": "approved",
	}
	if err := redisClient.SetCampaignMeta(ctx, campaignID, campaignMeta); err != nil {
		log.Fatalf("Failed to set campaign meta: %v", err)
	}

	// Set budget
	if err := redisClient.SetDailyBudget(ctx, campaignID, 10000); err != nil {
		log.Fatalf("Failed to set daily budget: %v", err)
	}
	if err := redisClient.SetTotalBudgetRemaining(ctx, campaignID, 100000); err != nil {
		log.Fatalf("Failed to set total budget: %v", err)
	}

	// Add campaign to zone candidates (eCPM score = 5.0)
	candidates := map[string]float64{
		campaignID: 5.0,
	}
	if err := redisClient.SetCampaignCandidates(ctx, zoneID, candidates); err != nil {
		log.Fatalf("Failed to set campaign candidates: %v", err)
	}

	fmt.Println("Redis seeded successfully!")
	fmt.Printf("Zone ID: %s\n", zoneID)
	fmt.Printf("Campaign ID: %s\n", campaignID)
	fmt.Println("\nTest the adserve endpoint:")
	fmt.Printf("curl 'http://localhost:8081/serve?zone=%s&format=banner'\n", zoneID)
}
