package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

// ResendClient wraps the Resend SDK
type ResendClient struct {
	client *resend.Client
	from   string
}

// NewResendClient creates a new Resend client
func NewResendClient(apiKey, fromEmail string) *ResendClient {
	return &ResendClient{
		client: resend.NewClient(apiKey),
		from:   fromEmail,
	}
}

// Send sends an email via Resend
func (c *ResendClient) Send(ctx context.Context, to, subject, html string) (string, error) {
	params := &resend.SendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	resp, err := c.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to send email via Resend: %w", err)
	}

	if resp.Id == "" {
		return "", fmt.Errorf("Resend returned empty ID")
	}

	return resp.Id, nil
}
