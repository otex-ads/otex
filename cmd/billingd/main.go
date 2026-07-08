package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"adnet/internal/billing"
	"adnet/internal/store/postgres"

	"github.com/google/uuid"
)

type Config struct {
	PostgresURL   string
	MpesaConsumerKey    string
	MpesaConsumerSecret string
	MpesaEnvironment    string
	MpesaShortcode      string
	MpesaPasskey        string
	MpesaCallbackURL    string
	Port         string
}

func main() {
	config := Config{
		PostgresURL:   getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/adnet?sslmode=disable"),
		MpesaConsumerKey:    getEnv("MPESA_CONSUMER_KEY", ""),
		MpesaConsumerSecret: getEnv("MPESA_CONSUMER_SECRET", ""),
		MpesaEnvironment:    getEnv("MPESA_ENVIRONMENT", "sandbox"),
		MpesaShortcode:      getEnv("MPESA_SHORTCODE", "174379"),
		MpesaPasskey:        getEnv("MPESA_PASSKEY", "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c619"),
		MpesaCallbackURL:    getEnv("MPESA_CALLBACK_URL", "http://localhost:8083/callback"),
		Port:         getEnv("PORT", "8083"),
	}

	ctx := context.Background()

	db, err := postgres.NewDB(ctx, config.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	mpesaConfig := billing.MpesaConfig{
		ConsumerKey:    config.MpesaConsumerKey,
		ConsumerSecret: config.MpesaConsumerSecret,
		Environment:    config.MpesaEnvironment,
		Shortcode:      config.MpesaShortcode,
		Passkey:        config.MpesaPasskey,
		CallbackURL:    config.MpesaCallbackURL,
	}

	mpesaClient := billing.NewMpesaClient(mpesaConfig)

	billingService := &BillingService{
		db:         db,
		mpesa:      mpesaClient,
	}

	// Start HTTP server for callbacks
	go billingService.startCallbackServer(config.Port)

	// Start payout processor
	go billingService.processPendingPayouts(ctx)

	log.Printf("Billing service starting on port %s", config.Port)

	// Keep running
	select {}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type BillingService struct {
	db    *postgres.DB
	mpesa *billing.MpesaClient
}

func (s *BillingService) startCallbackServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback/stk", s.handleSTKCallback)
	mux.HandleFunc("/callback/b2c", s.handleB2CCallback)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Callback server starting on port %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Callback server error: %v", err)
	}
}

func (s *BillingService) handleSTKCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var callback billing.STKPushCallback
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		log.Printf("Error decoding STK callback: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Extract transaction details
	mpesaCode := billing.GetMpesaCodeFromCallback(&callback)
	amount := billing.GetAmountFromCallback(&callback)
	phoneNumber := billing.GetPhoneNumberFromCallback(&callback)

	log.Printf("STK Callback: Code=%s, Amount=%s, Phone=%s, ResultCode=%d", 
		mpesaCode, amount, phoneNumber, callback.Body.StkCallback.ResultCode)

	// If successful, credit the wallet
	if callback.Body.StkCallback.ResultCode == 0 {
		// In a real implementation, you'd need to identify which account this belongs to
		// This would typically be done via the MerchantRequestID or AccountReference
		log.Printf("Payment successful: %s KES", amount)
		// TODO: Credit wallet based on phone number or reference
	}

	w.WriteHeader(http.StatusOK)
}

func (s *BillingService) handleB2CCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse B2C callback
	var callback struct {
		Result struct {
			ResultCode         int    `json:"ResultCode"`
			ResultDesc         string `json:"ResultDesc"`
			OriginatorConversationID string `json:"OriginatorConversationID"`
			ConversationID     string `json:"ConversationID"`
			TransactionID      string `json:"TransactionID"`
			ReceiverPartyPublicName string `json:"ReceiverPartyPublicName"`
			TransactionAmount  string `json:"TransactionAmount"`
		} `json:"Result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		log.Printf("Error decoding B2C callback: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("B2C Callback: ResultCode=%d, ResultDesc=%s, Amount=%s", 
		callback.Result.ResultCode, callback.Result.ResultDesc, callback.Result.TransactionAmount)

	// Update payout request status based on result
	if callback.Result.ResultCode == 0 {
		// Success - update payout request to paid
		// TODO: Find payout request by reference and update status
		log.Printf("B2C payment successful: %s", callback.Result.TransactionID)
	} else {
		// Failed - update payout request to failed
		log.Printf("B2C payment failed: %s", callback.Result.ResultDesc)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *BillingService) processPendingPayouts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processPendingPayoutsBatch(ctx)
		}
	}
}

func (s *BillingService) processPendingPayoutsBatch(ctx context.Context) {
	// Get pending payout requests
	const query = `
		SELECT id, publisher_id, amount_cents
		FROM payout_requests
		WHERE status = 'pending'
		LIMIT 10
	`

	rows, err := s.db.Pool().Query(ctx, query)
	if err != nil {
		log.Printf("Error fetching pending payouts: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, publisherID uuid.UUID
		var amountCents int64
		if err := rows.Scan(&id, &publisherID, &amountCents); err != nil {
			log.Printf("Error scanning payout: %v", err)
			continue
		}

		// Get publisher phone number (would need to be stored in accounts table)
		// For now, we'll skip actual B2C calls since we don't have phone numbers
		log.Printf("Processing payout %s for publisher %s: %d cents", id, publisherID, amountCents)

		// In production:
		// 1. Get publisher phone number
		// 2. Call M-Pesa B2C API
		// 3. Update payout request status based on result

		// For now, mark as processing
		if _, err := s.db.UpdatePayoutRequestStatus(ctx, id, "processing", nil); err != nil {
			log.Printf("Error updating payout status: %v", err)
		}
	}
}
