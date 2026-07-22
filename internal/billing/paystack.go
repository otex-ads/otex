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

// GetSecretKey returns the secret key for webhook verification
func (p *PaystackClient) GetSecretKey() string {
	return p.secretKey
}

// --- Transfer Recipients ---

// CreateTransferRecipientRequest represents the request to create a transfer recipient
type CreateTransferRecipientRequest struct {
	Type     string `json:"type"`      // mobile_money, nuban, basa, etc.
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Currency string `json:"currency"`
	BankCode string `json:"bank_code,omitempty"` // e.g. MPESA
}

// CreateTransferRecipientResponse represents the response from creating a transfer recipient
type CreateTransferRecipientResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Active        bool   `json:"active"`
		Currency      string `json:"currency"`
		Domain        string `json:"domain"`
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		RecipientCode string `json:"recipient_code"`
		Type          string `json:"type"`
		CreatedAt     string `json:"created_at"`
	} `json:"data"`
}

// CreateTransferRecipient creates a transfer recipient in Paystack
func (p *PaystackClient) CreateTransferRecipient(req CreateTransferRecipientRequest) (*CreateTransferRecipientResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.baseURL+"/transferrecipient", bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Paystack API error: %s", string(respBody))
	}

	var result CreateTransferRecipientResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("Paystack error: %s", result.Message)
	}

	return &result, nil
}

// --- Transfers ---

// InitiateTransferRequest represents the request to initiate a transfer
type InitiateTransferRequest struct {
	Source    string `json:"source"`    // "balance"
	Amount   int64  `json:"amount"`    // in subunits (cents/kobo)
	Recipient string `json:"recipient"` // RCP_xxxx
	Reason   string `json:"reason"`
	Reference string `json:"reference"`
}

// InitiateTransferResponse represents the response from initiating a transfer
type InitiateTransferResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID            int64  `json:"id"`
		TransferCode  string `json:"transfer_code"`
		Amount        int64  `json:"amount"`
		Currency      string `json:"currency"`
		Status        string `json:"status"`
		Reference     string `json:"reference"`
		Recipient     int64  `json:"recipient"`
		CreatedAt     string `json:"created_at"`
	} `json:"data"`
}

// InitiateTransfer initiates a transfer to a recipient
func (p *PaystackClient) InitiateTransfer(req InitiateTransferRequest) (*InitiateTransferResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.baseURL+"/transfer", bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Paystack API error: %s", string(respBody))
	}

	var result InitiateTransferResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("Paystack error: %s", result.Message)
	}

	return &result, nil
}

// --- Balance ---

// BalanceResponse represents the response from getting the Paystack balance
type BalanceResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    []struct {
		Currency string `json:"currency"`
		Balance  int64  `json:"balance"`
	} `json:"data"`
}

// GetBalance fetches the current Paystack account balance
func (p *PaystackClient) GetBalance() (*BalanceResponse, error) {
	httpReq, err := http.NewRequest("GET", p.baseURL+"/balance", nil)
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

	var result BalanceResponse
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
