package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"adnet/internal/auth"
	"adnet/internal/billing"
	"adnet/internal/mw"
	"adnet/internal/store/postgres"
	"adnet/internal/store/redis"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Config struct {
	PostgresURL   string
	RedisURL      string
	JWTSecret     string
	Port          string
	PaystackKey   string
}

func main() {
	config := Config{
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/adnet?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:        getEnv("PORT", "8080"),
		PaystackKey: getEnv("PAYSTACK_SECRET_KEY", ""),
	}

	ctx := context.Background()

	db, err := postgres.NewDB(ctx, config.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.NewClient(config.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	authService := auth.NewAuthService(auth.JWTConfig{
		SecretKey:     config.JWTSecret,
		AccessTokenTTL: 24 * time.Hour,
		RefreshTTL:    7 * 24 * time.Hour,
	})

	authMiddleware := mw.NewAuthMiddleware(authService)

	// Initialize Paystack client if key is provided
	var paystackClient *billing.PaystackClient
	if config.PaystackKey != "" {
		paystackClient = billing.NewPaystackClient(config.PaystackKey)
	}

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	
	// Auth routes
	authHandler := NewAuthHandler(db, authService)
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/refresh", authHandler.Refresh).Methods("POST", "OPTIONS")
	
	// Protected routes
	protected := api.PathPrefix("/v1").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	
	// Campaign routes (advertiser only)
	campaignHandler := NewCampaignHandler(db, redisClient)
	protected.HandleFunc("/campaigns", campaignHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/campaigns", campaignHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/campaigns/{id}", campaignHandler.Get).Methods("GET", "OPTIONS")
	protected.HandleFunc("/campaigns/{id}", campaignHandler.Update).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/campaigns/{id}/stats", campaignHandler.GetStats).Methods("GET", "OPTIONS")
	
	// Zone routes (publisher only)
	zoneHandler := NewZoneHandler(db)
	protected.HandleFunc("/sites", zoneHandler.ListSites).Methods("GET", "OPTIONS")
	protected.HandleFunc("/sites", zoneHandler.CreateSite).Methods("POST", "OPTIONS")
	protected.HandleFunc("/sites/{id}/zones", zoneHandler.CreateZone).Methods("POST", "OPTIONS")
	protected.HandleFunc("/zones/{id}", zoneHandler.GetZone).Methods("GET", "OPTIONS")
	protected.HandleFunc("/zones/{id}/stats", zoneHandler.GetZoneStats).Methods("GET", "OPTIONS")

	// Wallet routes
	walletHandler := NewWalletHandler(db, redisClient, paystackClient)
	protected.HandleFunc("/wallet", walletHandler.GetWallet).Methods("GET", "OPTIONS")
	protected.HandleFunc("/wallet/topup", walletHandler.TopUp).Methods("POST", "OPTIONS")
	protected.HandleFunc("/wallet/verify", walletHandler.VerifyTopUp).Methods("GET", "OPTIONS")
	protected.HandleFunc("/wallet/transactions", walletHandler.GetTransactions).Methods("GET", "OPTIONS")

	// Payout routes (publisher only)
	payoutHandler := NewPayoutHandler(db, redisClient)
	protected.HandleFunc("/payouts", payoutHandler.RequestPayout).Methods("POST", "OPTIONS")
	protected.HandleFunc("/payouts", payoutHandler.ListPayouts).Methods("GET", "OPTIONS")

	// Admin routes (admin only)
	adminHandler := NewAdminHandler(db, redisClient)
	protected.HandleFunc("/admin/creatives/pending", adminHandler.ListPendingCreatives).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/creatives/moderate", adminHandler.ModerateCreative).Methods("POST", "OPTIONS")
	protected.HandleFunc("/admin/accounts", adminHandler.ListAccounts).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/accounts/{id}/status", adminHandler.UpdateAccountStatus).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/admin/analytics", adminHandler.GetAnalytics).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/dashboard", adminHandler.GetRealtimeDashboard).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/cohort-analysis", adminHandler.GetCohortAnalysis).Methods("GET", "OPTIONS")

	// Marketplace routes (interconnection between advertisers and publishers)
	marketplaceHandler := NewMarketplaceHandler(db)
	protected.HandleFunc("/marketplace/sites", marketplaceHandler.GetAvailableSites).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/zones", marketplaceHandler.GetAvailableZones).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/campaigns", marketplaceHandler.GetActiveCampaigns).Methods("GET", "OPTIONS")

	// Configure CORS to allow requests from both portals
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://167.233.171.202:3000", "http://167.233.171.202:3001", "http://localhost:3000", "http://localhost:3001", "https://advertiser.otexads.com", "https://publisher.otexads.com", "https://api.otexads.com", "https://otexads.com", "https://www.otexads.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	srv := &http.Server{
		Handler:      handler,
		Addr:         ":" + config.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("API server starting on port %s", config.Port)
	log.Fatal(srv.ListenAndServe())
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
