package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Site struct {
	ID          uuid.UUID `json:"id"`
	PublisherID uuid.UUID `json:"publisher_id"`
	Domain      string    `json:"domain"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Zone struct {
	ID             uuid.UUID `json:"id"`
	SiteID         uuid.UUID `json:"site_id"`
	Name           string    `json:"name"`
	Format         string    `json:"format"`
	FloorPriceCents int      `json:"floor_price_cents"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

func (db *DB) CreateSite(ctx context.Context, publisherID uuid.UUID, domain, status string) (*Site, error) {
	const query = `
		INSERT INTO sites (publisher_id, domain, status)
		VALUES ($1, $2, $3)
		RETURNING id, publisher_id, domain, status, created_at
	`

	var site Site
	err := db.pool.QueryRow(ctx, query, publisherID, domain, status).Scan(
		&site.ID,
		&site.PublisherID,
		&site.Domain,
		&site.Status,
		&site.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &site, nil
}

func (db *DB) GetSiteByID(ctx context.Context, id uuid.UUID) (*Site, error) {
	const query = `
		SELECT id, publisher_id, domain, status, created_at
		FROM sites
		WHERE id = $1
	`

	var site Site
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&site.ID,
		&site.PublisherID,
		&site.Domain,
		&site.Status,
		&site.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &site, nil
}

func (db *DB) ListSitesByPublisher(ctx context.Context, publisherID uuid.UUID) ([]*Site, error) {
	const query = `
		SELECT id, publisher_id, domain, status, created_at
		FROM sites
		WHERE publisher_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.pool.Query(ctx, query, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []*Site
	for rows.Next() {
		var site Site
		if err := rows.Scan(
			&site.ID,
			&site.PublisherID,
			&site.Domain,
			&site.Status,
			&site.CreatedAt,
		); err != nil {
			return nil, err
		}
		sites = append(sites, &site)
	}

	return sites, nil
}

func (db *DB) UpdateSiteStatus(ctx context.Context, id uuid.UUID, status string) error {
	const query = `UPDATE sites SET status = $2 WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, id, status)
	return err
}

func (db *DB) CreateZone(ctx context.Context, siteID uuid.UUID, name, format string, floorPriceCents int, status string) (*Zone, error) {
	const query = `
		INSERT INTO zones (site_id, name, format, floor_price_cents, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, site_id, name, format, floor_price_cents, status, created_at
	`

	var zone Zone
	err := db.pool.QueryRow(ctx, query, siteID, name, format, floorPriceCents, status).Scan(
		&zone.ID,
		&zone.SiteID,
		&zone.Name,
		&zone.Format,
		&zone.FloorPriceCents,
		&zone.Status,
		&zone.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}

func (db *DB) GetZoneByID(ctx context.Context, id uuid.UUID) (*Zone, error) {
	const query = `
		SELECT id, site_id, name, format, floor_price_cents, status, created_at
		FROM zones
		WHERE id = $1
	`

	var zone Zone
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&zone.ID,
		&zone.SiteID,
		&zone.Name,
		&zone.Format,
		&zone.FloorPriceCents,
		&zone.Status,
		&zone.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &zone, nil
}

func (db *DB) ListZonesBySite(ctx context.Context, siteID uuid.UUID) ([]*Zone, error) {
	const query = `
		SELECT id, site_id, name, format, floor_price_cents, status, created_at
		FROM zones
		WHERE site_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.pool.Query(ctx, query, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*Zone
	for rows.Next() {
		var zone Zone
		if err := rows.Scan(
			&zone.ID,
			&zone.SiteID,
			&zone.Name,
			&zone.Format,
			&zone.FloorPriceCents,
			&zone.Status,
			&zone.CreatedAt,
		); err != nil {
			return nil, err
		}
		zones = append(zones, &zone)
	}

	return zones, nil
}

func (db *DB) UpdateZone(ctx context.Context, id uuid.UUID, name, format string, floorPriceCents int, status string) (*Zone, error) {
	const query = `
		UPDATE zones
		SET name = $2, format = $3, floor_price_cents = $4, status = $5
		WHERE id = $1
		RETURNING id, site_id, name, format, floor_price_cents, status, created_at
	`

	var zone Zone
	err := db.pool.QueryRow(ctx, query, id, name, format, floorPriceCents, status).Scan(
		&zone.ID,
		&zone.SiteID,
		&zone.Name,
		&zone.Format,
		&zone.FloorPriceCents,
		&zone.Status,
		&zone.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}
