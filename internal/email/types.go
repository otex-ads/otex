package email

import "time"

// TemplateName represents the available email templates
type TemplateName string

const (
	TemplateWelcome               TemplateName = "welcome"
	TemplateVerifyEmail          TemplateName = "verify-email"
	TemplatePasswordReset        TemplateName = "password-reset"
	TemplateDepositConfirmation  TemplateName = "deposit-confirmation"
	TemplateLowBalance           TemplateName = "low-balance"
	TemplatePayoutSent           TemplateName = "payout-sent"
	TemplatePayoutPendingApproval TemplateName = "payout-pending-approval"
	TemplateAdminSignupNotification TemplateName = "admin-signup-notification"
	TemplateSiteModerationDecision TemplateName = "site-moderation-decision"
	TemplateCampaignModerationDecision TemplateName = "campaign-moderation-decision"
	TemplateCreativeModerationDecision TemplateName = "creative-moderation-decision"
)

// SendRequest is the request to send an email
type SendRequest struct {
	Template TemplateName         `json:"template"`
	To       string               `json:"to"`
	Subject  string               `json:"subject"`
	Data     map[string]interface{} `json:"data"`
}

// RenderRequest is the request to the email renderer
type RenderRequest struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

// RenderResponse is the response from the email renderer
type RenderResponse struct {
	HTML string `json:"html"`
}

// Email represents a sent email
type Email struct {
	ID        string    `json:"id"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Template  string    `json:"template"`
	SentAt    time.Time `json:"sent_at"`
	ResendID  string    `json:"resend_id,omitempty"`
}
