package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"adnet/internal/fraud"
	"adnet/internal/store/redis"
	"adnet/pkg/httpx"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Config struct {
	RedisURL  string
	Port      string
	JWTSecret string
}

func main() {
	config := Config{
		RedisURL:  getEnv("REDIS_URL", "localhost:6379"),
		Port:      getEnv("PORT", "8081"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
	}

	redisClient, err := redis.NewClient(config.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	adserve := &AdserveHandler{
		redis:               redisClient,
		jwtSecret:           config.JWTSecret,
		fraudDetector:       fraud.NewFraudDetector(),
		rateLimiter:         fraud.NewRateLimiter(10, 60), // 10 requests per minute per IP
		clickValidator:      fraud.NewClickValidator(),
		ipReputationChecker: fraud.NewIPReputationChecker(),
	}

	// Start pub/sub subscriber for campaign updates
	go adserve.subscribeToCampaignUpdates(context.Background())

	r := mux.NewRouter()
	r.HandleFunc("/serve", adserve.ServeAd).Methods("GET")
	r.HandleFunc("/click", adserve.HandleClick).Methods("GET")
	r.HandleFunc("/tag.js", adserve.ServeTagJS).Methods("GET")
	r.HandleFunc("/healthz", adserve.Health).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + config.Port,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}

	log.Printf("Adserve service starting on port %s", config.Port)
	log.Fatal(srv.ListenAndServe())
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type AdserveHandler struct {
	redis               *redis.Client
	jwtSecret           string
	fraudDetector       *fraud.FraudDetector
	rateLimiter         *fraud.RateLimiter
	clickValidator      *fraud.ClickValidator
	ipReputationChecker *fraud.IPReputationChecker
}

type ServeRequest struct {
	ZoneID  string `json:"zone_id"`
	Format  string `json:"format"`
	IP      string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

type AdResponse struct {
	CampaignID  string                 `json:"campaign_id"`
	CreativeID  string                 `json:"creative_id"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	ImageURL    string                 `json:"image_url"`
	ClickURL    string                 `json:"click_url"`
	ClickToken  string                 `json:"click_token"`
	Format      string                 `json:"format"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ClickToken struct {
	CampaignID string    `json:"campaign_id"`
	ZoneID     string    `json:"zone_id"`
	Timestamp  time.Time `json:"timestamp"`
	Nonce      string    `json:"nonce"`
}

func (h *AdserveHandler) ServeAd(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone")
	format := r.URL.Query().Get("format")
	ip := getClientIP(r)

	if zoneID == "" {
		httpx.Error(w, http.StatusBadRequest, "zone parameter is required")
		return
	}

	ctx := r.Context()

	// Get zone metadata
	zoneMeta, err := h.redis.GetZoneMeta(ctx, zoneID)
	if err != nil || len(zoneMeta) == 0 {
		httpx.Error(w, http.StatusNotFound, "Zone not found")
		return
	}

	// The ad tag (tag.js) only sends the zone ID. A zone has a fixed ad
	// format, so fall back to the zone's configured format for matching.
	if format == "" {
		format = zoneMeta["format"]
	}

	// Get campaign candidates sorted by eCPM
	candidates, err := h.redis.GetCampaignCandidates(ctx, zoneID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to get candidates")
		return
	}

	if len(candidates) == 0 {
		httpx.Error(w, http.StatusNoContent, "No ads available")
		return
	}

	// Select winning campaign using multi-armed bandit algorithm
	var winner *AdResponse
	var winningCampaignMeta map[string]string
	var bestScore float64
	
	// Get zone performance data for optimization
	zoneCTR, _ := h.redis.GetZoneCTR(ctx, zoneID)
	zoneConversionRate, _ := h.redis.GetZoneConversionRate(ctx, zoneID)
	
	for _, campaignID := range candidates {
		// Check budget
		spendToday, _ := h.redis.GetCampaignSpendToday(ctx, campaignID)
		dailyBudget, _ := h.redis.GetDailyBudget(ctx, campaignID)
		if spendToday >= dailyBudget {
			continue
		}

		// Get campaign metadata
		campaignMeta, err := h.redis.GetCampaignMeta(ctx, campaignID)
		if err != nil || len(campaignMeta) == 0 {
			continue
		}

		// Check campaign approval status - only serve active (approved) campaigns
		if campaignMeta["status"] != "active" {
			continue
		}

		// Check creative approval status - only serve approved creatives
		if campaignMeta["creative_status"] != "approved" {
			continue
		}

		// Check format match
		if campaignMeta["format"] != format {
			continue
		}

		// Frequency cap check
		userHash := hashUser(ip, r.UserAgent())
		freqCap, _ := strconv.Atoi(campaignMeta["frequency_cap"])
		if freqCap > 0 {
			freqCount, _ := h.redis.GetFrequencyCount(ctx, userHash, campaignID)
			if freqCount >= int64(freqCap) {
				continue
			}
			// Increment frequency counter with 24h TTL
			h.redis.IncrementFrequency(ctx, userHash, campaignID, 24*time.Hour)
		}

		// Enhanced fraud checks
		// Check if IP is blocked in Redis
		blocked, _ := h.redis.IsIPBlocked(ctx, ip)
		if blocked {
			continue
		}

		// Check if IP is in datacenter range
		if h.fraudDetector.IsDatacenterIP(ip) {
			continue
		}

		// Check user agent for suspicious patterns
		if h.fraudDetector.CheckUserAgent(r.UserAgent()) {
			continue
		}

		// Rate limit check using in-memory rate limiter
		now := time.Now().Unix()
		if !h.rateLimiter.CheckRate(ip, now) {
			continue
		}

		// Also track in Redis for distributed rate limiting
		h.redis.IncrementRate(ctx, ip, time.Minute)

		// Calculate optimization score using multi-armed bandit
		// Score = (bid * predicted_ctr) + exploration_bonus
		bid, _ := strconv.ParseFloat(campaignMeta["bid"], 64)
		campaignCTR, _ := h.redis.GetCampaignCTR(ctx, campaignID)
		
		// Thompson sampling for exploration-exploitation
		alpha := 1.0 + float64(campaignCTR*1000) // successes
		beta := 1.0 + float64(1000-campaignCTR*1000) // failures
		sample := h.betaSample(alpha, beta)
		
		score := bid * sample
		
		// Apply zone performance modifiers
		if zoneCTR > 0.01 {
			score *= (1.0 + zoneCTR)
		}
		if zoneConversionRate > 0.001 {
			score *= (1.0 + zoneConversionRate*10)
		}
		
		// Check if this is the best candidate
		if score > bestScore {
			bestScore = score
			
			// Generate click token
			nonce := uuid.New().String()
			clickToken, err := h.generateClickToken(campaignID, zoneID, nonce)
			if err != nil {
				continue
			}

			winningCampaignMeta = campaignMeta
			
			// Build metadata based on format
			metadata := make(map[string]interface{})
			switch format {
			case "native":
				metadata["icon_url"] = campaignMeta["icon_url"]
				metadata["cta_text"] = campaignMeta["cta_text"]
				metadata["sponsored_by"] = campaignMeta["sponsored_by"]
			case "popunder", "interstitial":
				metadata["width"] = campaignMeta["width"]
				metadata["height"] = campaignMeta["height"]
			case "push":
				metadata["push_title"] = campaignMeta["push_title"]
				metadata["push_body"] = campaignMeta["push_body"]
				metadata["push_icon"] = campaignMeta["push_icon"]
			}
			
			winner = &AdResponse{
				CampaignID: campaignID,
				CreativeID: campaignMeta["creative_id"],
				Title:      campaignMeta["title"],
				Body:       campaignMeta["body"],
				ImageURL:   campaignMeta["image_url"],
				ClickURL:   campaignMeta["click_url"],
				ClickToken: clickToken,
				Format:     campaignMeta["format"],
				Metadata:   metadata,
			}
		}
	}

	if winner == nil {
		httpx.Error(w, http.StatusNoContent, "No matching ads")
		return
	}

	// Record impression to Redis Stream
	event := map[string]interface{}{
		"type":         "impression",
		"campaign_id":  winner.CampaignID,
		"zone_id":      zoneID,
		"creative_id":  winner.CreativeID,
		"user_hash":    hashUser(ip, r.UserAgent()),
		"country":      "KE", // TODO: implement geo lookup
		"device_type":  "desktop", // TODO: parse UA
		"cost_cents":   winningCampaignMeta["bid_cents"],
		"ip":           ip,
		"user_agent":   r.UserAgent(),
		"timestamp":    time.Now().Unix(),
	}
	h.redis.AddEvent(ctx, "events:impressions", event)

	// Increment spend counter in Redis
	bidCents, _ := strconv.ParseInt(winningCampaignMeta["bid_cents"], 10, 64)
	h.redis.IncrementCampaignSpend(ctx, winner.CampaignID, bidCents)

	httpx.JSON(w, http.StatusOK, winner)
}

func (h *AdserveHandler) HandleClick(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		httpx.Error(w, http.StatusBadRequest, "token parameter is required")
		return
	}

	ip := getClientIP(r)
	userHash := hashUser(ip, r.UserAgent())
	
	// Check IP reputation
	if !h.ipReputationChecker.CheckIPReputation(ip) {
		httpx.Error(w, http.StatusForbidden, "IP blocked due to suspicious reputation")
		return
	}

	// Verify click token
	clickToken, err := h.verifyClickToken(token)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Check token expiry (5 minutes)
	if time.Since(clickToken.Timestamp) > 5*time.Minute {
		httpx.Error(w, http.StatusUnauthorized, "Token expired")
		return
	}

	// Validate click timing
	now := time.Now().Unix()
	if !h.clickValidator.ValidateClick(userHash, clickToken.Timestamp.Unix(), now) {
		httpx.Error(w, http.StatusForbidden, "Suspicious click pattern detected")
		return
	}

	// Get campaign metadata to get click URL
	ctx := r.Context()
	campaignMeta, err := h.redis.GetCampaignMeta(ctx, clickToken.CampaignID)
	if err != nil || len(campaignMeta) == 0 {
		httpx.Error(w, http.StatusNotFound, "Campaign not found")
		return
	}

	// Record click to Redis Stream
	event := map[string]interface{}{
		"type":         "click",
		"campaign_id":  clickToken.CampaignID,
		"zone_id":      clickToken.ZoneID,
		"creative_id":  campaignMeta["creative_id"],
		"user_hash":    userHash,
		"cost_cents":   campaignMeta["bid_cents"],
		"ip":           ip,
		"user_agent":   r.UserAgent(),
		"timestamp":    now,
		"click_token":  token,
	}
	h.redis.AddEvent(ctx, "events:clicks", event)

	clickURL := campaignMeta["click_url"]

	// Redirect to advertiser URL
	http.Redirect(w, r, clickURL, http.StatusFound)
}

func (h *AdserveHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *AdserveHandler) ServeTagJS(w http.ResponseWriter, r *http.Request) {
	// Serve the tag.js file with aggressive caching
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600, immutable") // Cache for 1 hour
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Read the tag.js file
	tagJS, err := os.ReadFile("tag.js")
	if err != nil {
		http.Error(w, "Tag file not found", http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write(tagJS)
}

// betaSample generates a sample from Beta distribution using rejection sampling
// Used for Thompson sampling in multi-armed bandit optimization
func (h *AdserveHandler) betaSample(alpha, beta float64) float64 {
	// Use gamma distribution approximation: Beta(alpha, beta) = Gamma(alpha) / (Gamma(alpha) + Gamma(beta))
	g1 := h.gammaSample(alpha)
	g2 := h.gammaSample(beta)
	return g1 / (g1 + g2)
}

// gammaSample generates a sample from Gamma distribution using Marsaglia and Tsang's method
func (h *AdserveHandler) gammaSample(alpha float64) float64 {
	if alpha < 1 {
		// Use the transformation: Gamma(alpha) = Gamma(alpha+1) * U^(1/alpha)
		return h.gammaSample(alpha+1) * math.Pow(rand.Float64(), 1/alpha)
	}
	
	// Marsaglia and Tsang's method for alpha >= 1
	d := alpha - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)
	
	for {
		x := rand.NormFloat64()
		v := 1.0 + c*x
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := rand.Float64()
		if u < 1.0-0.0331*(x*x)*(x*x) {
			return d * v
		}
		if math.Log(u) < 0.5*x*x + d*(1.0-v+math.Log(v)) {
			return d * v
		}
	}
}

func (h *AdserveHandler) generateClickToken(campaignID, zoneID, nonce string) (string, error) {
	token := ClickToken{
		CampaignID: campaignID,
		ZoneID:     zoneID,
		Timestamp:  time.Now(),
		Nonce:      nonce,
	}

	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	sig := hmac.New(sha256.New, []byte(h.jwtSecret))
	sig.Write(data)
	signature := hex.EncodeToString(sig.Sum(nil))

	tokenStr := base64.URLEncoding.EncodeToString(data)
	return fmt.Sprintf("%s.%s", tokenStr, signature), nil
}

func (h *AdserveHandler) verifyClickToken(tokenStr string) (*ClickToken, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	sig := hmac.New(sha256.New, []byte(h.jwtSecret))
	sig.Write(data)
	expectedSignature := hex.EncodeToString(sig.Sum(nil))

	if parts[1] != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	var token ClickToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take first IP if multiple
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func hashUser(ip, userAgent string) string {
	h := sha512.New512_256()
	h.Write([]byte(ip + userAgent + "adnet-salt"))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (h *AdserveHandler) subscribeToCampaignUpdates(ctx context.Context) {
	pubsub := h.redis.SubscribeToCampaignUpdates(ctx)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			campaignID := msg.Payload
			log.Printf("Campaign update received: %s", campaignID)
			// In a full implementation, this would refresh the campaign's Redis keys
			// For now, we just log the update
		}
	}
}
