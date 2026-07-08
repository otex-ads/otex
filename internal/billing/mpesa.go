package billing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// M-Pesa Daraja API configuration
type MpesaConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	Environment    string // "sandbox" or "production"
	Shortcode      string
	Passkey        string
	CallbackURL    string
}

type MpesaClient struct {
	config     MpesaConfig
	httpClient *http.Client
	authToken  string
	tokenExpiry time.Time
}

func NewMpesaClient(config MpesaConfig) *MpesaClient {
	return &MpesaClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// STKPushRequest represents the request for M-Pesa STK Push
type STKPushRequest struct {
	BusinessShortCode string `json:"business_short_code"`
	Password          string `json:"password"`
	Timestamp         string `json:"timestamp"`
	TransactionType   string `json:"transaction_type"`
	Amount            string `json:"amount"`
	PartyA            string `json:"party_a"`
	PartyB            string `json:"party_b"`
	PhoneNumber       string `json:"phone_number"`
	CallBackURL       string `json:"call_back_url"`
	AccountReference  string `json:"account_reference"`
	TransactionDesc   string `json:"transaction_desc"`
}

// STKPushResponse represents the response from M-Pesa STK Push
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseMessage     string `json:"ResponseMessage"`
	CustomerMessage     string `json:"CustomerMessage"`
}

// STKPushCallback represents the callback from M-Pesa after STK Push
type STKPushCallback struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  []struct {
				Name  string `json:"Name"`
				Value string `json:"Value"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// B2CRequest represents the request for M-Pesa B2C (Business to Customer)
type B2CRequest struct {
	InitiatorName              string `json:"initiator_name"`
	SecurityCredential         string `json:"security_credential"`
	CommandID                  string `json:"command_id"`
	Amount                     string `json:"amount"`
	PartyA                     string `json:"party_a"`
	PartyB                     string `json:"party_b"`
	Remarks                    string `json:"remarks"`
	QueueTimeOutURL            string `json:"queue_time_out_url"`
	ResultURL                  string `json:"result_url"`
	Occasion                   string `json:"occasion"`
}

// B2CResponse represents the response from M-Pesa B2C
type B2CResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	ConversationID      string `json:"ConversationID"`
	OriginatorConversationID string `json:"OriginatorConversationID"`
}

// getBaseURL returns the appropriate base URL based on environment
func (c *MpesaClient) getBaseURL() string {
	if c.config.Environment == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

// authenticate gets OAuth token from M-Pesa
func (c *MpesaClient) authenticate() error {
	// Check if token is still valid
	if c.authToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	url := fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", c.getBaseURL())
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	
	req.SetBasicAuth(c.config.ConsumerKey, c.config.ConsumerSecret)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	var authResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	
	if err := json.Unmarshal(body, &authResp); err != nil {
		return err
	}
	
	c.authToken = authResp.AccessToken
	c.tokenExpiry = time.Now().Add(55 * time.Minute) // Token expires in 1 hour, refresh early
	
	return nil
}

// generatePassword generates the password for STK Push
func (c *MpesaClient) generatePassword(timestamp string) string {
	data := c.config.Shortcode + c.config.Passkey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// generateTimestamp generates timestamp in format YYYYMMDDHHmmss
func generateTimestamp() string {
	return time.Now().Format("20060102150405")
}

// InitiateSTKPush initiates an M-Pesa STK Push request
func (c *MpesaClient) InitiateSTKPush(phoneNumber, amount, accountRef string) (*STKPushResponse, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	timestamp := generateTimestamp()
	password := c.generatePassword(timestamp)

	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", c.getBaseURL())

	req := STKPushRequest{
		BusinessShortCode: c.config.Shortcode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            phoneNumber,
		PartyB:            c.config.Shortcode,
		PhoneNumber:       phoneNumber,
		CallBackURL:       c.config.CallbackURL,
		AccountReference:  accountRef,
		TransactionDesc:    "Ad Network Top-up",
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response STKPushResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// InitiateB2C initiates an M-Pesa B2C (Business to Customer) payment
func (c *MpesaClient) InitiateB2C(phoneNumber, amount, remarks string) (*B2CResponse, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/mpesa/b2c/v1/paymentrequest", c.getBaseURL())

	req := B2CRequest{
		InitiatorName:      "testapi",
		SecurityCredential: c.config.Passkey, // In production, this should be encrypted
		CommandID:          "BusinessPayment",
		Amount:             amount,
		PartyA:             c.config.Shortcode,
		PartyB:             phoneNumber,
		Remarks:            remarks,
		QueueTimeOutURL:    c.config.CallbackURL + "/timeout",
		ResultURL:          c.config.CallbackURL + "/result",
		Occasion:           "Payout",
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response B2CResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ParseSTKPushCallback parses the callback from M-Pesa STK Push
func ParseSTKPushCallback(data []byte) (*STKPushCallback, error) {
	var callback STKPushCallback
	if err := json.Unmarshal(data, &callback); err != nil {
		return nil, err
	}
	return &callback, nil
}

// GetMpesaCodeFromCallback extracts the M-Pesa transaction code from callback
func GetMpesaCodeFromCallback(callback *STKPushCallback) string {
	for _, meta := range callback.Body.StkCallback.CallbackMetadata {
		if meta.Name == "MpesaReceipt" {
			return meta.Value
		}
	}
	return ""
}

// GetAmountFromCallback extracts the amount from callback
func GetAmountFromCallback(callback *STKPushCallback) string {
	for _, meta := range callback.Body.StkCallback.CallbackMetadata {
		if meta.Name == "Amount" {
			return meta.Value
		}
	}
	return ""
}

// GetPhoneNumberFromCallback extracts the phone number from callback
func GetPhoneNumberFromCallback(callback *STKPushCallback) string {
	for _, meta := range callback.Body.StkCallback.CallbackMetadata {
		if meta.Name == "PhoneNumber" {
			return meta.Value
		}
	}
	return ""
}
