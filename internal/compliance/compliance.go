package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"adnet/internal/store/redis"
)

type ComplianceEngine struct {
	redis    *redis.Client
	gdpr     *GDPRManager
	ccpa     *CCPAManager
	adVerify *AdVerification
}

type GDPRManager struct {
	redis          *redis.Client
	consentTTL     time.Duration
	vendorList     map[string]bool
	purposeList    map[string]bool
	specialFeatures map[string]bool
}

type CCPAManager struct {
	redis *redis.Client
	optOutTTL time.Duration
}

type AdVerification struct {
	redis      *redis.Client
	blocklist  map[string]bool
	adsTxt    map[string]string
}

type ConsentRecord struct {
	UserID       string                 `json:"user_id"`
	ConsentGiven bool                   `json:"consent_given"`
	Timestamp    time.Time              `json:"timestamp"`
	VendorConsents map[string]bool      `json:"vendor_consents"`
	PurposeConsents map[string]bool     `json:"purpose_consents"`
	SpecialFeatures map[string]bool     `json:"special_features"`
	TCFString    string                 `json:"tcf_string"`
	GPPString    string                 `json:"gpp_string"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
}

type CCPARecord struct {
	UserID       string    `json:"user_id"`
	OptOutSale   bool      `json:"opt_out_sale"`
	OptOutTargeting bool   `json:"opt_out_targeting"`
	Timestamp    time.Time `json:"timestamp"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}

type AdVerificationRequest struct {
	AdID          string   `json:"ad_id"`
	CampaignID    string   `json:"campaign_id"`
	CreativeID    string   `json:"creative_id"`
	AdvertiserID  string   `json:"advertiser_id"`
	ImpressionID  string   `json:"impression_id"`
	VerificationData map[string]interface{} `json:"verification_data"`
}

type AdVerificationResult struct {
	Verified      bool      `json:"verified"`
	Reason        string    `json:"reason"`
	ScanTime      time.Time `json:"scan_time"`
	Blocked       bool      `json:"blocked"`
	Category      string    `json:"category,omitempty"`
	Confidence    float64   `json:"confidence"`
}

func NewComplianceEngine(redisClient *redis.Client) *ComplianceEngine {
	return &ComplianceEngine{
		redis: redisClient,
		gdpr: &GDPRManager{
			redis:          redisClient,
			consentTTL:     365 * 24 * time.Hour, // 1 year
			vendorList:     make(map[string]bool),
			purposeList:    make(map[string]bool),
			specialFeatures: make(map[string]bool),
		},
		ccpa: &CCPAManager{
			redis:    redisClient,
			optOutTTL: 180 * 24 * time.Hour, // 6 months
		},
		adVerify: &AdVerification{
			redis:     redisClient,
			blocklist: make(map[string]bool),
			adsTxt:   make(map[string]string),
		},
	}
}

// GDPR Compliance Methods

// CheckConsent checks if a user has given consent for specific vendor/purpose
func (e *ComplianceEngine) CheckConsent(ctx context.Context, userHash string, vendorID, purposeID string) (bool, error) {
	consent, err := e.getUserConsent(ctx, userHash)
	if err != nil {
		return false, err
	}

	if !consent.ConsentGiven {
		return false, nil
	}

	// Check vendor consent
	if vendorID != "" {
		if vendorConsent, ok := consent.VendorConsents[vendorID]; !ok || !vendorConsent {
			return false, nil
		}
	}

	// Check purpose consent
	if purposeID != "" {
		if purposeConsent, ok := consent.PurposeConsents[purposeID]; !ok || !purposeConsent {
			return false, nil
		}
	}

	return true, nil
}

// RecordConsent records a user's consent decision
func (e *ComplianceEngine) RecordConsent(ctx context.Context, consent ConsentRecord) error {
	consent.Timestamp = time.Now()
	
	data, err := json.Marshal(consent)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("consent:gdpr:%s", consent.UserID)
	return e.redis.Set(ctx, key, string(data), e.gdpr.consentTTL)
}

// GenerateTCFString generates a Transparency & Consent Framework (TCF) string
func (e *ComplianceEngine) GenerateTCFString(consent ConsentRecord) (string, error) {
	// Simplified TCF string generation
	// In production, this would follow the IAB Europe TCF specification
	
	version := "2"
	consentScreen := "1"
	purposeConsent := ""
	vendorConsent := ""
	
	// Build purpose consent bits
	for i := 1; i <= 10; i++ {
		purposeID := fmt.Sprintf("P%d", i)
		if consent.PurposeConsents[purposeID] {
			purposeConsent += "1"
		} else {
			purposeConsent += "0"
		}
	}
	
	// Build vendor consent bits
	vendorIDs := make([]string, 0, len(consent.VendorConsents))
	for vendorID := range consent.VendorConsents {
		vendorIDs = append(vendorIDs, vendorID)
	}
	
	// Simplified encoding
	tcfString := fmt.Sprintf("%s.%s.%s.%s", version, consentScreen, purposeConsent, vendorConsent)
	
	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString([]byte(tcfString))
	return encoded, nil
}

// ParseTCFString parses a TCF string to extract consent information
func (e *ComplianceEngine) ParseTCFString(tcfString string) (map[string]bool, error) {
	// Simplified TCF string parsing
	// In production, this would follow the IAB Europe TCF specification
	
	decoded, err := base64.StdEncoding.DecodeString(tcfString)
	if err != nil {
		return nil, err
	}
	
	parts := strings.Split(string(decoded), ".")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid TCF string format")
	}
	
	purposeConsent := parts[2]
	consents := make(map[string]bool)
	
	for i, bit := range purposeConsent {
		if bit == '1' {
			consents[fmt.Sprintf("P%d", i+1)] = true
		}
	}
	
	return consents, nil
}

// RegisterVendor registers a vendor for GDPR compliance
func (e *ComplianceEngine) RegisterVendor(vendorID string, enabled bool) {
	e.gdpr.vendorList[vendorID] = enabled
}

// RegisterPurpose registers a purpose for GDPR compliance
func (e *ComplianceEngine) RegisterPurpose(purposeID string, enabled bool) {
	e.gdpr.purposeList[purposeID] = enabled
}

// RegisterSpecialFeature registers a special feature for GDPR compliance
func (e *ComplianceEngine) RegisterSpecialFeature(featureID string, enabled bool) {
	e.gdpr.specialFeatures[featureID] = enabled
}

// CCPA Compliance Methods

// CheckCCPAOptOut checks if a user has opted out of data sale
func (e *ComplianceEngine) CheckCCPAOptOut(ctx context.Context, userHash string) (bool, error) {
	record, err := e.getCCPARecord(ctx, userHash)
	if err != nil {
		return false, nil // Default to not opted out
	}

	return record.OptOutSale, nil
}

// RecordCCPAOptOut records a user's CCPA opt-out decision
func (e *ComplianceEngine) RecordCCPAOptOut(ctx context.Context, record CCPARecord) error {
	record.Timestamp = time.Now()
	
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("consent:ccpa:%s", record.UserID)
	return e.redis.Set(ctx, key, string(data), e.ccpa.optOutTTL)
}

// IsCCPAApplicable checks if CCPA applies based on user location
func (e *ComplianceEngine) IsCCPAApplicable(ipAddress string) bool {
	// Check if IP is from California, USA
	// In production, this would use a proper geo-IP database
	
	// Simplified check - in production use MaxMind or similar
	return e.isUSIP(ipAddress)
}

// Ad Verification Methods

// VerifyAd verifies an ad for compliance and safety
func (e *ComplianceEngine) VerifyAd(ctx context.Context, request AdVerificationRequest) (*AdVerificationResult, error) {
	result := &AdVerificationResult{
		Verified: true,
		Reason:   "Ad passed verification",
		ScanTime: time.Now(),
		Blocked:  false,
		Confidence: 1.0,
	}

	// Check blocklist
	if e.adVerify.blocklist[request.AdvertiserID] {
		result.Verified = false
		result.Blocked = true
		result.Reason = "Advertiser is blocked"
		result.Category = "blocked_advertiser"
		result.Confidence = 1.0
		return result, nil
	}

	if e.adVerify.blocklist[request.CampaignID] {
		result.Verified = false
		result.Blocked = true
		result.Reason = "Campaign is blocked"
		result.Category = "blocked_campaign"
		result.Confidence = 1.0
		return result, nil
	}

	// Check ads.txt compliance
	if !e.checkAdsTxtCompliance(request.AdvertiserID) {
		result.Verified = false
		result.Reason = "ads.txt verification failed"
		result.Category = "ads_txt_violation"
		result.Confidence = 0.8
		return result, nil
	}

	// Additional verification checks could include:
	// - Brand safety
	// - Content classification
	// - Malware scanning
	// - Landing page verification

	return result, nil
}

// BlockAdvertiser adds an advertiser to the blocklist
func (e *ComplianceEngine) BlockAdvertiser(advertiserID string, reason string) {
	e.adVerify.blocklist[advertiserID] = true
	
	// Log to Redis for persistence
	ctx := context.Background()
	e.redis.Set(ctx, fmt.Sprintf("blocklist:advertiser:%s", advertiserID), reason, 0)
}

// UnblockAdvertiser removes an advertiser from the blocklist
func (e *ComplianceEngine) UnblockAdvertiser(advertiserID string) {
	delete(e.adVerify.blocklist, advertiserID)
	
	ctx := context.Background()
	e.redis.Del(ctx, fmt.Sprintf("blocklist:advertiser:%s", advertiserID))
}

// UpdateAdsTxt updates the ads.txt records
func (e *ComplianceEngine) UpdateAdsTxt(domain, publisherID string) {
	e.adVerify.adsTxt[domain] = publisherID
	
	ctx := context.Background()
	e.redis.Set(ctx, fmt.Sprintf("adstxt:%s", domain), publisherID, 0)
}

// Helper methods

func (e *ComplianceEngine) getUserConsent(ctx context.Context, userHash string) (*ConsentRecord, error) {
	key := fmt.Sprintf("consent:gdpr:%s", userHash)
	data, err := e.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var consent ConsentRecord
	if err := json.Unmarshal([]byte(data), &consent); err != nil {
		return nil, err
	}

	return &consent, nil
}

func (e *ComplianceEngine) getCCPARecord(ctx context.Context, userHash string) (*CCPARecord, error) {
	key := fmt.Sprintf("consent:ccpa:%s", userHash)
	data, err := e.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var record CCPARecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (e *ComplianceEngine) isUSIP(ipAddress string) bool {
	// Simplified US IP check
	// In production, use a proper geo-IP database
	
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	// Check if IP is in US ranges (simplified)
	// This is a placeholder - use MaxMind GeoIP2 in production
	return false
}

func (e *ComplianceEngine) checkAdsTxtCompliance(advertiserID string) bool {
	// Simplified ads.txt check
	// In production, this would fetch and parse the publisher's ads.txt file
	
	// For now, assume all advertisers are compliant
	return true
}

// HashUserData hashes user data for privacy compliance
func HashUserData(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// AnonymizeIP anonymizes an IP address by removing the last octet
func AnonymizeIP(ipAddress string) string {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return ipAddress
	}

	ip = ip.Mask(net.CIPMask(255, 255, 255, 0))
	return ip.String()
}

// GetConsentManagementHTML returns HTML for consent management UI
func (e *ComplianceEngine) GetConsentManagementHTML() string {
	return `
<div id="consent-manager" class="consent-manager">
	<h3>Privacy Settings</h3>
	<p>We use cookies and similar technologies to help personalize content and measure performance.</p>
	
	<div class="consent-category">
		<h4>Essential</h4>
		<p>Required for the site to function.</p>
		<input type="checkbox" checked disabled>
	</div>
	
	<div class="consent-category">
		<h4>Analytics</h4>
		<p>Help us improve our website.</p>
		<input type="checkbox" id="consent-analytics">
	</div>
	
	<div class="consent-category">
		<h4>Advertising</h4>
		<p>Help us show relevant ads.</p>
		<input type="checkbox" id="consent-advertising">
	</div>
	
	<button id="save-consent">Save Preferences</button>
	<button id="accept-all">Accept All</button>
	<button id="reject-all">Reject All</button>
</div>
`
}

// GetCCPAOptOutHTML returns HTML for CCPA opt-out UI
func (e *ComplianceEngine) GetCCPAOptOutHTML() string {
	return `
<div id="ccpa-opt-out" class="ccpa-opt-out">
	<h3>Do Not Sell My Personal Information</h3>
	<p>We respect your privacy and are committed to protecting your personal data.</p>
	
	<label>
		<input type="checkbox" id="ccpa-opt-out-checkbox">
		Do not sell my personal information
	</label>
	
	<button id="save-ccpa-preference">Save Preference</button>
</div>
`
}
