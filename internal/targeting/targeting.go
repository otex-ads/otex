package targeting

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"adnet/internal/store/redis"
)

type TargetingEngine struct {
	redis *redis.Client
}

type UserProfile struct {
	UserID        string                 `json:"user_id"`
	Interests     []string               `json:"interests"`
	Demographics  map[string]interface{} `json:"demographics"`
	Behavioral    BehavioralData         `json:"behavioral"`
	Contextual    ContextualData         `json:"contextual"`
	LookalikeScore float64              `json:"lookalike_score"`
	LastUpdated  time.Time              `json:"last_updated"`
}

type BehavioralData struct {
	PageViews        map[string]int64    `json:"page_views"`        // URL -> count
	TimeOnPage       map[string]int64    `json:"time_on_page"`       // URL -> seconds
	ClickHistory     []ClickEvent        `json:"click_history"`
	ConversionEvents []ConversionEvent  `json:"conversion_events"`
	SessionCount     int64               `json:"session_count"`
	LastSession     time.Time           `json:"last_session"`
}

type ContextualData struct {
	CurrentPage     string   `json:"current_page"`
	PageCategory    string   `json:"page_category"`
	Keywords        []string `json:"keywords"`
	PageLanguage    string   `json:"page_language"`
	ContentTags     []string `json:"content_tags"`
}

type ClickEvent struct {
	CampaignID  string    `json:"campaign_id"`
	CreativeID  string    `json:"creative_id"`
	Timestamp   time.Time `json:"timestamp"`
	Position    string    `json:"position"`
}

type ConversionEvent struct {
	CampaignID   string    `json:"campaign_id"`
	ConversionType string  `json:"conversion_type"`
	Value        float64   `json:"value"`
	Timestamp    time.Time `json:"timestamp"`
}

type TargetingCriteria struct {
	CampaignID      string                 `json:"campaign_id"`
	Geographic      GeographicTargeting    `json:"geographic"`
	Demographic     DemographicTargeting   `json:"demographic"`
	Behavioral      BehavioralTargeting   `json:"behavioral"`
	Contextual      ContextualTargeting   `json:"contextual"`
	Lookalike       LookalikeTargeting    `json:"lookalike"`
}

type GeographicTargeting struct {
	Countries       []string `json:"countries"`
	Regions         []string `json:"regions"`
	Cities          []string `json:"cities"`
	ExcludeLocations []string `json:"exclude_locations"`
}

type DemographicTargeting struct {
	AgeRange        [2]int    `json:"age_range"`        // [min, max]
	Gender          []string  `json:"gender"`
	IncomeLevel     []string  `json:"income_level"`
	Education       []string  `json:"education"`
	Interests       []string  `json:"interests"`
}

type BehavioralTargeting struct {
	RequiredInterests    []string `json:"required_interests"`
	ExcludedInterests    []string `json:"excluded_interests"`
	MinSessionCount      int      `json:"min_session_count"`
	RecencyDays          int      `json:"recency_days"`
	PreviousConversions  bool     `json:"previous_conversions"`
}

type ContextualTargeting struct {
	RequiredCategories  []string `json:"required_categories"`
	ExcludedCategories  []string `json:"excluded_categories"`
	RequiredKeywords    []string `json:"required_keywords"`
	ExcludedKeywords    []string `json:"excluded_keywords"`
	PageLanguage        []string `json:"page_language"`
}

type LookalikeTargeting struct {
	SeedAudienceIDs     []string `json:"seed_audience_ids"`
	SimilarityThreshold float64  `json:"similarity_threshold"`
	Percentage          int      `json:"percentage"` // Top X% of lookalike users
}

type MatchResult struct {
	Matched      bool    `json:"matched"`
	Score        float64 `json:"score"`
	Reasons      []string `json:"reasons"`
}

func NewTargetingEngine(redisClient *redis.Client) *TargetingEngine {
	return &TargetingEngine{
		redis: redisClient,
	}
}

// MatchUserToCampaign checks if a user matches a campaign's targeting criteria
func (e *TargetingEngine) MatchUserToCampaign(ctx context.Context, userHash string, criteria TargetingCriteria, pageContext ContextualData) (*MatchResult, error) {
	result := &MatchResult{
		Matched: true,
		Score:   0,
		Reasons: []string{},
	}

	// Get user profile from Redis
	profile, err := e.getUserProfile(ctx, userHash)
	if err != nil {
		// If no profile exists, use basic contextual matching only
		profile = &UserProfile{
			UserID:       userHash,
			Behavioral:  BehavioralData{},
			Contextual:  pageContext,
			LastUpdated: time.Now(),
		}
	}

	// Update contextual data with current page
	profile.Contextual = pageContext

	// Geographic targeting
	if len(criteria.Geographic.Countries) > 0 {
		userCountry := profile.Demographics["country"].(string)
		if !contains(criteria.Geographic.Countries, userCountry) {
			result.Matched = false
			result.Reasons = append(result.Reasons, "Geographic mismatch")
			return result, nil
		}
		result.Score += 0.2
	}

	// Demographic targeting
	if criteria.Demographic.AgeRange[1] > 0 {
		userAge, ok := profile.Demographics["age"].(int)
		if !ok || userAge < criteria.Demographic.AgeRange[0] || userAge > criteria.Demographic.AgeRange[1] {
			result.Matched = false
			result.Reasons = append(result.Reasons, "Age mismatch")
			return result, nil
		}
		result.Score += 0.15
	}

	if len(criteria.Demographic.Gender) > 0 {
		userGender, ok := profile.Demographics["gender"].(string)
		if !ok || !contains(criteria.Demographic.Gender, userGender) {
			result.Matched = false
			result.Reasons = append(result.Reasons, "Gender mismatch")
			return result, nil
		}
		result.Score += 0.1
	}

	// Interest targeting
	if len(criteria.Demographic.Interests) > 0 {
		matchedInterests := 0
		for _, interest := range criteria.Demographic.Interests {
			if contains(profile.Interests, interest) {
				matchedInterests++
			}
		}
		if matchedInterests == 0 {
			result.Matched = false
			result.Reasons = append(result.Reasons, "No matching interests")
			return result, nil
		}
		result.Score += float64(matchedInterests) / float64(len(criteria.Demographic.Interests)) * 0.2
	}

	// Behavioral targeting
	if len(criteria.Behavioral.RequiredInterests) > 0 {
		matchedBehavioral := 0
		for _, interest := range criteria.Behavioral.RequiredInterests {
			if contains(profile.Interests, interest) {
				matchedBehavioral++
			}
		}
		if matchedBehavioral == 0 {
			result.Matched = false
			result.Reasons = append(result.Reasons, "Behavioral interest mismatch")
			return result, nil
		}
		result.Score += float64(matchedBehavioral) / float64(len(criteria.Behavioral.RequiredInterests)) * 0.15
	}

	if criteria.Behavioral.MinSessionCount > 0 {
		if profile.Behavioral.SessionCount < int64(criteria.Behavioral.MinSessionCount) {
			result.Matched = false
			result.Reasons = append(result.Reasons, "Insufficient session count")
			return result, nil
		}
		result.Score += 0.1
	}

	if criteria.Behavioral.RecencyDays > 0 {
		daysSinceLastSession := int(time.Since(profile.Behavioral.LastSession).Hours() / 24)
		if daysSinceLastSession > criteria.Behavioral.RecencyDays {
			result.Matched = false
			result.Reasons = append(result.Reasons, "User inactive for too long")
			return result, nil
		}
		result.Score += 0.1
	}

	// Contextual targeting
	if len(criteria.Contextual.RequiredCategories) > 0 {
		if !contains(criteria.Contextual.RequiredCategories, pageContext.PageCategory) {
			result.Matched = false
			result.Reasons = append(result.Reasons, "Contextual category mismatch")
			return result, nil
		}
		result.Score += 0.15
	}

	if len(criteria.Contextual.RequiredKeywords) > 0 {
		matchedKeywords := 0
		for _, keyword := range criteria.Contextual.RequiredKeywords {
			if contains(pageContext.Keywords, keyword) {
				matchedKeywords++
			}
		}
		if matchedKeywords == 0 {
			result.Matched = false
			result.Reasons = append(result.Reasons, "No matching keywords")
			return result, nil
		}
		result.Score += float64(matchedKeywords) / float64(len(criteria.Contextual.RequiredKeywords)) * 0.1
	}

	// Lookalike targeting
	if len(criteria.Lookalike.SeedAudienceIDs) > 0 {
		if profile.LookalikeScore < criteria.Lookalike.SimilarityThreshold {
			result.Matched = false
			result.Reasons = append(result.Reasons, "Lookalike score below threshold")
			return result, nil
		}
		result.Score += profile.LookalikeScore * 0.2
	}

	return result, nil
}

// UpdateUserProfile updates a user's profile based on their behavior
func (e *TargetingEngine) UpdateUserProfile(ctx context.Context, userHash string, event map[string]interface{}) error {
	profile, err := e.getUserProfile(ctx, userHash)
	if err != nil {
		profile = &UserProfile{
			UserID:       userHash,
			Interests:    []string{},
			Demographics: make(map[string]interface{}),
			Behavioral: BehavioralData{
				PageViews:        make(map[string]int64),
				TimeOnPage:       make(map[string]int64),
				ClickHistory:     []ClickEvent{},
				ConversionEvents: []ConversionEvent{},
			},
			Contextual:    ContextualData{},
			LastUpdated:  time.Now(),
		}
	}

	eventType := event["type"].(string)

	switch eventType {
	case "page_view":
		url := event["url"].(string)
		profile.Behavioral.PageViews[url]++
		if duration, ok := event["duration"].(int64); ok {
			profile.Behavioral.TimeOnPage[url] += duration
		}
		// Extract interests from page content
		if keywords, ok := event["keywords"].([]string); ok {
			for _, keyword := range keywords {
				if !contains(profile.Interests, keyword) {
					profile.Interests = append(profile.Interests, keyword)
				}
			}
		}

	case "click":
		click := ClickEvent{
			CampaignID: event["campaign_id"].(string),
			CreativeID: event["creative_id"].(string),
			Timestamp:  time.Now(),
			Position:   event["position"].(string),
		}
		profile.Behavioral.ClickHistory = append(profile.Behavioral.ClickHistory, click)
		// Keep only last 100 clicks
		if len(profile.Behavioral.ClickHistory) > 100 {
			profile.Behavioral.ClickHistory = profile.Behavioral.ClickHistory[len(profile.Behavioral.ClickHistory)-100:]
		}

	case "conversion":
		conversion := ConversionEvent{
			CampaignID:       event["campaign_id"].(string),
			ConversionType:   event["conversion_type"].(string),
			Value:            event["value"].(float64),
			Timestamp:        time.Now(),
		}
		profile.Behavioral.ConversionEvents = append(profile.Behavioral.ConversionEvents, conversion)

	case "session_start":
		profile.Behavioral.SessionCount++
		profile.Behavioral.LastSession = time.Now()
	}

	profile.LastUpdated = time.Now()

	// Save to Redis
	return e.saveUserProfile(ctx, userHash, profile)
}

// CalculateLookalikeScore calculates similarity score for lookalike targeting
func (e *TargetingEngine) CalculateLookalikeScore(ctx context.Context, userHash string, seedAudience []string) (float64, error) {
	profile, err := e.getUserProfile(ctx, userHash)
	if err != nil {
		return 0, err
	}

	var totalSimilarity float64
	var count int

	for _, seedID := range seedAudience {
		seedProfile, err := e.getUserProfile(ctx, seedID)
		if err != nil {
			continue
		}

		similarity := e.calculateSimilarity(profile, seedProfile)
		totalSimilarity += similarity
		count++
	}

	if count == 0 {
		return 0, nil
	}

	return totalSimilarity / float64(count), nil
}

// calculateSimilarity computes similarity between two user profiles
func (e *TargetingEngine) calculateSimilarity(p1, p2 *UserProfile) float64 {
	var score float64

	// Interest overlap (40% weight)
	interestOverlap := e.calculateJaccardSimilarity(p1.Interests, p2.Interests)
	score += interestOverlap * 0.4

	// Behavioral similarity (30% weight)
	behavioralScore := e.calculateBehavioralSimilarity(p1, p2)
	score += behavioralScore * 0.3

	// Demographic similarity (30% weight)
	demographicScore := e.calculateDemographicSimilarity(p1, p2)
	score += demographicScore * 0.3

	return score
}

func (e *TargetingEngine) calculateJaccardSimilarity(s1, s2 []string) float64 {
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, item := range s1 {
		set1[item] = true
	}
	for _, item := range s2 {
		set2[item] = true
	}

	intersection := 0
	for item := range set1 {
		if set2[item] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func (e *TargetingEngine) calculateBehavioralSimilarity(p1, p2 *UserProfile) float64 {
	// Compare click patterns
	clickSimilarity := e.calculateJaccardSimilarity(
		e.extractCampaignIDs(p1.Behavioral.ClickHistory),
		e.extractCampaignIDs(p2.Behavioral.ClickHistory),
	)

	// Compare conversion patterns
	convSimilarity := e.calculateJaccardSimilarity(
		e.extractConversionTypes(p1.Behavioral.ConversionEvents),
		e.extractConversionTypes(p2.Behavioral.ConversionEvents),
	)

	return (clickSimilarity + convSimilarity) / 2
}

func (e *TargetingEngine) calculateDemographicSimilarity(p1, p2 *UserProfile) float64 {
	var score float64
	var factors int

	// Age similarity
	if age1, ok1 := p1.Demographics["age"].(int); ok1 {
		if age2, ok2 := p2.Demographics["age"].(int); ok2 {
			ageDiff := math.Abs(float64(age1 - age2))
			ageSimilarity := math.Max(0, 1-ageDiff/50) // Normalize to 0-1
			score += ageSimilarity
			factors++
		}
	}

	// Gender match
	if gender1, ok1 := p1.Demographics["gender"].(string); ok1 {
		if gender2, ok2 := p2.Demographics["gender"].(string); ok2 {
			if gender1 == gender2 {
				score += 1
			}
			factors++
		}
	}

	// Location match
	if country1, ok1 := p1.Demographics["country"].(string); ok1 {
		if country2, ok2 := p2.Demographics["country"].(string); ok2 {
			if country1 == country2 {
				score += 1
			}
			factors++
		}
	}

	if factors == 0 {
		return 0
	}

	return score / float64(factors)
}

func (e *TargetingEngine) extractCampaignIDs(clicks []ClickEvent) []string {
	ids := make([]string, len(clicks))
	for i, click := range clicks {
		ids[i] = click.CampaignID
	}
	return ids
}

func (e *TargetingEngine) extractConversionTypes(conversions []ConversionEvent) []string {
	types := make([]string, len(conversions))
	for i, conv := range conversions {
		types[i] = conv.ConversionType
	}
	return types
}

func (e *TargetingEngine) getUserProfile(ctx context.Context, userHash string) (*UserProfile, error) {
	data, err := e.redis.Get(ctx, fmt.Sprintf("profile:%s", userHash))
	if err != nil {
		return nil, err
	}

	var profile UserProfile
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (e *TargetingEngine) saveUserProfile(ctx context.Context, userHash string, profile *UserProfile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	// Cache for 30 days
	return e.redis.Set(ctx, fmt.Sprintf("profile:%s", userHash), string(data), 30*24*time.Hour)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
