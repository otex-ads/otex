package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func NewClient(addr string) (*Client, error) {
	// Strip redis:// prefix if present
	if len(addr) > 8 && addr[:8] == "redis://" {
		addr = addr[8:]
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{client: rdb}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Client() *redis.Client {
	return c.client
}

// Campaign candidates sorted set (zone_id -> campaign IDs ranked by eCPM)
func (c *Client) GetCampaignCandidates(ctx context.Context, zoneID string) ([]string, error) {
	key := fmt.Sprintf("campaign:%s:candidates", zoneID)
	return c.client.ZRevRange(ctx, key, 0, -1).Result()
}

func (c *Client) SetCampaignCandidates(ctx context.Context, zoneID string, campaigns map[string]float64) error {
	key := fmt.Sprintf("campaign:%s:candidates", zoneID)
	pipe := c.client.Pipeline()
	for campaignID, score := range campaigns {
		pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: campaignID})
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Campaign metadata hash
func (c *Client) GetCampaignMeta(ctx context.Context, campaignID string) (map[string]string, error) {
	key := fmt.Sprintf("campaign:%s:meta", campaignID)
	return c.client.HGetAll(ctx, key).Result()
}

func (c *Client) SetCampaignMeta(ctx context.Context, campaignID string, meta map[string]string) error {
	key := fmt.Sprintf("campaign:%s:meta", campaignID)
	return c.client.HSet(ctx, key, meta).Err()
}

// Budget counters
func (c *Client) GetCampaignSpendToday(ctx context.Context, campaignID string) (int64, error) {
	key := fmt.Sprintf("campaign:%s:spend:today", campaignID)
	val, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *Client) IncrementCampaignSpend(ctx context.Context, campaignID string, amount int64) (int64, error) {
	key := fmt.Sprintf("campaign:%s:spend:today", campaignID)
	return c.client.IncrBy(ctx, key, amount).Result()
}

func (c *Client) GetDailyBudget(ctx context.Context, campaignID string) (int64, error) {
	key := fmt.Sprintf("campaign:%s:budget:daily", campaignID)
	val, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *Client) SetDailyBudget(ctx context.Context, campaignID string, amount int64) error {
	key := fmt.Sprintf("campaign:%s:budget:daily", campaignID)
	return c.client.Set(ctx, key, amount, 0).Err()
}

func (c *Client) GetTotalBudgetRemaining(ctx context.Context, campaignID string) (int64, error) {
	key := fmt.Sprintf("campaign:%s:budget:total_remaining", campaignID)
	val, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *Client) SetTotalBudgetRemaining(ctx context.Context, campaignID string, amount int64) error {
	key := fmt.Sprintf("campaign:%s:budget:total_remaining", campaignID)
	return c.client.Set(ctx, key, amount, 0).Err()
}

// Frequency cap
func (c *Client) GetFrequencyCount(ctx context.Context, userHash, campaignID string) (int64, error) {
	key := fmt.Sprintf("freq:%s:%s", userHash, campaignID)
	val, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *Client) IncrementFrequency(ctx context.Context, userHash, campaignID string, ttl time.Duration) (int64, error) {
	key := fmt.Sprintf("freq:%s:%s", userHash, campaignID)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		c.client.Expire(ctx, key, ttl)
	}
	return count, nil
}

// Zone metadata
func (c *Client) GetZoneMeta(ctx context.Context, zoneID string) (map[string]string, error) {
	key := fmt.Sprintf("zone:%s:meta", zoneID)
	return c.client.HGetAll(ctx, key).Result()
}

func (c *Client) SetZoneMeta(ctx context.Context, zoneID string, meta map[string]string) error {
	key := fmt.Sprintf("zone:%s:meta", zoneID)
	return c.client.HSet(ctx, key, meta).Err()
}

// Fraud detection
func (c *Client) IsIPBlocked(ctx context.Context, ip string) (bool, error) {
	key := "fraud:ip_block"
	return c.client.SIsMember(ctx, key, ip).Result()
}

func (c *Client) BlockIP(ctx context.Context, ip string) error {
	key := "fraud:ip_block"
	return c.client.SAdd(ctx, key, ip).Err()
}

func (c *Client) GetRateCount(ctx context.Context, ip string) (int64, error) {
	key := fmt.Sprintf("fraud:rate:%s", ip)
	val, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *Client) IncrementRate(ctx context.Context, ip string, ttl time.Duration) (int64, error) {
	key := fmt.Sprintf("fraud:rate:%s", ip)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		c.client.Expire(ctx, key, ttl)
	}
	return count, nil
}

// Pub/Sub for campaign updates
func (c *Client) PublishCampaignUpdate(ctx context.Context, campaignID string) error {
	return c.client.Publish(ctx, "campaign_updates", campaignID).Err()
}

func (c *Client) SubscribeToCampaignUpdates(ctx context.Context) *redis.PubSub {
	return c.client.Subscribe(ctx, "campaign_updates")
}

// Redis Streams for events
func (c *Client) AddEvent(ctx context.Context, stream string, event map[string]interface{}) error {
	_, err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: event,
	}).Result()
	return err
}

func (c *Client) ReadEventsGroup(ctx context.Context, stream, group, consumer string, count int64) ([]redis.XStream, error) {
	return c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    1 * time.Second,
	}).Result()
}

func (c *Client) AckEvent(ctx context.Context, stream, group string, id string) error {
	return c.client.XAck(ctx, stream, group, id).Err()
}

func (c *Client) CreateConsumerGroup(ctx context.Context, stream, group string) error {
	// Try to create group, ignore if already exists
	err := c.client.XGroupCreate(ctx, stream, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

// Performance tracking for optimization
func (c *Client) GetCampaignCTR(ctx context.Context, campaignID string) (float64, error) {
	key := fmt.Sprintf("perf:campaign:%s:ctr", campaignID)
	val, err := c.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0.0, nil
	}
	return val, err
}

func (c *Client) UpdateCampaignCTR(ctx context.Context, campaignID string, ctr float64) error {
	key := fmt.Sprintf("perf:campaign:%s:ctr", campaignID)
	return c.client.Set(ctx, key, ctr, 24*time.Hour).Err()
}

func (c *Client) GetZoneCTR(ctx context.Context, zoneID string) (float64, error) {
	key := fmt.Sprintf("perf:zone:%s:ctr", zoneID)
	val, err := c.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0.0, nil
	}
	return val, err
}

func (c *Client) UpdateZoneCTR(ctx context.Context, zoneID string, ctr float64) error {
	key := fmt.Sprintf("perf:zone:%s:ctr", zoneID)
	return c.client.Set(ctx, key, ctr, 24*time.Hour).Err()
}

func (c *Client) GetZoneConversionRate(ctx context.Context, zoneID string) (float64, error) {
	key := fmt.Sprintf("perf:zone:%s:conv_rate", zoneID)
	val, err := c.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0.0, nil
	}
	return val, err
}

func (c *Client) UpdateZoneConversionRate(ctx context.Context, zoneID string, rate float64) error {
	key := fmt.Sprintf("perf:zone:%s:conv_rate", zoneID)
	return c.client.Set(ctx, key, rate, 24*time.Hour).Err()
}

// Generic Redis operations

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.client.Publish(ctx, channel, message).Err()
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}
