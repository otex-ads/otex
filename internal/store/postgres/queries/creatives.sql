-- name: CreateCreative :one
INSERT INTO creatives (campaign_id, format, title, body, icon_url, image_url, click_url, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetCreativeByID :one
SELECT * FROM creatives
WHERE id = $1;

-- name: ListCreativesByCampaign :many
SELECT * FROM creatives
WHERE campaign_id = $1
ORDER BY created_at DESC;

-- name: UpdateCreativeStatus :one
UPDATE creatives
SET status = $2
WHERE id = $1
RETURNING *;

-- name: ListPendingCreatives :many
SELECT c.*, a.company_name as advertiser_name
FROM creatives c
JOIN campaigns cam ON c.campaign_id = cam.id
JOIN accounts a ON cam.advertiser_id = a.id
WHERE c.status = 'pending_review'
ORDER BY c.created_at ASC;
