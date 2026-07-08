-- name: CreateSite :one
INSERT INTO sites (publisher_id, domain, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSiteByID :one
SELECT * FROM sites
WHERE id = $1;

-- name: ListSitesByPublisher :many
SELECT * FROM sites
WHERE publisher_id = $1
ORDER BY created_at DESC;

-- name: UpdateSiteStatus :one
UPDATE sites
SET status = $2
WHERE id = $1
RETURNING *;

-- name: CreateZone :one
INSERT INTO zones (site_id, name, format, floor_price_cents, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetZoneByID :one
SELECT * FROM zones
WHERE id = $1;

-- name: ListZonesBySite :many
SELECT * FROM zones
WHERE site_id = $1
ORDER BY created_at DESC;

-- name: UpdateZone :one
UPDATE zones
SET name = $2,
    format = $3,
    floor_price_cents = $4,
    status = $5
WHERE id = $6
RETURNING *;
