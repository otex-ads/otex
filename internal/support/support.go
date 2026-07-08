package support

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"adnet/internal/store/redis"
)

type SupportEngine struct {
	redis    *redis.Client
	ticketing *TicketingSystem
	chat     *LiveChatSystem
}

type TicketingSystem struct {
	redis *redis.Client
}

type LiveChatSystem struct {
	redis *redis.Client
}

type Ticket struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	AccountType string                 `json:"account_type"` // "advertiser" or "publisher"
	Subject     string                 `json:"subject"`
	Description string                `json:"description"`
	Category    string                 `json:"category"`
	Priority    string                 `json:"priority"` // "low", "medium", "high", "urgent"
	Status      string                 `json:"status"` // "open", "in_progress", "waiting_on_user", "resolved", "closed"
	AssignedTo  string                 `json:"assigned_to,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type TicketComment struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	UserID    string    `json:"user_id"`
	IsAdmin   bool      `json:"is_admin"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

type ChatSession struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	AccountType string                 `json:"account_type"`
	Status      string                 `json:"status"` // "active", "ended", "transferred"
	AssignedTo  string                 `json:"assigned_to,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	EndedAt     *time.Time             `json:"ended_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ChatMessage struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	IsAdmin   bool                   `json:"is_admin"`
	Content   string                 `json:"content"`
	CreatedAt time.Time              `json:"created_at"`
	Read      bool                   `json:"read"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type SupportAgent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"` // "agent", "supervisor", "admin"
	Status      string    `json:"status"` // "online", "away", "offline", "busy"
	Department  string    `json:"department"`
	Skills      []string  `json:"skills"`
	MaxChats    int       `json:"max_chats"`
	CurrentChats int      `json:"current_chats"`
	LastActive  time.Time `json:"last_active"`
}

type ChatQueue struct {
	WaitingUsers []ChatQueueItem `json:"waiting_users"`
	Agents       []SupportAgent  `json:"agents"`
}

type ChatQueueItem struct {
	UserID      string    `json:"user_id"`
	AccountType string    `json:"account_type"`
	QueuedAt    time.Time `json:"queued_at"`
	Priority    int       `json:"priority"`
}

type KnowledgeBaseArticle struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Language    string                 `json:"language"`
	Views       int64                  `json:"views"`
	Helpful     int64                  `json:"helpful"`
	NotHelpful  int64                  `json:"not_helpful"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Published   bool                   `json:"published"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func NewSupportEngine(redisClient *redis.Client) *SupportEngine {
	return &SupportEngine{
		redis: redisClient,
		ticketing: &TicketingSystem{
			redis: redisClient,
		},
		chat: &LiveChatSystem{
			redis: redisClient,
		},
	}
}

// Ticketing Methods

// CreateTicket creates a new support ticket
func (e *SupportEngine) CreateTicket(ctx context.Context, ticket Ticket) (*Ticket, error) {
	ticket.ID = generateTicketID()
	ticket.Status = "open"
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	if err := e.ticketing.saveTicket(ctx, &ticket); err != nil {
		return nil, err
	}

	// Notify agents
	e.notifyAgents(ctx, "ticket_created", ticket)

	return &ticket, nil
}

// GetTicket retrieves a ticket by ID
func (e *SupportEngine) GetTicket(ctx context.Context, ticketID string) (*Ticket, error) {
	return e.ticketing.getTicket(ctx, ticketID)
}

// UpdateTicket updates a ticket
func (e *SupportEngine) UpdateTicket(ctx context.Context, ticketID string, updates map[string]interface{}) error {
	ticket, err := e.ticketing.getTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "status":
			ticket.Status = value.(string)
			if ticket.Status == "resolved" || ticket.Status == "closed" {
				now := time.Now()
				ticket.ResolvedAt = &now
			}
		case "priority":
			ticket.Priority = value.(string)
		case "assigned_to":
			ticket.AssignedTo = value.(string)
		case "tags":
			ticket.Tags = value.([]string)
		}
	}

	ticket.UpdatedAt = time.Now()

	return e.ticketing.saveTicket(ctx, ticket)
}

// AddComment adds a comment to a ticket
func (e *SupportEngine) AddComment(ctx context.Context, ticketID, userID string, isAdmin bool, content string, attachments []Attachment) (*TicketComment, error) {
	comment := TicketComment{
		ID:        generateCommentID(),
		TicketID:  ticketID,
		UserID:    userID,
		IsAdmin:   isAdmin,
		Content:   content,
		CreatedAt: time.Now(),
		Attachments: attachments,
	}

	if err := e.ticketing.saveComment(ctx, &comment); err != nil {
		return nil, err
	}

	// Update ticket timestamp
	e.UpdateTicket(ctx, ticketID, map[string]interface{}{"updated_at": time.Now()})

	// Notify relevant parties
	e.notifyTicketUpdate(ctx, ticketID, comment)

	return &comment, nil
}

// GetTicketComments retrieves all comments for a ticket
func (e *SupportEngine) GetTicketComments(ctx context.Context, ticketID string) ([]TicketComment, error) {
	return e.ticketing.getComments(ctx, ticketID)
}

// ListUserTickets retrieves all tickets for a user
func (e *SupportEngine) ListUserTickets(ctx context.Context, userID string, status string) ([]Ticket, error) {
	return e.ticketing.listUserTickets(ctx, userID, status)
}

// ListAllTickets retrieves all tickets (for admin)
func (e *SupportEngine) ListAllTickets(ctx context.Context, status, priority string, limit int) ([]Ticket, error) {
	return e.ticketing.listAllTickets(ctx, status, priority, limit)
}

// Live Chat Methods

// StartChatSession starts a new chat session
func (e *SupportEngine) StartChatSession(ctx context.Context, userID, accountType string) (*ChatSession, error) {
	session := ChatSession{
		ID:          generateSessionID(),
		UserID:      userID,
		AccountType: accountType,
		Status:      "active",
		StartedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	if err := e.chat.saveSession(ctx, &session); err != nil {
		return nil, err
	}

	// Add to queue
	e.chat.addToQueue(ctx, ChatQueueItem{
		UserID:      userID,
		AccountType: accountType,
		QueuedAt:    time.Now(),
		Priority:    1,
	})

	// Try to assign an agent
	e.chat.assignAgent(ctx, &session)

	return &session, nil
}

// GetChatSession retrieves a chat session
func (e *SupportEngine) GetChatSession(ctx context.Context, sessionID string) (*ChatSession, error) {
	return e.chat.getSession(ctx, sessionID)
}

// EndChatSession ends a chat session
func (e *SupportEngine) EndChatSession(ctx context.Context, sessionID string) error {
	session, err := e.chat.getSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Status = "ended"
	now := time.Now()
	session.EndedAt = &now

	return e.chat.saveSession(ctx, session)
}

// SendChatMessage sends a message in a chat session
func (e *SupportEngine) SendChatMessage(ctx context.Context, sessionID, userID string, isAdmin bool, content string) (*ChatMessage, error) {
	message := ChatMessage{
		ID:        generateMessageID(),
		SessionID: sessionID,
		UserID:    userID,
		IsAdmin:   isAdmin,
		Content:   content,
		CreatedAt: time.Now(),
		Read:      false,
		Metadata:  make(map[string]interface{}),
	}

	if err := e.chat.saveMessage(ctx, &message); err != nil {
		return nil, err
	}

	// Notify via pub/sub
	e.chat.publishMessage(ctx, sessionID, message)

	return &message, nil
}

// GetChatMessages retrieves messages for a session
func (e *SupportEngine) GetChatMessages(ctx context.Context, sessionID string, limit int) ([]ChatMessage, error) {
	return e.chat.getMessages(ctx, sessionID, limit)
}

// MarkMessagesAsRead marks messages as read
func (e *SupportEngine) MarkMessagesAsRead(ctx context.Context, sessionID, userID string) error {
	return e.chat.markAsRead(ctx, sessionID, userID)
}

// GetActiveChatSessions retrieves active chat sessions for an agent
func (e *SupportEngine) GetActiveChatSessions(ctx context.Context, agentID string) ([]ChatSession, error) {
	return e.chat.getActiveSessions(ctx, agentID)
}

// Support Agent Methods

// RegisterAgent registers a support agent
func (e *SupportEngine) RegisterAgent(ctx context.Context, agent SupportAgent) error {
	agent.ID = generateAgentID()
	agent.Status = "offline"
	agent.LastActive = time.Now()
	agent.CurrentChats = 0

	return e.chat.saveAgent(ctx, &agent)
}

// UpdateAgentStatus updates an agent's status
func (e *SupportEngine) UpdateAgentStatus(ctx context.Context, agentID, status string) error {
	agent, err := e.chat.getAgent(ctx, agentID)
	if err != nil {
		return err
	}

	agent.Status = status
	agent.LastActive = time.Now()

	return e.chat.saveAgent(ctx, agent)
}

// GetAgent retrieves an agent
func (e *SupportEngine) GetAgent(ctx context.Context, agentID string) (*SupportAgent, error) {
	return e.chat.getAgent(ctx, agentID)
}

// ListAgents retrieves all agents
func (e *SupportEngine) ListAgents(ctx context.Context, department string) ([]SupportAgent, error) {
	return e.chat.listAgents(ctx, department)
}

// Knowledge Base Methods

// CreateArticle creates a knowledge base article
func (e *SupportEngine) CreateArticle(ctx context.Context, article KnowledgeBaseArticle) (*KnowledgeBaseArticle, error) {
	article.ID = generateArticleID()
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()
	article.Views = 0
	article.Helpful = 0
	article.NotHelpful = 0

	if err := e.saveArticle(ctx, &article); err != nil {
		return nil, err
	}

	return &article, nil
}

// GetArticle retrieves a knowledge base article
func (e *SupportEngine) GetArticle(ctx context.Context, articleID string) (*KnowledgeBaseArticle, error) {
	article, err := e.getArticle(ctx, articleID)
	if err != nil {
		return nil, err
	}

	// Increment view count
	article.Views++
	e.saveArticle(ctx, article)

	return article, nil
}

// SearchArticles searches knowledge base articles
func (e *SupportEngine) SearchArticles(ctx context.Context, query string, category string, limit int) ([]KnowledgeBaseArticle, error) {
	return e.searchArticles(ctx, query, category, limit)
}

// RateArticle rates an article as helpful or not
func (e *SupportEngine) RateArticle(ctx context.Context, articleID string, helpful bool) error {
	article, err := e.getArticle(ctx, articleID)
	if err != nil {
		return err
	}

	if helpful {
		article.Helpful++
	} else {
		article.NotHelpful++
	}

	article.UpdatedAt = time.Now()

	return e.saveArticle(ctx, article)
}

// Helper methods for TicketingSystem

func (s *TicketingSystem) saveTicket(ctx context.Context, ticket *Ticket) error {
	data, err := json.Marshal(ticket)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("ticket:%s", ticket.ID)
	if err := s.redis.Set(ctx, key, string(data), 0); err != nil {
		return err
	}

	// Add to user's ticket list
	userKey := fmt.Sprintf("user:tickets:%s", ticket.UserID)
	s.redis.SAdd(ctx, userKey, ticket.ID)

	// Add to global ticket list
	s.redis.SAdd(ctx, "tickets:all", ticket.ID)

	return nil
}

func (s *TicketingSystem) getTicket(ctx context.Context, ticketID string) (*Ticket, error) {
	key := fmt.Sprintf("ticket:%s", ticketID)
	data, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var ticket Ticket
	if err := json.Unmarshal([]byte(data), &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (s *TicketingSystem) saveComment(ctx context.Context, comment *TicketComment) error {
	data, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("ticket:comment:%s", comment.ID)
	if err := s.redis.Set(ctx, key, string(data), 0); err != nil {
		return err
	}

	// Add to ticket's comment list
	ticketKey := fmt.Sprintf("ticket:comments:%s", comment.TicketID)
	s.redis.SAdd(ctx, ticketKey, comment.ID)

	return nil
}

func (s *TicketingSystem) getComments(ctx context.Context, ticketID string) ([]TicketComment, error) {
	ticketKey := fmt.Sprintf("ticket:comments:%s", ticketID)
	commentIDs, err := s.redis.SMembers(ctx, ticketKey)
	if err != nil {
		return nil, err
	}

	var comments []TicketComment
	for _, commentID := range commentIDs {
		key := fmt.Sprintf("ticket:comment:%s", commentID)
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			continue
		}

		var comment TicketComment
		if err := json.Unmarshal([]byte(data), &comment); err != nil {
			continue
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (s *TicketingSystem) listUserTickets(ctx context.Context, userID string, status string) ([]Ticket, error) {
	userKey := fmt.Sprintf("user:tickets:%s", userID)
	ticketIDs, err := s.redis.SMembers(ctx, userKey)
	if err != nil {
		return nil, err
	}

	var tickets []Ticket
	for _, ticketID := range ticketIDs {
		ticket, err := s.getTicket(ctx, ticketID)
		if err != nil {
			continue
		}

		if status == "" || ticket.Status == status {
			tickets = append(tickets, *ticket)
		}
	}

	return tickets, nil
}

func (s *TicketingSystem) listAllTickets(ctx context.Context, status, priority string, limit int) ([]Ticket, error) {
	ticketIDs, err := s.redis.SMembers(ctx, "tickets:all")
	if err != nil {
		return nil, err
	}

	var tickets []Ticket
	count := 0
	for _, ticketID := range ticketIDs {
		if limit > 0 && count >= limit {
			break
		}

		ticket, err := s.getTicket(ctx, ticketID)
		if err != nil {
			continue
		}

		if (status == "" || ticket.Status == status) && (priority == "" || ticket.Priority == priority) {
			tickets = append(tickets, *ticket)
			count++
		}
	}

	return tickets, nil
}

// Helper methods for LiveChatSystem

func (c *LiveChatSystem) saveSession(ctx context.Context, session *ChatSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("chat:session:%s", session.ID)
	return c.redis.Set(ctx, key, string(data), 0)
}

func (c *LiveChatSystem) getSession(ctx context.Context, sessionID string) (*ChatSession, error) {
	key := fmt.Sprintf("chat:session:%s", sessionID)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var session ChatSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (c *LiveChatSystem) saveMessage(ctx context.Context, message *ChatMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("chat:message:%s", message.ID)
	if err := c.redis.Set(ctx, key, string(data), 0); err != nil {
		return err
	}

	// Add to session's message list
	sessionKey := fmt.Sprintf("chat:messages:%s", message.SessionID)
	c.redis.SAdd(ctx, sessionKey, message.ID)

	return nil
}

func (c *LiveChatSystem) getMessages(ctx context.Context, sessionID string, limit int) ([]ChatMessage, error) {
	sessionKey := fmt.Sprintf("chat:messages:%s", sessionID)
	messageIDs, err := c.redis.SMembers(ctx, sessionKey)
	if err != nil {
		return nil, err
	}

	var messages []ChatMessage
	count := 0
	for _, messageID := range messageIDs {
		if limit > 0 && count >= limit {
			break
		}

		key := fmt.Sprintf("chat:message:%s", messageID)
		data, err := c.redis.Get(ctx, key)
		if err != nil {
			continue
		}

		var message ChatMessage
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			continue
		}

		messages = append(messages, message)
		count++
	}

	return messages, nil
}

func (c *LiveChatSystem) markAsRead(ctx context.Context, sessionID, userID string) error {
	sessionKey := fmt.Sprintf("chat:messages:%s", sessionID)
	messageIDs, err := c.redis.SMembers(ctx, sessionKey)
	if err != nil {
		return err
	}

	for _, messageID := range messageIDs {
		key := fmt.Sprintf("chat:message:%s", messageID)
		data, err := c.redis.Get(ctx, key)
		if err != nil {
			continue
		}

		var message ChatMessage
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			continue
		}

		if message.UserID != userID {
			now := time.Now()
			message.Read = true
			message.ReadAt = &now
			c.saveMessage(ctx, &message)
		}
	}

	return nil
}

func (c *LiveChatSystem) addToQueue(ctx context.Context, item ChatQueueItem) error {
	queueKey := "chat:queue"
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return c.redis.LPush(ctx, queueKey, string(data))
}

func (c *LiveChatSystem) assignAgent(ctx context.Context, session *ChatSession) error {
	// Find available agent
	agents, err := c.listAgents(ctx, "")
	if err != nil {
		return err
	}

	for _, agent := range agents {
		if agent.Status == "online" && agent.CurrentChats < agent.MaxChats {
			session.AssignedTo = agent.ID
			agent.CurrentChats++
			c.saveAgent(ctx, &agent)
			c.saveSession(ctx, session)
			return nil
		}
	}

	return nil
}

func (c *LiveChatSystem) getActiveSessions(ctx context.Context, agentID string) ([]ChatSession, error) {
	// In production, this would use an index
	// For now, return empty
	return []ChatSession{}, nil
}

func (c *LiveChatSystem) saveAgent(ctx context.Context, agent *SupportAgent) error {
	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("chat:agent:%s", agent.ID)
	return c.redis.Set(ctx, key, string(data), 0)
}

func (c *LiveChatSystem) getAgent(ctx context.Context, agentID string) (*SupportAgent, error) {
	key := fmt.Sprintf("chat:agent:%s", agentID)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var agent SupportAgent
	if err := json.Unmarshal([]byte(data), &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *LiveChatSystem) listAgents(ctx context.Context, department string) ([]SupportAgent, error) {
	// In production, this would use an index
	// For now, return empty
	return []SupportAgent{}, nil
}

func (c *LiveChatSystem) publishMessage(ctx context.Context, sessionID string, message ChatMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	channel := fmt.Sprintf("chat:session:%s", sessionID)
	return c.redis.Publish(ctx, channel, string(data))
}

// Helper methods for Knowledge Base

func (e *SupportEngine) saveArticle(ctx context.Context, article *KnowledgeBaseArticle) error {
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("kb:article:%s", article.ID)
	if err := e.redis.Set(ctx, key, string(data), 0); err != nil {
		return err
	}

	// Add to category index
	if article.Category != "" {
		categoryKey := fmt.Sprintf("kb:category:%s", article.Category)
		e.redis.SAdd(ctx, categoryKey, article.ID)
	}

	// Add to search index
	for _, tag := range article.Tags {
		tagKey := fmt.Sprintf("kb:tag:%s", tag)
		e.redis.SAdd(ctx, tagKey, article.ID)
	}

	return nil
}

func (e *SupportEngine) getArticle(ctx context.Context, articleID string) (*KnowledgeBaseArticle, error) {
	key := fmt.Sprintf("kb:article:%s", articleID)
	data, err := e.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var article KnowledgeBaseArticle
	if err := json.Unmarshal([]byte(data), &article); err != nil {
		return nil, err
	}

	return &article, nil
}

func (e *SupportEngine) searchArticles(ctx context.Context, query string, category string, limit int) ([]KnowledgeBaseArticle, error) {
	// Simplified search - in production, use a proper search engine like Elasticsearch
	var articleIDs []string

	if category != "" {
		categoryKey := fmt.Sprintf("kb:category:%s", category)
		ids, err := e.redis.SMembers(ctx, categoryKey)
		if err == nil {
			articleIDs = ids
		}
	}

	var articles []KnowledgeBaseArticle
	count := 0
	for _, articleID := range articleIDs {
		if limit > 0 && count >= limit {
			break
		}

		article, err := e.getArticle(ctx, articleID)
		if err != nil {
			continue
		}

		if !article.Published {
			continue
		}

		// Simple text matching
		if query == "" || containsSubstring(article.Title, query) || containsSubstring(article.Content, query) {
			articles = append(articles, *article)
			count++
		}
	}

	return articles, nil
}

// Notification methods

func (e *SupportEngine) notifyAgents(ctx context.Context, eventType string, data interface{}) {
	notification := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	notificationData, _ := json.Marshal(notification)
	e.redis.Publish(ctx, "support:notifications", string(notificationData))
}

func (e *SupportEngine) notifyTicketUpdate(ctx context.Context, ticketID string, comment TicketComment) {
	notification := map[string]interface{}{
		"type":      "ticket_comment",
		"ticket_id": ticketID,
		"comment":   comment,
		"timestamp": time.Now().Unix(),
	}

	notificationData, _ := json.Marshal(notification)
	e.redis.Publish(ctx, fmt.Sprintf("ticket:%s:updates", ticketID), string(notificationData))
}

// ID generators

func generateTicketID() string {
	return fmt.Sprintf("TCK-%d", time.Now().UnixNano())
}

func generateCommentID() string {
	return fmt.Sprintf("CMT-%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("CHT-%d", time.Now().UnixNano())
}

func generateMessageID() string {
	return fmt.Sprintf("MSG-%d", time.Now().UnixNano())
}

func generateAgentID() string {
	return fmt.Sprintf("AGT-%d", time.Now().UnixNano())
}

func generateArticleID() string {
	return fmt.Sprintf("KBA-%d", time.Now().UnixNano())
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsInString(s, substr))
}

func containsInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
