package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"adnet/internal/auth"
	"adnet/internal/billing"
	"adnet/internal/email"
	"adnet/internal/mw"
	"adnet/internal/store/postgres"
	"adnet/internal/store/redis"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Config struct {
	PostgresURL      string
	RedisURL         string
	JWTSecret        string
	Port             string
	PaystackKey      string
	EmailRendererURL string
}

func main() {
	config := Config{
		PostgresURL:      getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/adnet?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:             getEnv("PORT", "8080"),
		PaystackKey:      getEnv("PAYSTACK_SECRET_KEY", ""),
		EmailRendererURL: getEnv("EMAIL_RENDERER_URL", "http://email-renderer:3000"),
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

	// Initialize email renderer client
	emailRenderer := email.NewRendererClient(config.EmailRendererURL)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	// Public routes (no auth required)
	public := api.PathPrefix("/v1").Subrouter()
	authHandler := NewAuthHandler(db, authService, emailRenderer)
	public.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	public.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	public.HandleFunc("/auth/refresh", authHandler.Refresh).Methods("POST", "OPTIONS")
	public.HandleFunc("/auth/verify-email", authHandler.RequestVerifyEmail).Methods("POST", "OPTIONS")
	public.HandleFunc("/auth/password-reset", authHandler.RequestPasswordReset).Methods("POST", "OPTIONS")
	public.HandleFunc("/auth/password-reset/confirm", authHandler.ConfirmPasswordReset).Methods("POST", "OPTIONS")
	public.HandleFunc("/targeting-options", authHandler.GetTargetingOptions).Methods("GET", "OPTIONS")

	// Protected routes (auth required)
	protected := api.PathPrefix("/v1").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	
	// Campaign routes (advertiser only)
	campaignHandler := NewCampaignHandler(db, redisClient)
	protected.HandleFunc("/campaigns", campaignHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/campaigns", campaignHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/campaigns/{id}", campaignHandler.Get).Methods("GET", "OPTIONS")
	protected.HandleFunc("/campaigns/{id}", campaignHandler.Update).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/campaigns/{id}/stats", campaignHandler.GetStats).Methods("GET", "OPTIONS")

	// Creative routes (advertiser only)
	creativeHandler := NewCreativeHandler(db)
	protected.HandleFunc("/creatives", creativeHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/creatives", creativeHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/creatives/{id}", creativeHandler.Delete).Methods("DELETE", "OPTIONS")

	// Zone routes (publisher only)
	zoneHandler := NewZoneHandler(db)
	protected.HandleFunc("/sites", zoneHandler.ListSites).Methods("GET", "OPTIONS")
	protected.HandleFunc("/sites", zoneHandler.CreateSite).Methods("POST", "OPTIONS")
	protected.HandleFunc("/sites/{id}", zoneHandler.UpdateSite).Methods("PUT", "PATCH", "OPTIONS")
	protected.HandleFunc("/sites/{id}", zoneHandler.DeleteSite).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/sites/{id}/zones", zoneHandler.CreateZone).Methods("POST", "OPTIONS")
	protected.HandleFunc("/zones", zoneHandler.ListZones).Methods("GET", "OPTIONS")
	protected.HandleFunc("/zones", zoneHandler.CreateZone).Methods("POST", "OPTIONS")
	protected.HandleFunc("/zones/{id}", zoneHandler.GetZone).Methods("GET", "OPTIONS")
	protected.HandleFunc("/zones/{id}", zoneHandler.UpdateZone).Methods("PUT", "PATCH", "OPTIONS")
	protected.HandleFunc("/zones/{id}", zoneHandler.DeleteZone).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/zones/{id}/stats", zoneHandler.GetZoneStats).Methods("GET", "OPTIONS")
	protected.HandleFunc("/zones/{id}/tag", zoneHandler.GetZoneTag).Methods("GET", "OPTIONS")

	// Wallet routes
	walletHandler := NewWalletHandler(db, redisClient, paystackClient)
	protected.HandleFunc("/wallet", walletHandler.GetWallet).Methods("GET", "OPTIONS")
	protected.HandleFunc("/wallet/topup", walletHandler.TopUp).Methods("POST", "OPTIONS")
	protected.HandleFunc("/wallet/verify", walletHandler.VerifyTopUp).Methods("GET", "OPTIONS")
	protected.HandleFunc("/wallet/transactions", walletHandler.GetTransactions).Methods("GET", "OPTIONS")

	// Payout routes (publisher only)
	payoutHandler := NewPayoutHandler(db, redisClient, paystackClient, emailRenderer)
	protected.HandleFunc("/payouts", payoutHandler.RequestPayout).Methods("POST", "OPTIONS")
	protected.HandleFunc("/payouts", payoutHandler.ListPayouts).Methods("GET", "OPTIONS")
	protected.HandleFunc("/payouts/balance", payoutHandler.GetBalance).Methods("GET", "OPTIONS")

	// Transfer recipient routes (publisher only)
	recipientHandler := NewRecipientHandler(db, paystackClient)
	protected.HandleFunc("/recipients", recipientHandler.SaveRecipient).Methods("POST", "OPTIONS")
	protected.HandleFunc("/recipients", recipientHandler.ListRecipients).Methods("GET", "OPTIONS")
	protected.HandleFunc("/recipients/default", recipientHandler.GetDefault).Methods("GET", "OPTIONS")

	// Admin routes (admin only)
	adminHandler := NewAdminHandler(db, redisClient, emailRenderer)
	protected.HandleFunc("/admin/stats", adminHandler.GetStats).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/analytics", adminHandler.GetDetailedAnalytics).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/users", adminHandler.ListUsers).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/campaigns", adminHandler.ListCampaigns).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/campaigns/{id}/status", adminHandler.UpdateCampaignStatus).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/admin/accounts/{id}/status", adminHandler.UpdateAccountStatus).Methods("PATCH", "OPTIONS")
	protected.HandleFunc("/admin/creatives/pending", adminHandler.ListPendingCreatives).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/creatives/moderate", adminHandler.ModerateCreative).Methods("POST", "OPTIONS")
	protected.HandleFunc("/admin/sites/pending", adminHandler.ListPendingSites).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/sites/moderate", adminHandler.ModerateSite).Methods("POST", "OPTIONS")
	protected.HandleFunc("/admin/campaigns/pending", adminHandler.ListPendingCampaigns).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/campaigns/moderate", adminHandler.ModerateCampaign).Methods("POST", "OPTIONS")

	// Admin financial routes
	financialHandler := NewFinancialHandler(db, paystackClient)
	protected.HandleFunc("/admin/financials", financialHandler.GetFinancials).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/financials/gateway-balance", financialHandler.GetGatewayBalance).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/financials/deposits", financialHandler.ListDeposits).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/financials/payouts", financialHandler.ListAllPayouts).Methods("GET", "OPTIONS")
	protected.HandleFunc("/admin/financials/payouts/{id}/retry", financialHandler.RetryPayout).Methods("POST", "OPTIONS")

	// Webhook route (unauthenticated — signature-verified)
	webhookHandler := NewWebhookHandler(db, redisClient, paystackClient, emailRenderer)
	api.HandleFunc("/webhooks/paystack", webhookHandler.PaystackWebhook).Methods("POST")

	// Marketplace routes (interconnection between advertisers and publishers)
	marketplaceHandler := NewMarketplaceHandler(db, redisClient)
	protected.HandleFunc("/marketplace/sites", marketplaceHandler.GetAvailableSites).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/zones", marketplaceHandler.GetAvailableZones).Methods("GET", "OPTIONS")
	protected.HandleFunc("/marketplace/campaigns", marketplaceHandler.GetActiveCampaigns).Methods("GET", "OPTIONS")

	// Configure CORS to allow requests from both portals
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
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
