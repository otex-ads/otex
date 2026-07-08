package onboarding

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"adnet/internal/store/redis"
)

type OnboardingEngine struct {
	redis *redis.Client
	kyc   *KYCManager
}

type KYCManager struct {
	redis      *redis.Client
	providers  map[string]KYCProvider
}

type KYCProvider interface {
	SubmitVerification(ctx context.Context, request KYCRequest) (*KYCResponse, error)
	CheckStatus(ctx context.Context, verificationID string) (*KYCStatus, error)
}

type KYCRequest struct {
	UserID       string                 `json:"user_id"`
	AccountType  string                 `json:"account_type"` // "advertiser" or "publisher"
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone"`
	DateOfBirth  string                 `json:"date_of_birth"`
	Address      Address                `json:"address"`
	DocumentType string                 `json:"document_type"` // "passport", "id_card", "driving_license"
	DocumentData DocumentData           `json:"document_data"`
	BusinessInfo *BusinessInfo          `json:"business_info,omitempty"`
	TaxInfo      *TaxInfo               `json:"tax_info,omitempty"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type DocumentData struct {
	FrontImage  string `json:"front_image"`  // Base64 encoded
	BackImage   string `json:"back_image"`   // Base64 encoded
	SelfieImage string `json:"selfie_image"` // Base64 encoded
	DocumentNumber string `json:"document_number"`
	ExpiryDate string `json:"expiry_date"`
}

type BusinessInfo struct {
	CompanyName   string   `json:"company_name"`
	RegistrationNumber string `json:"registration_number"`
	TaxID        string   `json:"tax_id"`
	Website      string   `json:"website"`
	Industry     string   `json:"industry"`
	BusinessType string   `json:"business_type"`
}

type TaxInfo struct {
	TaxID        string `json:"tax_id"`
	TaxCountry   string `json:"tax_country"`
	VATNumber    string `json:"vat_number"`
}

type KYCResponse struct {
	VerificationID string    `json:"verification_id"`
	Status         string    `json:"status"` // "pending", "approved", "rejected", "requires_action"
	Message        string    `json:"message"`
	SubmittedAt    time.Time `json:"submitted_at"`
}

type KYCStatus struct {
	VerificationID string    `json:"verification_id"`
	Status         string    `json:"status"`
	CheckedAt      time.Time `json:"checked_at"`
	Details        map[string]interface{} `json:"details,omitempty"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type OnboardingFlow struct {
	UserID         string                 `json:"user_id"`
	AccountType    string                 `json:"account_type"`
	CurrentStep    string                 `json:"current_step"`
	CompletedSteps []string               `json:"completed_steps"`
	Data           map[string]interface{} `json:"data"`
	StartedAt      time.Time              `json:"started_at"`
	LastUpdated    time.Time              `json:"last_updated"`
	Status         string                 `json:"status"` // "in_progress", "completed", "failed"
}

type OnboardingStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Order       int                    `json:"order"`
	DataFields  []DataField            `json:"data_fields"`
	Validation  ValidationRules        `json:"validation"`
}

type DataField struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // "text", "email", "phone", "date", "file", "select"
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
	Options     []string `json:"options,omitempty"`
}

type ValidationRules struct {
	MinLength    int      `json:"min_length,omitempty"`
	MaxLength    int      `json:"max_length,omitempty"`
	Pattern      string   `json:"pattern,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	CustomValidation string `json:"custom_validation,omitempty"`
}

func NewOnboardingEngine(redisClient *redis.Client) *OnboardingEngine {
	return &OnboardingEngine{
		redis: redisClient,
		kyc: &KYCManager{
			redis:     redisClient,
			providers: make(map[string]KYCProvider),
		},
	}
}

// RegisterKYCProvider registers a KYC verification provider
func (e *OnboardingEngine) RegisterKYCProvider(name string, provider KYCProvider) {
	e.kyc.providers[name] = provider
}

// StartOnboarding initiates the onboarding flow for a new user
func (e *OnboardingEngine) StartOnboarding(ctx context.Context, userID, accountType string) (*OnboardingFlow, error) {
	flow := &OnboardingFlow{
		UserID:         userID,
		AccountType:    accountType,
		CurrentStep:    "basic_info",
		CompletedSteps: []string{},
		Data:           make(map[string]interface{}),
		StartedAt:      time.Now(),
		LastUpdated:    time.Now(),
		Status:         "in_progress",
	}

	if err := e.saveOnboardingFlow(ctx, flow); err != nil {
		return nil, err
	}

	return flow, nil
}

// GetOnboardingFlow retrieves the current onboarding flow for a user
func (e *OnboardingEngine) GetOnboardingFlow(ctx context.Context, userID string) (*OnboardingFlow, error) {
	key := fmt.Sprintf("onboarding:%s", userID)
	data, err := e.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var flow OnboardingFlow
	if err := json.Unmarshal([]byte(data), &flow); err != nil {
		return nil, err
	}

	return &flow, nil
}

// CompleteStep completes a step in the onboarding flow
func (e *OnboardingEngine) CompleteStep(ctx context.Context, userID, stepID string, stepData map[string]interface{}) error {
	flow, err := e.GetOnboardingFlow(ctx, userID)
	if err != nil {
		return err
	}

	// Validate step data
	if err := e.validateStepData(stepID, stepData); err != nil {
		return err
	}

	// Store step data
	for key, value := range stepData {
		flow.Data[key] = value
	}

	// Mark step as completed
	flow.CompletedSteps = append(flow.CompletedSteps, stepID)

	// Determine next step
	nextStep := e.getNextStep(flow.AccountType, stepID)
	if nextStep != "" {
		flow.CurrentStep = nextStep
	} else {
		flow.Status = "completed"
	}

	flow.LastUpdated = time.Now()

	return e.saveOnboardingFlow(ctx, flow)
}

// SubmitKYCVerification submits KYC verification for a user
func (e *OnboardingEngine) SubmitKYCVerification(ctx context.Context, request KYCRequest) (*KYCResponse, error) {
	// Use default provider (in production, select based on user region/country)
	provider, ok := e.kyc.providers["default"]
	if !ok {
		return nil, fmt.Errorf("no KYC provider available")
	}

	response, err := provider.SubmitVerification(ctx, request)
	if err != nil {
		return nil, err
	}

	// Store verification ID in onboarding flow
	flow, err := e.GetOnboardingFlow(ctx, request.UserID)
	if err == nil {
		flow.Data["kyc_verification_id"] = response.VerificationID
		flow.Data["kyc_status"] = response.Status
		e.saveOnboardingFlow(ctx, flow)
	}

	return response, nil
}

// CheckKYCStatus checks the status of a KYC verification
func (e *OnboardingEngine) CheckKYCStatus(ctx context.Context, userID string) (*KYCStatus, error) {
	flow, err := e.GetOnboardingFlow(ctx, userID)
	if err != nil {
		return nil, err
	}

	verificationID, ok := flow.Data["kyc_verification_id"].(string)
	if !ok {
		return nil, fmt.Errorf("no KYC verification found for user")
	}

	provider, ok := e.kyc.providers["default"]
	if !ok {
		return nil, fmt.Errorf("no KYC provider available")
	}

	return provider.CheckStatus(ctx, verificationID)
}

// GetOnboardingSteps returns the steps for a given account type
func (e *OnboardingEngine) GetOnboardingSteps(accountType string) []OnboardingStep {
	switch accountType {
	case "advertiser":
		return e.getAdvertiserSteps()
	case "publisher":
		return e.getPublisherSteps()
	default:
		return []OnboardingStep{}
	}
}

// ValidateOnboarding validates the entire onboarding flow
func (e *OnboardingEngine) ValidateOnboarding(ctx context.Context, userID string) (bool, []string, error) {
	flow, err := e.GetOnboardingFlow(ctx, userID)
	if err != nil {
		return false, []string{}, err
	}

	steps := e.GetOnboardingSteps(flow.AccountType)
	var errors []string

	for _, step := range steps {
		if step.Required {
			if !contains(flow.CompletedSteps, step.ID) {
				errors = append(errors, fmt.Sprintf("Required step '%s' not completed", step.Name))
			}
		}
	}

	// Check KYC status if required
	if flow.Data["kyc_verification_id"] != nil {
		status, err := e.CheckKYCStatus(ctx, userID)
		if err != nil {
			errors = append(errors, "Failed to check KYC status")
		} else if status.Status != "approved" {
			errors = append(errors, fmt.Sprintf("KYC verification not approved: %s", status.Status))
		}
	}

	return len(errors) == 0, errors, nil
}

// Helper methods

func (e *OnboardingEngine) saveOnboardingFlow(ctx context.Context, flow *OnboardingFlow) error {
	data, err := json.Marshal(flow)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("onboarding:%s", flow.UserID)
	// Cache for 7 days
	return e.redis.Set(ctx, key, string(data), 7*24*time.Hour)
}

func (e *OnboardingEngine) validateStepData(stepID string, data map[string]interface{}) error {
	// Basic validation - in production, implement more sophisticated validation
	switch stepID {
	case "basic_info":
		if _, ok := data["email"]; !ok {
			return fmt.Errorf("email is required")
		}
		if _, ok := data["phone"]; !ok {
			return fmt.Errorf("phone is required")
		}
	case "kyc_verification":
		if _, ok := data["document_type"]; !ok {
			return fmt.Errorf("document_type is required")
		}
	}
	return nil
}

func (e *OnboardingEngine) getNextStep(accountType, currentStep string) string {
	steps := e.getStepOrder(accountType)
	for i, step := range steps {
		if step == currentStep && i < len(steps)-1 {
			return steps[i+1]
		}
	}
	return ""
}

func (e *OnboardingEngine) getStepOrder(accountType string) []string {
	switch accountType {
	case "advertiser":
		return []string{
			"basic_info",
			"business_info",
			"payment_method",
			"kyc_verification",
			"campaign_setup",
		}
	case "publisher":
		return []string{
			"basic_info",
			"site_verification",
			"payment_method",
			"kyc_verification",
			"zone_setup",
		}
	default:
		return []string{}
	}
}

func (e *OnboardingEngine) getAdvertiserSteps() []OnboardingStep {
	return []OnboardingStep{
		{
			ID:          "basic_info",
			Name:        "Basic Information",
			Description: "Provide your personal contact information",
			Required:    true,
			Order:       1,
			DataFields: []DataField{
				{ID: "first_name", Name: "First Name", Type: "text", Required: true},
				{ID: "last_name", Name: "Last Name", Type: "text", Required: true},
				{ID: "email", Name: "Email", Type: "email", Required: true},
				{ID: "phone", Name: "Phone", Type: "phone", Required: true},
			},
		},
		{
			ID:          "business_info",
			Name:        "Business Information",
			Description: "Provide your business details",
			Required:    true,
			Order:       2,
			DataFields: []DataField{
				{ID: "company_name", Name: "Company Name", Type: "text", Required: true},
				{ID: "registration_number", Name: "Registration Number", Type: "text", Required: true},
				{ID: "tax_id", Name: "Tax ID", Type: "text", Required: true},
				{ID: "website", Name: "Website", Type: "text", Required: false},
				{ID: "industry", Name: "Industry", Type: "select", Required: true, Options: []string{"Technology", "E-commerce", "Finance", "Healthcare", "Education", "Other"}},
			},
		},
		{
			ID:          "payment_method",
			Name:        "Payment Method",
			Description: "Set up your payment method",
			Required:    true,
			Order:       3,
			DataFields: []DataField{
				{ID: "payment_type", Name: "Payment Type", Type: "select", Required: true, Options: []string{"credit_card", "bank_transfer", "mobile_money"}},
				{ID: "currency", Name: "Currency", Type: "select", Required: true, Options: []string{"USD", "KES", "UGX", "TZS", "NGN"}},
			},
		},
		{
			ID:          "kyc_verification",
			Name:        "Identity Verification",
			Description: "Verify your identity with government ID",
			Required:    true,
			Order:       4,
			DataFields: []DataField{
				{ID: "document_type", Name: "Document Type", Type: "select", Required: true, Options: []string{"passport", "id_card", "driving_license"}},
				{ID: "front_image", Name: "Document Front", Type: "file", Required: true},
				{ID: "back_image", Name: "Document Back", Type: "file", Required: false},
				{ID: "selfie_image", Name: "Selfie", Type: "file", Required: true},
			},
		},
		{
			ID:          "campaign_setup",
			Name:        "Campaign Setup",
			Description: "Create your first campaign",
			Required:    false,
			Order:       5,
			DataFields: []DataField{
				{ID: "campaign_name", Name: "Campaign Name", Type: "text", Required: true},
				{ID: "budget", Name: "Daily Budget", Type: "text", Required: true},
			},
		},
	}
}

func (e *OnboardingEngine) getPublisherSteps() []OnboardingStep {
	return []OnboardingStep{
		{
			ID:          "basic_info",
			Name:        "Basic Information",
			Description: "Provide your personal contact information",
			Required:    true,
			Order:       1,
			DataFields: []DataField{
				{ID: "first_name", Name: "First Name", Type: "text", Required: true},
				{ID: "last_name", Name: "Last Name", Type: "text", Required: true},
				{ID: "email", Name: "Email", Type: "email", Required: true},
				{ID: "phone", Name: "Phone", Type: "phone", Required: true},
			},
		},
		{
			ID:          "site_verification",
			Name:        "Site Verification",
			Description: "Verify ownership of your website",
			Required:    true,
			Order:       2,
			DataFields: []DataField{
				{ID: "site_url", Name: "Website URL", Type: "text", Required: true},
				{ID: "site_name", Name: "Site Name", Type: "text", Required: true},
				{ID: "site_category", Name: "Site Category", Type: "select", Required: true, Options: []string{"Technology", "News", "Entertainment", "Sports", "Finance", "Health", "Other"}},
			},
		},
		{
			ID:          "payment_method",
			Name:        "Payment Method",
			Description: "Set up your payment method",
			Required:    true,
			Order:       3,
			DataFields: []DataField{
				{ID: "payment_type", Name: "Payment Type", Type: "select", Required: true, Options: []string{"bank_transfer", "mobile_money"}},
				{ID: "currency", Name: "Currency", Type: "select", Required: true, Options: []string{"USD", "KES", "UGX", "TZS", "NGN"}},
			},
		},
		{
			ID:          "kyc_verification",
			Name:        "Identity Verification",
			Description: "Verify your identity with government ID",
			Required:    true,
			Order:       4,
			DataFields: []DataField{
				{ID: "document_type", Name: "Document Type", Type: "select", Required: true, Options: []string{"passport", "id_card", "driving_license"}},
				{ID: "front_image", Name: "Document Front", Type: "file", Required: true},
				{ID: "back_image", Name: "Document Back", Type: "file", Required: false},
				{ID: "selfie_image", Name: "Selfie", Type: "file", Required: true},
			},
		},
		{
			ID:          "zone_setup",
			Name:        "Ad Zone Setup",
			Description: "Create your first ad zone",
			Required:    false,
			Order:       5,
			DataFields: []DataField{
				{ID: "zone_name", Name: "Zone Name", Type: "text", Required: true},
				{ID: "ad_format", Name: "Ad Format", Type: "select", Required: true, Options: []string{"banner", "native", "popunder", "push"}},
			},
		},
	}
}

// GenerateVerificationToken generates a secure token for verification
func GenerateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
