package events

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTypeImpression EventType = "impression"
	EventTypeClick      EventType = "click"
	EventTypeConversion EventType = "conversion"
)

type Event struct {
	Type        EventType          `json:"type"`
	CampaignID  string             `json:"campaign_id"`
	ZoneID      string             `json:"zone_id"`
	CreativeID  string             `json:"creative_id"`
	UserHash    string             `json:"user_hash"`
	Country     string             `json:"country,omitempty"`
	DeviceType  string             `json:"device_type,omitempty"`
	CostCents   int64              `json:"cost_cents"`
	IP          string             `json:"ip"`
	UserAgent   string             `json:"user_agent"`
	Timestamp   time.Time          `json:"timestamp"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

type ImpressionEvent struct {
	Event
}

type ClickEvent struct {
	Event
	ClickToken string `json:"click_token"`
}

type ConversionEvent struct {
	Event
	ClickID     int64  `json:"click_id"`
	PayoutCents int64  `json:"payout_cents"`
	PostbackID  string `json:"postback_id"`
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func EventFromJSON(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
