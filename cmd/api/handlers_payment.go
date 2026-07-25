package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"adnet/internal/billing"
	"adnet/internal/mw"
	"adnet/internal/store/postgres"
	"adnet/internal/store/redis"
	"adnet/pkg/httpx"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// --- Webhook Handler (unauthenticated, signature-verified) ---

type WebhookHandler struct {
	db       *postgres.DB
	redis    *redis.Client
	paystack *billing.PaystackClient
}

func NewWebhookHandler(db *postgres.DB, redisClient *redis.Client, paystack *billing.PaystackClient) *WebhookHandler {
	return &WebhookHandler{db: db, redis: redisClient, paystack: paystack}
}

func (h *WebhookHandler) PaystackWebhook(w http.ResponseWriter, r *http.Request) {
	if h.paystack == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "Payment service not configured")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Verify webhook signature
	signature := r.Header.Get("x-paystack-signature")
	if signature == "" {
		httpx.Error(w, http.StatusUnauthorized, "Missing signature")
		return
	}

	mac := hmac.New(sha512.New, []byte(h.paystack.GetSecretKey()))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		httpx.Error(w, http.StatusUnauthorized, "Invalid signature")
		return
	}

	var event struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	var eventData struct {
		ID        int64  `json:"id"`
		Reference string `json:"reference"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
		Status    string `json:"status"`
	}
	json.Unmarshal(event.Data, &eventData)

	// Store event for audit (idempotent via paystack_id)
	paystackID := fmt.Sprintf("%d", eventData.ID)
	dbEvent, err := h.db.CreatePaystackEvent(
		r.Context(), event.Event, &paystackID, &eventData.Reference,
		&eventData.Amount, &eventData.Currency, &eventData.Status, body,
	)
	if err != nil {
		log.Printf("Failed to store paystack event: %v", err)
	}
	if dbEvent == nil {
		w.WriteHeader(http.StatusOK) // duplicate, already processed
		return
	}

	ctx := r.Context()

	switch event.Event {
	case "charge.success":
		h.handleChargeSuccess(ctx, eventData.Reference, eventData.Amount, dbEvent.ID)
	case "transfer.success":
		h.handleTransferSuccess(ctx, eventData.Reference, dbEvent.ID)
	case "transfer.failed", "transfer.reversed":
		h.handleTransferFailed(ctx, eventData.Reference, event.Event, dbEvent.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) handleChargeSuccess(ctx context.Context, reference string, amount int64, eventID uuid.UUID) {
	// Try to get existing transaction first
	tx, err := h.db.GetTransactionByReference(ctx, reference)
	
	// If transaction doesn't exist, we need to extract account_id from the reference
	// Reference format: adnet-topup-{account_id}-{timestamp}
	var accountID uuid.UUID
	if err != nil || tx == nil {
		// Parse account ID from reference
		// Format: adnet-topup-{account_id}-{timestamp}
		parts := strings.Split(reference, "-")
		if len(parts) >= 3 {
			accountID, err = uuid.Parse(parts[2])
			if err != nil {
				log.Printf("Webhook charge.success: failed to parse account ID from ref %s: %v", reference, err)
				return
			}
		} else {
			log.Printf("Webhook charge.success: invalid reference format %s", reference)
			return
		}
	} else {
		accountID = tx.AccountID
	}

	// Create transaction record if it doesn't exist
	if tx == nil {
		ref := reference
		_, err = h.db.CreateTransaction(ctx, accountID, "topup", amount, &ref)
		if err != nil {
			log.Printf("Webhook charge.success: failed to create transaction for ref %s: %v", reference, err)
			return
		}
	}

	_, err = h.db.UpdateWalletBalance(ctx, accountID, amount)
	if err != nil {
		log.Printf("Webhook charge.success: failed to credit wallet for account %s: %v", accountID, err)
		return
	}

	h.db.MarkPaystackEventProcessed(ctx, eventID)
	log.Printf("Webhook charge.success: credited %d cents to account %s (ref: %s)", amount, accountID, reference)
}

func (h *WebhookHandler) handleTransferSuccess(ctx context.Context, reference string, eventID uuid.UUID) {
	payout, err := h.db.GetPayoutByPaystackReference(ctx, reference)
	if err != nil || payout == nil {
		log.Printf("Webhook transfer.success: payout not found for ref %s", reference)
		return
	}

	h.db.UpdatePayoutRequestStatus(ctx, payout.ID, "paid", nil)
	h.db.MarkPaystackEventProcessed(ctx, eventID)
	log.Printf("Webhook transfer.success: payout %s marked as paid (ref: %s)", payout.ID, reference)
}

func (h *WebhookHandler) handleTransferFailed(ctx context.Context, reference, eventType string, eventID uuid.UUID) {
	payout, err := h.db.GetPayoutByPaystackReference(ctx, reference)
	if err != nil || payout == nil {
		log.Printf("Webhook %s: payout not found for ref %s", eventType, reference)
		return
	}

	h.db.UpdatePayoutRequestStatus(ctx, payout.ID, "failed", nil)

	// Refund locked funds back to publisher's wallet
	_, err = h.db.UpdateWalletBalance(ctx, payout.PublisherID, payout.AmountCents)
	if err != nil {
		log.Printf("Webhook %s: failed to refund wallet for publisher %s: %v", eventType, payout.PublisherID, err)
	}

	ref := fmt.Sprintf("refund-%s", reference)
	h.db.CreateTransaction(ctx, payout.PublisherID, "refund", payout.AmountCents, &ref)

	h.db.MarkPaystackEventProcessed(ctx, eventID)
	log.Printf("Webhook %s: payout %s failed, refunded %d cents to publisher %s", eventType, payout.ID, payout.AmountCents, payout.PublisherID)
}

// --- Publisher Transfer Recipient Handler ---

type RecipientHandler struct {
	db       *postgres.DB
	paystack *billing.PaystackClient
}

func NewRecipientHandler(db *postgres.DB, paystack *billing.PaystackClient) *RecipientHandler {
	return &RecipientHandler{db: db, paystack: paystack}
}

type SaveRecipientRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email,omitempty"`
	BankCode string `json:"bank_code"` // e.g. MPESA
}

func (h *RecipientHandler) SaveRecipient(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "publisher" {
		httpx.Error(w, http.StatusForbidden, "Publisher access required")
		return
	}

	if h.paystack == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "Payment service not configured")
		return
	}

	var req SaveRecipientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Phone == "" {
		httpx.Error(w, http.StatusBadRequest, "Name and phone are required")
		return
	}

	if req.BankCode == "" {
		req.BankCode = "MPESA"
	}

	// Register recipient with Paystack
	paystackResp, err := h.paystack.CreateTransferRecipient(billing.CreateTransferRecipientRequest{
		Type:     "mobile_money",
		Name:     req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
		Currency: "KES",
		BankCode: req.BankCode,
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to register payout details: "+err.Error())
		return
	}

	// Store in DB
	var phone, email, bankCode *string
	phone = &req.Phone
	if req.Email != "" {
		email = &req.Email
	}
	bankCode = &req.BankCode

	recipient, err := h.db.CreateTransferRecipient(
		r.Context(), accountID, paystackResp.Data.RecipientCode,
		"mobile_money", req.Name, phone, email, bankCode, "KES",
	)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to save recipient")
		return
	}

	httpx.JSON(w, http.StatusCreated, recipient)
}

func (h *RecipientHandler) ListRecipients(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipients, err := h.db.ListTransferRecipients(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, recipients)
}

func (h *RecipientHandler) GetDefault(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipient, err := h.db.GetDefaultTransferRecipient(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	if recipient == nil {
		httpx.JSON(w, http.StatusOK, map[string]interface{}{"configured": false})
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"configured": true,
		"recipient":  recipient,
	})
}

// --- Admin Financial Handlers ---

type FinancialHandler struct {
	db       *postgres.DB
	paystack *billing.PaystackClient
}

func NewFinancialHandler(db *postgres.DB, paystack *billing.PaystackClient) *FinancialHandler {
	return &FinancialHandler{db: db, paystack: paystack}
}

func (h *FinancialHandler) GetFinancials(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	summary, err := h.db.GetFinancialSummary(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, summary)
}

func (h *FinancialHandler) GetGatewayBalance(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	if h.paystack == nil {
		httpx.JSON(w, http.StatusOK, map[string]interface{}{
			"available": false,
			"balances":  []interface{}{},
		})
		return
	}

	balResp, err := h.paystack.GetBalance()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to fetch gateway balance: "+err.Error())
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"available": true,
		"balances":  balResp.Data,
	})
}

func (h *FinancialHandler) ListDeposits(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	deposits, err := h.db.ListAllDeposits(r.Context(), 100)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, deposits)
}

func (h *FinancialHandler) ListAllPayouts(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	payouts, err := h.db.ListAllPayouts(r.Context(), 100)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, payouts)
}

func (h *FinancialHandler) RetryPayout(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	vars := mux.Vars(r)
	payoutID, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid payout ID")
		return
	}

	// Get the payout request
	payouts, err := h.db.ListAllPayouts(r.Context(), 1000)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	var targetPayout *postgres.PayoutRequest
	for _, p := range payouts {
		if p.ID == payoutID {
			targetPayout = p
			break
		}
	}

	if targetPayout == nil {
		httpx.Error(w, http.StatusNotFound, "Payout not found")
		return
	}

	if targetPayout.Status != "failed" {
		httpx.Error(w, http.StatusBadRequest, "Can only retry failed payouts")
		return
	}

	// Reset status to pending for retry
	h.db.UpdatePayoutRequestStatus(r.Context(), payoutID, "pending", nil)

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "pending", "message": "Payout queued for retry"})
}
