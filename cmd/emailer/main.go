package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"adnet/internal/email"
)

type Emailer struct {
	renderer  *email.RendererClient
	resend    *email.ResendClient
	port      string
}

func main() {
	rendererURL := getEnv("EMAIL_RENDERER_URL", "http://email-renderer:3000")
	resendAPIKey := getEnv("RESEND_API_KEY", "")
	fromEmail := getEnv("RESEND_FROM_EMAIL", "noreply@otexads.com")
	port := getEnv("PORT", "8085")

	if resendAPIKey == "" {
		log.Fatal("RESEND_API_KEY is required")
	}

	e := &Emailer{
		renderer: email.NewRendererClient(rendererURL),
		resend:   email.NewResendClient(resendAPIKey, fromEmail),
		port:     port,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/send", e.handleSend)
	mux.HandleFunc("/health", e.handleHealth)

	server := &http.Server{
		Addr:    ":" + e.port,
		Handler: mux,
	}

	log.Printf("Emailer starting on port %s", e.port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func (e *Emailer) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req email.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Render HTML from template
	html, err := e.renderer.Render(ctx, req.Template, req.Data)
	if err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	// Send via Resend
	resendID, err := e.resend.Send(ctx, req.To, req.Subject, html)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"success":  true,
		"resend_id": resendID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (e *Emailer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
