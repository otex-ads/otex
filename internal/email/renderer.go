package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RendererClient calls the email renderer service
type RendererClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRendererClient creates a new renderer client
func NewRendererClient(baseURL string) *RendererClient {
	return &RendererClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Render renders an email template to HTML
func (c *RendererClient) Render(ctx context.Context, template TemplateName, data map[string]interface{}) (string, error) {
	req := RenderRequest{
		Template: string(template),
		Data:     data,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal render request: %w", err)
	}

	url := fmt.Sprintf("%s/api/public/email/render", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call renderer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("renderer returned status %d: %s", resp.StatusCode, string(respBody))
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(html), nil
}
