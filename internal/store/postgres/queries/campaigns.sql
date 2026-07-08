-- name: CreateCampaign :one
INSERT INTO campaigns (advertiser_id, name, status, pricing_model, bid_amount_cents, daily_budget_cents, total_budget_cents, timezone, starts_at, ends_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetCampaignByID :one
SELECT * FROM campaigns
WHERE id = $1;

-- name: ListCampaignsByAdvertiser :many
SELECT * FROM campaigns
WHERE advertiser_id = $1
ORDER BY created_at DESC;

-- name: UpdateCampaign :one
UPDATE campaigns
SET name = $2,
    status = $3,
    pricing_model = $4,
    bid_amount_cents = $5,
    daily_budget_cents = $6,
    total_budget_cents = $7,
    timezone = $8,
    starts_at = $9,
    ends_at = $10
WHERE id = $11
RETURNING *;

-- name: UpdateCampaignStatus :one
UPDATE campaigns
SET status = $2
WHERE id = $1
RETURNING *;

-- name: ListActiveCampaigns :many
SELECT * FROM campaigns
WHERE status = 'active'
  AND (starts_at IS NULL OR starts_at <= NOW())
  AND (ends_at IS NULL OR ends_at > NOW());
