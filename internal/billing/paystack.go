package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PaystackClient handles interactions with Paystack API
type PaystackClient struct {
	secretKey string
	baseURL   string
	client    *http.Client
}

// NewPaystackClient creates a new Paystack client
func NewPaystackClient(secretKey string) *PaystackClient {
	return &PaystackClient{
		secretKey: secretKey,
		baseURL:   "https://api.paystack.co",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitializeTransactionRequest represents the request to initialize a transaction
type InitializeTransactionRequest struct {
	Amount       string `json:"amount"`       // in kobo (lowest currency unit)
	Email        string `json:"email"`
	Currency     string `json:"currency,omitempty"`
	Reference    string `json:"reference,omitempty"`
	CallbackURL  string `json:"callback_url,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Channels     []string `json:"channels,omitempty"` // card, bank, ussd, qr, mobile_money
}

// InitializeTransactionResponse represents the response from transaction initialization
type InitializeTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// VerifyTransactionResponse represents the response from transaction verification
type VerifyTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID            int64  `json:"id"`
		Domain        string `json:"domain"`
		Status        string `json:"status"`
		Reference     string `json:"reference"`
		Amount        int64  `json:"amount"`
		Message       string `json:"message"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt        string `json:"paid_at"`
		CreatedAt      string `json:"created_at"`
		Channel       string `json:"channel"`
		Currency      string `json:"currency"`
		IPAddress     string `json:"ip_address"`
		Customer      struct {
			ID           int64  `json:"id"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
		} `json:"customer"`
	} `json:"data"`
}

// InitializeTransaction initializes a new payment transaction
func (p *PaystackClient) InitializeTransaction(req InitializeTransactionRequest) (*InitializeTransactionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.baseURL+"/transaction/initialize", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Paystack API error: %s", string(respBody))
	}

	var result InitializeTransactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("Paystack error: %s", result.Message)
	}

	return &result, nil
}

// VerifyTransaction verifies a payment transaction
func (p *PaystackClient) VerifyTransaction(reference string) (*VerifyTransactionResponse, error) {
	httpReq, err := http.NewRequest("GET", p.baseURL+"/transaction/verify/"+reference, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.secretKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Paystack API error: %s", string(respBody))
	}

	var result VerifyTransactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("Paystack error: %s", result.Message)
	}

	return &result, nil
}

// CreateCustomerRequest represents the request to create a customer
type CreateCustomerRequest struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateCustomerResponse represents the response from customer creation
type CreateCustomerResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID           int64  `json:"id"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		CustomerCode string `json:"customer_code"`
		Phone        string `json:"phone"`
		CreatedAt     string `json:"created_at"`
	} `json:"data"`
}

// CreateCustomer creates a new customer in Paystack
func (p *PaystackClient) CreateCustomer(req CreateCustomerRequest) (*CreateCustomerResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.baseURL+"/customer", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Paystack API error: %s", string(respBody))
	}

	var result CreateCustomerResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("Paystack error: %s", result.Message)
	}

	return &result, nil
}

// BanksResponse represents the response from listing banks
type BanksResponse struct {
	Status  bool `json:"status"`
	Message string `json:"message"`
	Data    []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Code        string `json:"code"`
		LongCode    string `json:"longcode"`
		Gateway     string `json:"gateway"`
		PayWithBank bool   `json:"pay_with_bank"`
		Active      bool   `json:"active"`
		Country     string `json:"country"`
		Currency    string `json:"currency"`
		Type        string `json:"type"`
		Slug        string `json:"slug"`
	} `json:"data"`
}

// ListBanks lists all banks supported by Paystack
func (p *PaystackClient) ListBanks(country string) (*BanksResponse, error) {
	url := p.baseURL + "/bank"
	if country != "" {
		url += "?country=" + country
	}

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.secretKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Paystack API error: %s", string(respBody))
	}

	var result BanksResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("Paystack error: %s", result.Message)
	}

	return &result, nil
}
