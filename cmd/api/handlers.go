package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"adnet/internal/auth"
	"adnet/internal/billing"
	"adnet/internal/email"
	"adnet/internal/mw"
	"adnet/internal/store/postgres"
	"adnet/internal/store/redis"
	"adnet/pkg/httpx"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AuthHandler struct {
	db          *postgres.DB
	authService *auth.AuthService
	emailer     *email.RendererClient
}

func NewAuthHandler(db *postgres.DB, authService *auth.AuthService, emailer *email.RendererClient) *AuthHandler {
	return &AuthHandler{db: db, authService: authService, emailer: emailer}
}

type RegisterRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	AccountType  string `json:"account_type"`
	CompanyName  string `json:"company_name,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AccountID    string `json:"account_id"`
	Email        string `json:"email"`
	Type         string `json:"type"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.AccountType == "" {
		httpx.Error(w, http.StatusBadRequest, "Email, password, and account_type are required")
		return
	}

	if req.AccountType != string(auth.AccountTypeAdvertiser) && 
	   req.AccountType != string(auth.AccountTypePublisher) &&
	   req.AccountType != string(auth.AccountTypeAdmin) {
		httpx.Error(w, http.StatusBadRequest, "Invalid account_type")
		return
	}

	existing, err := h.db.GetAccountByEmail(r.Context(), req.Email)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing != nil {
		httpx.Error(w, http.StatusConflict, "Email already registered")
		return
	}

	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	var companyName *string
	if req.CompanyName != "" {
		companyName = &req.CompanyName
	}

	account, err := h.db.CreateAccount(r.Context(), req.AccountType, req.Email, passwordHash, companyName)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	accessToken, err := h.authService.GenerateAccessToken(account.ID, account.Email, auth.AccountType(account.Type))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken(account.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Send welcome email asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		name := req.Email
		if req.CompanyName != "" {
			name = req.CompanyName
		}

		emailData := map[string]interface{}{
			"name":        name,
			"email":       req.Email,
			"role":        req.AccountType,
			"dashboardUrl": fmt.Sprintf("https://%s.otexads.com", req.AccountType),
			"verifyUrl":   fmt.Sprintf("https://%s.otexads.com/verify?token=%s", req.AccountType, accessToken),
		}

		// Render email HTML
		html, err := h.emailer.Render(ctx, email.TemplateWelcome, emailData)
		if err != nil {
			log.Printf("Failed to render welcome email: %v", err)
			return
		}

		// Send via emailer service
		emailReq := email.SendRequest{
			Template: email.TemplateWelcome,
			To:       req.Email,
			Subject:  "Welcome to OtexAds",
			Data:     emailData,
		}

		// POST to emailer service
		reqBody, _ := json.Marshal(emailReq)
		httpReq, _ := http.NewRequest("POST", "http://emailer:8085/send", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			log.Printf("Failed to send welcome email: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Emailer returned status %d", resp.StatusCode)
		} else {
			log.Printf("Welcome email sent to %s", req.Email)
		}
	}()

	httpx.JSON(w, http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccountID:    account.ID.String(),
		Email:        account.Email,
		Type:         account.Type,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	account, err := h.db.GetAccountByEmail(r.Context(), req.Email)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if account == nil {
		httpx.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if account.Status != string(auth.AccountStatusActive) {
		httpx.Error(w, http.StatusForbidden, "Account is not active")
		return
	}

	if !h.authService.CheckPassword(req.Password, account.PasswordHash) {
		httpx.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	accessToken, err := h.authService.GenerateAccessToken(account.ID, account.Email, auth.AccountType(account.Type))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken(account.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	httpx.JSON(w, http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccountID:    account.ID.String(),
		Email:        account.Email,
		Type:         account.Type,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	accountID, err := h.authService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	account, err := h.db.GetAccountByID(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if account == nil {
		httpx.Error(w, http.StatusUnauthorized, "Account not found")
		return
	}

	accessToken, err := h.authService.GenerateAccessToken(account.ID, account.Email, auth.AccountType(account.Type))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	newRefreshToken, err := h.authService.GenerateRefreshToken(account.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	httpx.JSON(w, http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		AccountID:    account.ID.String(),
		Email:        account.Email,
		Type:         account.Type,
	})
}

type CampaignHandler struct {
	db    *postgres.DB
	redis *redis.Client
}

func NewCampaignHandler(db *postgres.DB, redisClient *redis.Client) *CampaignHandler {
	return &CampaignHandler{db: db, redis: redisClient}
}

type Campaign struct {
	ID              uuid.UUID  `json:"id"`
	AdvertiserID    uuid.UUID  `json:"advertiser_id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	PricingModel    string     `json:"pricing_model"`
	BidAmountCents  int        `json:"bid_amount_cents"`
	DailyBudgetCents int       `json:"daily_budget_cents"`
	TotalBudgetCents int       `json:"total_budget_cents"`
	Timezone        string     `json:"timezone"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateCampaignRequest struct {
	Name            string     `json:"name"`
	PricingModel    string     `json:"pricing_model"`
	BidAmountCents  int        `json:"bid_amount_cents"`
	DailyBudgetCents int       `json:"daily_budget_cents"`
	TotalBudgetCents int       `json:"total_budget_cents"`
	Timezone        string     `json:"timezone"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	Format          string     `json:"format"`
	Targeting       *struct {
		Countries []string `json:"countries"`
		Devices   []string `json:"devices"`
		OS        []string `json:"os"`
	} `json:"targeting"`
	CreativeID      *string    `json:"creativeId,omitempty"`
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.PricingModel == "" || req.BidAmountCents <= 0 {
		httpx.Error(w, http.StatusBadRequest, "Invalid campaign parameters")
		return
	}

	campaign, err := h.db.CreateCampaign(r.Context(), accountID, req.Name, "pending",
		req.PricingModel, req.BidAmountCents, req.DailyBudgetCents, req.TotalBudgetCents,
		req.Timezone, req.StartsAt, req.EndsAt)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}

	// Create targeting rule if provided
	if req.Targeting != nil {
		_, err = h.db.CreateTargetingRule(r.Context(), campaign.ID, req.Targeting.Countries, req.Targeting.Devices, req.Targeting.OS)
		if err != nil {
			// Log error but don't fail the request - targeting is optional
			// Campaign was created successfully
		}
	}

	// Set initial campaign metadata in Redis
	campaignMeta := map[string]string{
		"status":         "pending",
		"creative_status": "pending",
		"format":         req.Format,
	}
	if req.CreativeID != nil {
		campaignMeta["creative_id"] = *req.CreativeID
	}
	if err := h.redis.SetCampaignMeta(r.Context(), campaign.ID.String(), campaignMeta); err != nil {
		// Log error but don't fail the request
		// Redis update is best-effort
	}

	// Publish campaign update to Redis
	h.redis.PublishCampaignUpdate(r.Context(), campaign.ID.String())

	httpx.JSON(w, http.StatusCreated, campaign)
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	campaigns, err := h.db.ListCampaignsByAdvertiser(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to list campaigns")
		return
	}

	httpx.JSON(w, http.StatusOK, campaigns)
}

func (h *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	campaign, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if campaign == nil {
		httpx.Error(w, http.StatusNotFound, "Campaign not found")
		return
	}

	httpx.JSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	_, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Get date range from query params
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Query impressions table
	impressionsQuery := `
		SELECT COALESCE(COUNT(*), 0) as count, COALESCE(SUM(cost_cents), 0) as spend
		FROM impressions
		WHERE campaign_id = $1 AND is_fraud = false
	`

	var impArgs []interface{}
	impArgs = append(impArgs, id)

	if startDate != "" {
		impressionsQuery += " AND DATE(occurred_at) >= $" + string(rune(len(impArgs)+1))
		impArgs = append(impArgs, startDate)
	}
	if endDate != "" {
		impressionsQuery += " AND DATE(occurred_at) <= $" + string(rune(len(impArgs)+1))
		impArgs = append(impArgs, endDate)
	}

	var impressionsCount, impressionsSpend int64
	err = h.db.Pool().QueryRow(r.Context(), impressionsQuery, impArgs...).Scan(&impressionsCount, &impressionsSpend)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Query clicks table
	clicksQuery := `
		SELECT COALESCE(COUNT(*), 0) as count, COALESCE(SUM(cost_cents), 0) as spend
		FROM clicks
		WHERE campaign_id = $1 AND is_fraud = false
	`

	var clickArgs []interface{}
	clickArgs = append(clickArgs, id)

	if startDate != "" {
		clicksQuery += " AND DATE(occurred_at) >= $" + string(rune(len(clickArgs)+1))
		clickArgs = append(clickArgs, startDate)
	}
	if endDate != "" {
		clicksQuery += " AND DATE(occurred_at) <= $" + string(rune(len(clickArgs)+1))
		clickArgs = append(clickArgs, endDate)
	}

	var clicksCount, clicksSpend int64
	err = h.db.Pool().QueryRow(r.Context(), clicksQuery, clickArgs...).Scan(&clicksCount, &clicksSpend)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Query conversions table
	conversionsQuery := `
		SELECT COALESCE(COUNT(*), 0) as count, COALESCE(SUM(payout_cents), 0) as payout
		FROM conversions
		WHERE campaign_id = $1
	`

	var convArgs []interface{}
	convArgs = append(convArgs, id)

	if startDate != "" {
		conversionsQuery += " AND DATE(occurred_at) >= $" + string(rune(len(convArgs)+1))
		convArgs = append(convArgs, startDate)
	}
	if endDate != "" {
		conversionsQuery += " AND DATE(occurred_at) <= $" + string(rune(len(convArgs)+1))
		convArgs = append(convArgs, endDate)
	}

	var conversionsCount, conversionsPayout int64
	err = h.db.Pool().QueryRow(r.Context(), conversionsQuery, convArgs...).Scan(&conversionsCount, &conversionsPayout)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Calculate CTR
	var ctr float64
	if impressionsCount > 0 {
		ctr = float64(clicksCount) / float64(impressionsCount) * 100
	}

	stats := map[string]interface{}{
		"impressions": impressionsCount,
		"clicks":      clicksCount,
		"conversions": conversionsCount,
		"ctr":         ctr,
		"spend_cents": impressionsSpend + clicksSpend,
	}

	httpx.JSON(w, http.StatusOK, stats)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	campaign, err := h.db.UpdateCampaign(r.Context(), id, req.Name, "draft",
		req.PricingModel, req.BidAmountCents, req.DailyBudgetCents, req.TotalBudgetCents,
		req.Timezone, req.StartsAt, req.EndsAt)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	// Publish campaign update to Redis
	h.redis.PublishCampaignUpdate(r.Context(), campaign.ID.String())

	httpx.JSON(w, http.StatusOK, campaign)
}

type ZoneHandler struct {
	db *postgres.DB
}

func NewZoneHandler(db *postgres.DB) *ZoneHandler {
	return &ZoneHandler{db: db}
}

type Site struct {
	ID           uuid.UUID `json:"id"`
	PublisherID  uuid.UUID `json:"publisher_id"`
	Domain       string    `json:"domain"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Zone struct {
	ID             uuid.UUID `json:"id"`
	SiteID         uuid.UUID `json:"site_id"`
	Name           string    `json:"name"`
	Format         string    `json:"format"`
	FloorPriceCents int      `json:"floor_price_cents"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateSiteRequest struct {
	Domain string `json:"domain"`
}

type CreateZoneRequest struct {
	SiteID          uuid.UUID `json:"siteId"`
	Name            string    `json:"name"`
	Format          string    `json:"format"`
	Size            string    `json:"size"`
	FloorPriceCents int       `json:"floor_price_cents"`
}

type UpdateZoneRequest struct {
	Name            string `json:"name"`
	Format          string `json:"format"`
	Size            string `json:"size"`
	Status          string `json:"status"`
	FloorPriceCents int    `json:"floor_price_cents"`
}

func (h *ZoneHandler) CreateSite(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Domain == "" {
		httpx.Error(w, http.StatusBadRequest, "Domain is required")
		return
	}

	site, err := h.db.CreateSite(r.Context(), accountID, req.Domain, "pending_review")
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to create site")
		return
	}

	httpx.JSON(w, http.StatusCreated, site)
}

func (h *ZoneHandler) ListSites(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sites, err := h.db.ListSitesByPublisher(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to list sites")
		return
	}

	httpx.JSON(w, http.StatusOK, sites)
}

type UpdateSiteRequest struct {
	Domain string `json:"domain"`
}

func (h *ZoneHandler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	siteID, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	existing, err := h.db.GetSiteByID(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing == nil {
		httpx.Error(w, http.StatusNotFound, "Site not found")
		return
	}
	if existing.PublisherID != accountID {
		httpx.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	var req UpdateSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	domain := req.Domain
	if domain == "" {
		domain = existing.Domain
	}

	site, err := h.db.UpdateSite(r.Context(), siteID, domain)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update site")
		return
	}

	httpx.JSON(w, http.StatusOK, site)
}

func (h *ZoneHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	siteID, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	existing, err := h.db.GetSiteByID(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing == nil {
		httpx.Error(w, http.StatusNotFound, "Site not found")
		return
	}
	if existing.PublisherID != accountID {
		httpx.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.db.DeleteSite(r.Context(), siteID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to delete site")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ZoneHandler) GetZoneTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneID, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid zone ID")
		return
	}

	// Verify zone exists
	zone, err := h.db.GetZoneByID(r.Context(), zoneID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if zone == nil {
		httpx.Error(w, http.StatusNotFound, "Zone not found")
		return
	}

	// Return the ad tag HTML snippet
	tagHTML := fmt.Sprintf(`<script async src="https://cdn.otexads.com/tag.js" data-zone-id="%s"></script>`, zoneID)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(tagHTML))
}

func (h *ZoneHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	zones, err := h.db.ListZonesByPublisher(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to list zones")
		return
	}

	httpx.JSON(w, http.StatusOK, zones)
}

func (h *ZoneHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Site ID may come from the URL (/sites/{id}/zones) or the request body (/zones)
	siteID := req.SiteID
	if vars := mux.Vars(r); vars["id"] != "" {
		if parsed, err := uuid.Parse(vars["id"]); err == nil {
			siteID = parsed
		}
	}
	if siteID == uuid.Nil {
		httpx.Error(w, http.StatusBadRequest, "Site ID is required")
		return
	}

	if req.Name == "" || req.Format == "" {
		httpx.Error(w, http.StatusBadRequest, "Name and format are required")
		return
	}

	// Verify the site belongs to the requesting publisher
	site, err := h.db.GetSiteByID(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if site == nil {
		httpx.Error(w, http.StatusNotFound, "Site not found")
		return
	}
	if site.PublisherID != accountID {
		httpx.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	zone, err := h.db.CreateZone(r.Context(), siteID, req.Name, req.Format, req.FloorPriceCents, "active")
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to create zone")
		return
	}

	httpx.JSON(w, http.StatusCreated, zone)
}

func (h *ZoneHandler) UpdateZone(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	zoneID, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid zone ID")
		return
	}

	existing, err := h.db.GetZoneByID(r.Context(), zoneID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing == nil {
		httpx.Error(w, http.StatusNotFound, "Zone not found")
		return
	}

	// Verify ownership via the parent site
	site, err := h.db.GetSiteByID(r.Context(), existing.SiteID)
	if err != nil || site == nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if site.PublisherID != accountID {
		httpx.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	var req UpdateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name := req.Name
	if name == "" {
		name = existing.Name
	}
	format := req.Format
	if format == "" {
		format = existing.Format
	}
	status := req.Status
	if status == "" {
		status = existing.Status
	}
	floor := req.FloorPriceCents
	if floor == 0 {
		floor = existing.FloorPriceCents
	}

	zone, err := h.db.UpdateZone(r.Context(), zoneID, name, format, floor, status)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update zone")
		return
	}

	httpx.JSON(w, http.StatusOK, zone)
}

func (h *ZoneHandler) DeleteZone(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	zoneID, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid zone ID")
		return
	}

	existing, err := h.db.GetZoneByID(r.Context(), zoneID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing == nil {
		httpx.Error(w, http.StatusNotFound, "Zone not found")
		return
	}

	site, err := h.db.GetSiteByID(r.Context(), existing.SiteID)
	if err != nil || site == nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if site.PublisherID != accountID {
		httpx.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.db.DeleteZone(r.Context(), zoneID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to delete zone")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *ZoneHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid zone ID")
		return
	}

	zone, err := h.db.GetZoneByID(r.Context(), id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if zone == nil {
		httpx.Error(w, http.StatusNotFound, "Zone not found")
		return
	}

	httpx.JSON(w, http.StatusOK, zone)
}

func (h *ZoneHandler) GetZoneStats(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid zone ID")
		return
	}

	// Get date range from query params
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	query := `
		SELECT 
			COALESCE(SUM(impressions), 0) as impressions,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(revenue_cents), 0) as revenue_cents,
			COUNT(DISTINCT DATE(created_at)) as days
		FROM stats
		WHERE zone_id = $1 AND publisher_id = $2
	`

	var args []interface{}
	args = append(args, id, accountID)

	if startDate != "" {
		query += " AND DATE(created_at) >= $" + string(rune(len(args)+1))
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND DATE(created_at) <= $" + string(rune(len(args)+1))
		args = append(args, endDate)
	}

	var stats struct {
		Impressions   int64 `json:"impressions"`
		Clicks        int64 `json:"clicks"`
		RevenueCents  int64 `json:"revenue_cents"`
		Days          int   `json:"days"`
	}

	err = h.db.Pool().QueryRow(r.Context(), query, args...).Scan(
		&stats.Impressions, &stats.Clicks, &stats.RevenueCents, &stats.Days,
	)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, stats)
}

type WalletHandler struct {
	db        *postgres.DB
	redis     *redis.Client
	paystack  *billing.PaystackClient
}

func NewWalletHandler(db *postgres.DB, redisClient *redis.Client, paystack *billing.PaystackClient) *WalletHandler {
	return &WalletHandler{db: db, redis: redisClient, paystack: paystack}
}

type TopUpRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Email       string `json:"email"`
	Channel     string `json:"channel,omitempty"` // card, bank, ussd, qr, mobile_money
}

type TopUpResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	Reference        string `json:"reference"`
	AccessCode       string `json:"access_code"`
}

func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	wallet, err := h.db.GetWallet(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if wallet == nil {
		// Create wallet if it doesn't exist
		wallet, err = h.db.CreateWallet(r.Context(), accountID, 0, "KES")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Failed to create wallet")
			return
		}
	}

	httpx.JSON(w, http.StatusOK, wallet)
}

func (h *WalletHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AmountCents <= 0 || req.Email == "" {
		httpx.Error(w, http.StatusBadRequest, "Invalid top-up parameters")
		return
	}

	// Generate unique reference
	reference := fmt.Sprintf("adnet-topup-%s-%d", accountID.String(), time.Now().Unix())

	// Initialize Paystack transaction
	paystackReq := billing.InitializeTransactionRequest{
		Amount:      fmt.Sprintf("%d", req.AmountCents),
		Email:       req.Email,
		Currency:    "KES",
		Reference:   reference,
		CallbackURL: "https://advertiser.otexads.com/wallet?verify=true",
		Metadata: map[string]interface{}{
			"account_id": accountID.String(),
			"type":       "wallet_topup",
		},
	}

	if req.Channel != "" {
		paystackReq.Channels = []string{req.Channel}
	}

	paystackResp, err := h.paystack.InitializeTransaction(paystackReq)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to initialize payment: "+err.Error())
		return
	}

	// Don't create transaction record yet - only credit when payment is successful via webhook
	// Transaction will be created in the webhook handler on charge.success event

	response := TopUpResponse{
		AuthorizationURL: paystackResp.Data.AuthorizationURL,
		Reference:        paystackResp.Data.Reference,
		AccessCode:       paystackResp.Data.AccessCode,
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *WalletHandler) VerifyTopUp(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	reference := r.URL.Query().Get("reference")
	if reference == "" {
		httpx.Error(w, http.StatusBadRequest, "Reference is required")
		return
	}

	// Verify transaction with Paystack
	paystackResp, err := h.paystack.VerifyTransaction(reference)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to verify payment: "+err.Error())
		return
	}

	// Check if transaction is successful
	if paystackResp.Data.Status != "success" {
		httpx.Error(w, http.StatusBadRequest, "Payment not successful")
		return
	}

	// Update wallet balance
	wallet, err := h.db.GetWallet(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if wallet == nil {
		wallet, err = h.db.CreateWallet(r.Context(), accountID, 0, "KES")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Failed to create wallet")
			return
		}
	}

	// Add funds to wallet (delta-based)
	_, err = h.db.UpdateWalletBalance(r.Context(), accountID, paystackResp.Data.Amount)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update wallet")
		return
	}

	// Re-fetch updated balance
	updatedWallet, _ := h.db.GetWallet(r.Context(), accountID)
	newBal := int64(0)
	if updatedWallet != nil {
		newBal = updatedWallet.BalanceCents
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"status":      "success",
		"amount":      paystackResp.Data.Amount,
		"new_balance": newBal,
	})
}

func (h *WalletHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 50
	transactions, err := h.db.ListTransactions(r.Context(), accountID, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, transactions)
}

type PayoutHandler struct {
	db       *postgres.DB
	redis    *redis.Client
	paystack *billing.PaystackClient
}

func NewPayoutHandler(db *postgres.DB, redisClient *redis.Client, paystack *billing.PaystackClient) *PayoutHandler {
	return &PayoutHandler{db: db, redis: redisClient, paystack: paystack}
}

type PayoutRequestBody struct {
	AmountCents int64 `json:"amount_cents"`
}

func (h *PayoutHandler) RequestPayout(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req PayoutRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AmountCents <= 0 {
		httpx.Error(w, http.StatusBadRequest, "Invalid amount")
		return
	}

	// Check wallet balance
	wallet, err := h.db.GetWallet(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if wallet == nil || wallet.BalanceCents < req.AmountCents {
		httpx.Error(w, http.StatusBadRequest, "Insufficient balance")
		return
	}

	// Create payout request
	payout, err := h.db.CreatePayoutRequest(r.Context(), accountID, req.AmountCents)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to create payout request")
		return
	}

	// Lock funds by deducting from wallet
	_, err = h.db.UpdateWalletBalance(r.Context(), accountID, -req.AmountCents)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to lock funds")
		return
	}

	// Record the payout transaction
	ref := fmt.Sprintf("payout-%s", payout.ID.String())
	h.db.CreateTransaction(r.Context(), accountID, "payout", -req.AmountCents, &ref)

	// If Paystack is configured, initiate transfer
	if h.paystack != nil {
		recipient, err := h.db.GetDefaultTransferRecipient(r.Context(), accountID)
		if err == nil && recipient != nil {
			paystackRef := fmt.Sprintf("adnet-payout-%s-%d", payout.ID.String(), time.Now().Unix())

			transferResp, err := h.paystack.InitiateTransfer(billing.InitiateTransferRequest{
				Source:    "balance",
				Amount:    req.AmountCents,
				Recipient: recipient.RecipientCode,
				Reason:    "Publisher payout",
				Reference: paystackRef,
			})

			if err != nil {
				// Transfer initiation failed — refund wallet, mark failed
				h.db.UpdateWalletBalance(r.Context(), accountID, req.AmountCents)
				h.db.UpdatePayoutRequestStatus(r.Context(), payout.ID, "failed", nil)
				refundRef := fmt.Sprintf("refund-%s", paystackRef)
				h.db.CreateTransaction(r.Context(), accountID, "refund", req.AmountCents, &refundRef)
				httpx.Error(w, http.StatusInternalServerError, "Failed to initiate transfer: "+err.Error())
				return
			}

			// Update payout with Paystack details
			h.db.UpdatePayoutWithPaystack(r.Context(), payout.ID, "processing",
				transferResp.Data.TransferCode, paystackRef, recipient.RecipientCode)
		}
	}

	httpx.JSON(w, http.StatusCreated, payout)
}

func (h *PayoutHandler) ListPayouts(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	payouts, err := h.db.ListPayoutRequests(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, payouts)
}

func (h *PayoutHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	accountID, ok := mw.AccountIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	wallet, err := h.db.GetWallet(r.Context(), accountID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	if wallet == nil {
		httpx.JSON(w, http.StatusOK, map[string]interface{}{
			"available": 0,
			"pending":   0,
			"totalPaid": 0,
		})
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"available": wallet.BalanceCents,
		"pending":   0,
		"totalPaid": 0,
	})
}

type AdminHandler struct {
	db    *postgres.DB
	redis *redis.Client
}

func NewAdminHandler(db *postgres.DB, redisClient *redis.Client) *AdminHandler {
	return &AdminHandler{db: db, redis: redisClient}
}

type ModerationRequest struct {
	CreativeID uuid.UUID `json:"creative_id"`
	Action     string    `json:"action"` // "approve" or "reject"
	Reason     string    `json:"reason,omitempty"`
}

type SiteModerationRequest struct {
	SiteID uuid.UUID `json:"site_id"`
	Action string    `json:"action"` // "approve" or "reject"
	Reason string    `json:"reason,omitempty"`
}

type CampaignModerationRequest struct {
	CampaignID uuid.UUID `json:"campaign_id"`
	Action     string    `json:"action"` // "approve" or "reject"
	Reason     string    `json:"reason,omitempty"`
}

func (h *AdminHandler) ListPendingCreatives(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	const query = `
		SELECT id, campaign_id, type, status, created_at
		FROM creatives
		WHERE status = 'pending_review'
		ORDER BY created_at ASC
	`

	rows, err := h.db.Pool().Query(r.Context(), query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var creatives []map[string]interface{}
	for rows.Next() {
		var id, campaignID uuid.UUID
		var creativeType, status string
		var createdAt time.Time
		if err := rows.Scan(&id, &campaignID, &creativeType, &status, &createdAt); err != nil {
			continue
		}
		creatives = append(creatives, map[string]interface{}{
			"id":          id,
			"campaign_id": campaignID,
			"type":        creativeType,
			"status":      status,
			"created_at":  createdAt,
		})
	}

	httpx.JSON(w, http.StatusOK, creatives)
}

func (h *AdminHandler) ModerateCreative(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req ModerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		httpx.Error(w, http.StatusBadRequest, "Invalid action")
		return
	}

	newStatus := "approved"
	if req.Action == "reject" {
		newStatus = "rejected"
	}

	if err := h.db.UpdateCreativeStatus(r.Context(), req.CreativeID, newStatus); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update creative")
		return
	}

	// Get the creative to find its campaign ID
	creative, err := h.db.GetCreativeByID(r.Context(), req.CreativeID)
	if err == nil && creative != nil {
		// Update Redis campaign metadata with creative status
		campaignMeta := map[string]string{
			"creative_status": newStatus,
		}
		if err := h.redis.SetCampaignMeta(r.Context(), creative.CampaignID.String(), campaignMeta); err != nil {
			// Log error but don't fail the request
			// Redis update is best-effort
		}
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": newStatus})
}

func (h *AdminHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	const query = `
		SELECT id, email, type, status, created_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := h.db.Pool().Query(r.Context(), query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var accounts []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var email, accountType, status string
		var createdAt time.Time
		if err := rows.Scan(&id, &email, &accountType, &status, &createdAt); err != nil {
			continue
		}
		accounts = append(accounts, map[string]interface{}{
			"id":         id,
			"email":      email,
			"type":       accountType,
			"status":     status,
			"created_at": createdAt,
		})
	}

	httpx.JSON(w, http.StatusOK, accounts)
}

func (h *AdminHandler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Status != "active" && req.Status != "suspended" && req.Status != "banned" {
		httpx.Error(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := h.db.UpdateAccountStatus(r.Context(), id, req.Status); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update account")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": req.Status})
}

func (h *AdminHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Get date range from query params
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	query := `
		SELECT 
			COALESCE(SUM(impressions), 0) as impressions,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(conversions), 0) as conversions,
			COALESCE(SUM(spend_cents), 0) as spend_cents,
			COALESCE(SUM(revenue_cents), 0) as revenue_cents,
			COUNT(DISTINCT advertiser_id) as active_advertisers,
			COUNT(DISTINCT publisher_id) as active_publishers,
			COUNT(DISTINCT campaign_id) as active_campaigns,
			COUNT(DISTINCT zone_id) as active_zones
		FROM stats
		WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	if startDate != "" {
		query += " AND DATE(created_at) >= $" + string(rune(argIndex))
		args = append(args, startDate)
		argIndex++
	}
	if endDate != "" {
		query += " AND DATE(created_at) <= $" + string(rune(argIndex))
		args = append(args, endDate)
		argIndex++
	}

	var analytics struct {
		Impressions        int64 `json:"impressions"`
		Clicks             int64 `json:"clicks"`
		Conversions        int64 `json:"conversions"`
		SpendCents         int64 `json:"spend_cents"`
		RevenueCents       int64 `json:"revenue_cents"`
		ActiveAdvertisers  int   `json:"active_advertisers"`
		ActivePublishers   int   `json:"active_publishers"`
		ActiveCampaigns    int   `json:"active_campaigns"`
		ActiveZones        int   `json:"active_zones"`
	}

	err := h.db.Pool().QueryRow(r.Context(), query, args...).Scan(
		&analytics.Impressions, &analytics.Clicks, &analytics.Conversions,
		&analytics.SpendCents, &analytics.RevenueCents, &analytics.ActiveAdvertisers,
		&analytics.ActivePublishers, &analytics.ActiveCampaigns, &analytics.ActiveZones,
	)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	httpx.JSON(w, http.StatusOK, analytics)
}

func (h *AdminHandler) GetRealtimeDashboard(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	ctx := r.Context()
	
	// Get real-time stats from Redis streams
	// For now, aggregate from last hour
	query := `
		SELECT 
			DATE_TRUNC('hour', created_at) as hour,
			COALESCE(SUM(impressions), 0) as impressions,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(conversions), 0) as conversions,
			COALESCE(SUM(spend_cents), 0) as spend_cents,
			COALESCE(SUM(revenue_cents), 0) as revenue_cents
		FROM stats
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY DATE_TRUNC('hour', created_at)
		ORDER BY hour DESC
		LIMIT 24
	`

	rows, err := h.db.Pool().Query(ctx, query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var hourlyStats []map[string]interface{}
	for rows.Next() {
		var hour time.Time
		var impressions, clicks, conversions, spendCents, revenueCents int64
		if err := rows.Scan(&hour, &impressions, &clicks, &conversions, &spendCents, &revenueCents); err != nil {
			continue
		}
		
		ctr := 0.0
		if impressions > 0 {
			ctr = float64(clicks) / float64(impressions) * 100
		}
		
		conversionRate := 0.0
		if clicks > 0 {
			conversionRate = float64(conversions) / float64(clicks) * 100
		}
		
		hourlyStats = append(hourlyStats, map[string]interface{}{
			"hour":            hour.Format("2006-01-02T15:00:00Z"),
			"impressions":     impressions,
			"clicks":          clicks,
			"conversions":     conversions,
			"spend_cents":     spendCents,
			"revenue_cents":   revenueCents,
			"ctr":             ctr,
			"conversion_rate": conversionRate,
		})
	}

	// Get top performing campaigns
	topCampaignsQuery := `
		SELECT 
			campaign_id,
			COALESCE(SUM(impressions), 0) as impressions,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(spend_cents), 0) as spend_cents
		FROM stats
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY campaign_id
		ORDER BY clicks DESC
		LIMIT 10
	`

	rows, err = h.db.Pool().Query(ctx, topCampaignsQuery)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var topCampaigns []map[string]interface{}
	for rows.Next() {
		var campaignID string
		var impressions, clicks, spendCents int64
		if err := rows.Scan(&campaignID, &impressions, &clicks, &spendCents); err != nil {
			continue
		}
		
		ctr := 0.0
		if impressions > 0 {
			ctr = float64(clicks) / float64(impressions) * 100
		}
		
		topCampaigns = append(topCampaigns, map[string]interface{}{
			"campaign_id":  campaignID,
			"impressions":  impressions,
			"clicks":       clicks,
			"spend_cents":  spendCents,
			"ctr":          ctr,
		})
	}

	// Get top performing zones
	topZonesQuery := `
		SELECT 
			zone_id,
			COALESCE(SUM(impressions), 0) as impressions,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(revenue_cents), 0) as revenue_cents
		FROM stats
		WHERE created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY zone_id
		ORDER BY revenue_cents DESC
		LIMIT 10
	`

	rows, err = h.db.Pool().Query(ctx, topZonesQuery)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var topZones []map[string]interface{}
	for rows.Next() {
		var zoneID string
		var impressions, clicks, revenueCents int64
		if err := rows.Scan(&zoneID, &impressions, &clicks, &revenueCents); err != nil {
			continue
		}
		
		ctr := 0.0
		if impressions > 0 {
			ctr = float64(clicks) / float64(impressions) * 100
		}
		
		topZones = append(topZones, map[string]interface{}{
			"zone_id":       zoneID,
			"impressions":   impressions,
			"clicks":        clicks,
			"revenue_cents": revenueCents,
			"ctr":           ctr,
		})
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"hourly_stats":    hourlyStats,
		"top_campaigns":   topCampaigns,
		"top_zones":       topZones,
	})
}

func (h *AdminHandler) GetCohortAnalysis(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	ctx := r.Context()
	
	// Cohort analysis: group advertisers by signup month and track their spend over time
	query := `
		WITH cohorts AS (
			SELECT 
				DATE_TRUNC('month', created_at) as cohort_month,
				id as advertiser_id
			FROM accounts
			WHERE type = 'advertiser'
		),
		spend_by_cohort AS (
			SELECT 
				c.cohort_month,
				DATE_TRUNC('month', s.created_at) as activity_month,
				COALESCE(SUM(s.spend_cents), 0) as total_spend
			FROM cohorts c
			LEFT JOIN stats s ON s.advertiser_id = c.advertiser_id
			GROUP BY c.cohort_month, DATE_TRUNC('month', s.created_at)
		)
		SELECT 
			cohort_month,
			activity_month,
			EXTRACT(MONTH FROM AGE(activity_month, cohort_month)) as month_number,
			total_spend
		FROM spend_by_cohort
		WHERE activity_month >= cohort_month
		ORDER BY cohort_month DESC, month_number ASC
	`

	rows, err := h.db.Pool().Query(ctx, query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type CohortData struct {
		CohortMonth   string  `json:"cohort_month"`
		ActivityMonth string  `json:"activity_month"`
		MonthNumber   int     `json:"month_number"`
		TotalSpend    int64   `json:"total_spend"`
	}

	var cohortData []CohortData
	for rows.Next() {
		var cohortMonth, activityMonth time.Time
		var monthNumber int
		var totalSpend int64
		if err := rows.Scan(&cohortMonth, &activityMonth, &monthNumber, &totalSpend); err != nil {
			continue
		}
		cohortData = append(cohortData, CohortData{
			CohortMonth:   cohortMonth.Format("2006-01"),
			ActivityMonth: activityMonth.Format("2006-01"),
			MonthNumber:   monthNumber,
			TotalSpend:    totalSpend,
		})
	}

	// Group by cohort for easier frontend consumption
	cohorts := make(map[string][]map[string]interface{})
	for _, data := range cohortData {
		if cohorts[data.CohortMonth] == nil {
			cohorts[data.CohortMonth] = []map[string]interface{}{}
		}
		cohorts[data.CohortMonth] = append(cohorts[data.CohortMonth], map[string]interface{}{
			"month_number": data.MonthNumber,
			"total_spend":  data.TotalSpend,
		})
	}

	httpx.JSON(w, http.StatusOK, cohorts)
}

// MarketplaceHandler handles interconnection between advertisers and publishers
type MarketplaceHandler struct {
	db *postgres.DB
}

func NewMarketplaceHandler(db *postgres.DB) *MarketplaceHandler {
	return &MarketplaceHandler{db: db}
}

// GetAvailableSites - allows advertisers to see all publisher sites for targeting
func (h *MarketplaceHandler) GetAvailableSites(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "advertiser" {
		httpx.Error(w, http.StatusForbidden, "Advertiser access required")
		return
	}

	ctx := r.Context()
	
	query := `
		SELECT s.id, s.name, s.domain, s.category, s.description, s.status,
		       COUNT(DISTINCT z.id) as zone_count,
		       COALESCE(SUM(st.impressions), 0) as total_impressions,
		       COALESCE(SUM(st.clicks), 0) as total_clicks
		FROM sites s
		LEFT JOIN zones z ON z.site_id = s.id
		LEFT JOIN stats st ON st.zone_id = z.id
		WHERE s.status = 'active'
		GROUP BY s.id, s.name, s.domain, s.category, s.description, s.status
		ORDER BY total_impressions DESC
	`

	rows, err := h.db.Pool().Query(ctx, query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type SiteInfo struct {
		ID              string  `json:"id"`
		Name            string  `json:"name"`
		Domain          string  `json:"domain"`
		Category        *string `json:"category"`
		Description     *string `json:"description"`
		Status          string  `json:"status"`
		ZoneCount       int     `json:"zone_count"`
		TotalImpressions int64  `json:"total_impressions"`
		TotalClicks     int64  `json:"total_clicks"`
	}

	var sites []SiteInfo
	for rows.Next() {
		var s SiteInfo
		if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.Category, &s.Description, &s.Status, 
			&s.ZoneCount, &s.TotalImpressions, &s.TotalClicks); err != nil {
			continue
		}
		sites = append(sites, s)
	}

	httpx.JSON(w, http.StatusOK, sites)
}

// GetAvailableZones - allows advertisers to see zones for a specific site
func (h *MarketplaceHandler) GetAvailableZones(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "advertiser" {
		httpx.Error(w, http.StatusForbidden, "Advertiser access required")
		return
	}

	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		httpx.Error(w, http.StatusBadRequest, "site_id parameter required")
		return
	}

	ctx := r.Context()
	
	query := `
		SELECT z.id, z.name, z.site_id, z.size, z.format, z.status,
		       COALESCE(SUM(st.impressions), 0) as total_impressions,
		       COALESCE(SUM(st.clicks), 0) as total_clicks,
		       COALESCE(SUM(st.revenue_cents), 0) as total_revenue
		FROM zones z
		LEFT JOIN stats st ON st.zone_id = z.id
		WHERE z.site_id = $1 AND z.status = 'active'
		GROUP BY z.id, z.name, z.site_id, z.size, z.format, z.status
		ORDER BY total_impressions DESC
	`

	rows, err := h.db.Pool().Query(ctx, query, siteID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type ZoneInfo struct {
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		SiteID         string  `json:"site_id"`
		Size           string  `json:"size"`
		Format         string  `json:"format"`
		Status         string  `json:"status"`
		TotalImpressions int64 `json:"total_impressions"`
		TotalClicks    int64  `json:"total_clicks"`
		TotalRevenue   int64  `json:"total_revenue"`
	}

	var zones []ZoneInfo
	for rows.Next() {
		var z ZoneInfo
		if err := rows.Scan(&z.ID, &z.Name, &z.SiteID, &z.Size, &z.Format, &z.Status,
			&z.TotalImpressions, &z.TotalClicks, &z.TotalRevenue); err != nil {
			continue
		}
		zones = append(zones, z)
	}

	httpx.JSON(w, http.StatusOK, zones)
}

// GetActiveCampaigns - allows publishers to see active advertiser campaigns
func (h *MarketplaceHandler) GetActiveCampaigns(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "publisher" {
		httpx.Error(w, http.StatusForbidden, "Publisher access required")
		return
	}

	ctx := r.Context()
	
	query := `
		SELECT c.id, c.name, c.advertiser_id, a.company_name, c.status, c.budget_cents,
		       c.daily_budget_cents, c.start_date, c.end_date,
		       COALESCE(SUM(st.spend_cents), 0) as total_spend,
		       COALESCE(SUM(st.impressions), 0) as total_impressions
		FROM campaigns c
		JOIN accounts a ON a.id = c.advertiser_id
		LEFT JOIN stats st ON st.campaign_id = c.id
		WHERE c.status = 'active' AND (c.end_date IS NULL OR c.end_date > NOW())
		GROUP BY c.id, c.name, c.advertiser_id, a.company_name, c.status, c.budget_cents,
		         c.daily_budget_cents, c.start_date, c.end_date
		ORDER BY total_spend DESC
		LIMIT 100
	`

	rows, err := h.db.Pool().Query(ctx, query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type CampaignInfo struct {
		ID             string     `json:"id"`
		Name           string     `json:"name"`
		AdvertiserID   string     `json:"advertiser_id"`
		CompanyName    *string    `json:"company_name"`
		Status         string     `json:"status"`
		BudgetCents    int64      `json:"budget_cents"`
		DailyBudgetCents int64    `json:"daily_budget_cents"`
		StartDate      *time.Time `json:"start_date"`
		EndDate        *time.Time `json:"end_date"`
		TotalSpend     int64      `json:"total_spend"`
		TotalImpressions int64    `json:"total_impressions"`
	}

	var campaigns []CampaignInfo
	for rows.Next() {
		var c CampaignInfo
		if err := rows.Scan(&c.ID, &c.Name, &c.AdvertiserID, &c.CompanyName, &c.Status,
			&c.BudgetCents, &c.DailyBudgetCents, &c.StartDate, &c.EndDate,
			&c.TotalSpend, &c.TotalImpressions); err != nil {
			continue
		}
		campaigns = append(campaigns, c)
	}

	httpx.JSON(w, http.StatusOK, campaigns)
}

type PlatformStats struct {
	TotalUsers       int64 `json:"total_users"`
	TotalAdvertisers int64 `json:"total_advertisers"`
	TotalPublishers  int64 `json:"total_publishers"`
	ActiveCampaigns  int64 `json:"active_campaigns"`
	TotalCampaigns   int64 `json:"total_campaigns"`
	TotalRevenueCents int64 `json:"total_revenue_cents"`
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	ctx := r.Context()

	var stats PlatformStats

	h.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM accounts WHERE type = 'advertiser'").Scan(&stats.TotalAdvertisers)
	h.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM accounts WHERE type = 'publisher'").Scan(&stats.TotalPublishers)
	stats.TotalUsers = stats.TotalAdvertisers + stats.TotalPublishers

	h.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM campaigns WHERE status = 'active'").Scan(&stats.ActiveCampaigns)
	h.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM campaigns").Scan(&stats.TotalCampaigns)

	h.db.Pool().QueryRow(ctx, "SELECT COALESCE(SUM(spend_cents), 0) FROM stats").Scan(&stats.TotalRevenueCents)

	httpx.JSON(w, http.StatusOK, stats)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	ctx := r.Context()

	query := `
		SELECT id, type, email, company_name, status, created_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := h.db.Pool().Query(ctx, query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var users []postgres.Account
	for rows.Next() {
		var user postgres.Account
		if err := rows.Scan(&user.ID, &user.Type, &user.Email, &user.CompanyName, &user.Status, &user.CreatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	httpx.JSON(w, http.StatusOK, users)
}

func (h *AdminHandler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	ctx := r.Context()

	query := `
		SELECT id, advertiser_id, name, status, pricing_model, bid_amount_cents, 
		       daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at, created_at
		FROM campaigns
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := h.db.Pool().Query(ctx, query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var campaigns []postgres.Campaign
	for rows.Next() {
		var camp postgres.Campaign
		if err := rows.Scan(&camp.ID, &camp.AdvertiserID, &camp.Name, &camp.Status, &camp.PricingModel,
			&camp.BidAmountCents, &camp.DailyBudgetCents, &camp.TotalBudgetCents, &camp.Timezone,
			&camp.StartsAt, &camp.EndsAt, &camp.CreatedAt); err != nil {
			continue
		}
		campaigns = append(campaigns, camp)
	}

	httpx.JSON(w, http.StatusOK, campaigns)
}

func (h *AdminHandler) ListPendingSites(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	const query = `
		SELECT id, publisher_id, domain, status, created_at
		FROM sites
		WHERE status = 'pending_review'
		ORDER BY created_at ASC
	`

	rows, err := h.db.Pool().Query(r.Context(), query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var sites []map[string]interface{}
	for rows.Next() {
		var id, publisherID uuid.UUID
		var domain, status string
		var createdAt time.Time
		if err := rows.Scan(&id, &publisherID, &domain, &status, &createdAt); err != nil {
			continue
		}
		sites = append(sites, map[string]interface{}{
			"id":           id,
			"publisher_id": publisherID,
			"domain":       domain,
			"status":       status,
			"created_at":   createdAt,
		})
	}

	httpx.JSON(w, http.StatusOK, sites)
}

func (h *AdminHandler) ModerateSite(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req SiteModerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		httpx.Error(w, http.StatusBadRequest, "Invalid action")
		return
	}

	newStatus := "active"
	if req.Action == "reject" {
		newStatus = "rejected"
	}

	const query = `UPDATE sites SET status = $1 WHERE id = $2`
	if _, err := h.db.Pool().Exec(r.Context(), query, newStatus, req.SiteID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update site")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": newStatus})
}

func (h *AdminHandler) ListPendingCampaigns(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	const query = `
		SELECT id, advertiser_id, name, status, pricing_model, bid_amount_cents, 
		       daily_budget_cents, total_budget_cents, created_at
		FROM campaigns
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`

	rows, err := h.db.Pool().Query(r.Context(), query)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var campaigns []map[string]interface{}
	for rows.Next() {
		var id, advertiserID uuid.UUID
		var name, status, pricingModel string
		var bidAmountCents, dailyBudgetCents, totalBudgetCents int
		var createdAt time.Time
		if err := rows.Scan(&id, &advertiserID, &name, &status, &pricingModel, &bidAmountCents,
			&dailyBudgetCents, &totalBudgetCents, &createdAt); err != nil {
			continue
		}
		campaigns = append(campaigns, map[string]interface{}{
			"id":                 id,
			"advertiser_id":      advertiserID,
			"name":               name,
			"status":             status,
			"pricing_model":      pricingModel,
			"bid_amount_cents":   bidAmountCents,
			"daily_budget_cents": dailyBudgetCents,
			"total_budget_cents": totalBudgetCents,
			"created_at":         createdAt,
		})
	}

	httpx.JSON(w, http.StatusOK, campaigns)
}

func (h *AdminHandler) ModerateCampaign(w http.ResponseWriter, r *http.Request) {
	accountType, ok := mw.AccountTypeFromContext(r.Context())
	if !ok || accountType != "admin" {
		httpx.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req CampaignModerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		httpx.Error(w, http.StatusBadRequest, "Invalid action")
		return
	}

	newStatus := "active"
	if req.Action == "reject" {
		newStatus = "rejected"
	}

	const query = `UPDATE campaigns SET status = $1 WHERE id = $2`
	if _, err := h.db.Pool().Exec(r.Context(), query, newStatus, req.CampaignID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	// Update Redis campaign metadata with new status
	campaignMeta := map[string]string{
		"status": newStatus,
	}
	if err := h.redis.SetCampaignMeta(r.Context(), req.CampaignID.String(), campaignMeta); err != nil {
		// Log error but don't fail the request
		// Redis update is best-effort
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": newStatus})
}

func (h *AdminHandler) UpdateCampaignStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID, err := uuid.Parse(vars["id"])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Status != "active" && req.Status != "paused" && req.Status != "draft" {
		httpx.Error(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := h.db.UpdateCampaignStatus(r.Context(), campaignID, req.Status); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	// Update Redis campaign metadata with new status
	campaignMeta := map[string]string{
		"status": req.Status,
	}
	if err := h.redis.SetCampaignMeta(r.Context(), campaignID.String(), campaignMeta); err != nil {
		// Log error but don't fail the request
		// Redis update is best-effort
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": req.Status})
}

