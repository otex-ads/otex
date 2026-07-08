-- ACCOUNTS
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL CHECK (type IN ('advertiser','publisher','admin')),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    company_name TEXT,
    status TEXT NOT NULL DEFAULT 'active', -- active, suspended, pending_review
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    key_hash TEXT NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

-- ADVERTISER SIDE
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES accounts(id),
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft', -- draft, active, paused, exhausted, completed, rejected
    pricing_model TEXT NOT NULL CHECK (pricing_model IN ('cpc','cpm','cpa')),
    bid_amount_cents INT NOT NULL,
    daily_budget_cents INT NOT NULL,
    total_budget_cents INT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'Africa/Nairobi',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE creatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    format TEXT NOT NULL CHECK (format IN ('push','popunder','native','banner','interstitial','in_page_push')),
    title TEXT,
    body TEXT,
    icon_url TEXT,
    image_url TEXT,
    click_url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending_review', -- pending_review, approved, rejected
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE targeting_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    countries TEXT[],          -- ISO codes, empty = all
    device_types TEXT[],       -- mobile, desktop, tablet
    os TEXT[],                 -- android, ios, windows, macos, linux
    browsers TEXT[],
    carriers TEXT[],
    connection_types TEXT[],   -- wifi, cellular
    zone_ids UUID[],           -- restrict to specific publisher zones, empty = marketplace-wide
    exclude_zone_ids UUID[],
    dayparting JSONB,          -- {"mon":[["08:00","20:00"]], ...}
    frequency_cap_per_user INT DEFAULT 0 -- 0 = unlimited
);

-- PUBLISHER SIDE
CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publisher_id UUID NOT NULL REFERENCES accounts(id),
    domain TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending_review',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    name TEXT NOT NULL,
    format TEXT NOT NULL CHECK (format IN ('push','popunder','native','banner','interstitial','in_page_push')),
    floor_price_cents INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- EVENTS (high volume — partitioned)
CREATE TABLE impressions (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    campaign_id UUID NOT NULL,
    zone_id UUID NOT NULL,
    creative_id UUID NOT NULL,
    user_hash TEXT NOT NULL,       -- hashed device/IP fingerprint, not raw PII
    country TEXT,
    device_type TEXT,
    cost_cents INT NOT NULL,
    is_fraud BOOLEAN NOT NULL DEFAULT false,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY RANGE (occurred_at);

CREATE TABLE clicks (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    campaign_id UUID NOT NULL,
    zone_id UUID NOT NULL,
    creative_id UUID NOT NULL,
    user_hash TEXT NOT NULL,
    cost_cents INT NOT NULL,
    is_fraud BOOLEAN NOT NULL DEFAULT false,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY RANGE (occurred_at);

CREATE TABLE conversions (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    campaign_id UUID NOT NULL,
    click_id BIGINT,
    payout_cents INT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY RANGE (occurred_at);

-- Create initial partitions
CREATE TABLE impressions_2026_w27 PARTITION OF impressions
    FOR VALUES FROM ('2026-07-02') TO ('2026-07-09');
CREATE TABLE impressions_2026_w28 PARTITION OF impressions
    FOR VALUES FROM ('2026-07-09') TO ('2026-07-16');

CREATE TABLE clicks_2026_w27 PARTITION OF clicks
    FOR VALUES FROM ('2026-07-02') TO ('2026-07-09');
CREATE TABLE clicks_2026_w28 PARTITION OF clicks
    FOR VALUES FROM ('2026-07-09') TO ('2026-07-16');

CREATE TABLE conversions_2026_w27 PARTITION OF conversions
    FOR VALUES FROM ('2026-07-02') TO ('2026-07-09');
CREATE TABLE conversions_2026_w28 PARTITION OF conversions
    FOR VALUES FROM ('2026-07-09') TO ('2026-07-16');

-- BILLING
CREATE TABLE wallets (
    account_id UUID PRIMARY KEY REFERENCES accounts(id),
    balance_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'KES'
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    type TEXT NOT NULL CHECK (type IN ('topup','spend','payout','refund','adjustment')),
    amount_cents BIGINT NOT NULL,
    reference TEXT,             -- M-Pesa transaction code etc.
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payout_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publisher_id UUID NOT NULL REFERENCES accounts(id),
    amount_cents BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, processing, paid, failed
    mpesa_receipt TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);

-- INDEXES
CREATE INDEX idx_campaigns_status_advertiser ON campaigns(status, advertiser_id);
CREATE INDEX idx_creatives_campaign_status ON creatives(campaign_id, status);
CREATE INDEX idx_zones_site ON zones(site_id);
CREATE INDEX idx_impressions_campaign_time ON impressions(campaign_id, occurred_at);
CREATE INDEX idx_impressions_zone_time ON impressions(zone_id, occurred_at);
CREATE INDEX idx_clicks_campaign_time ON clicks(campaign_id, occurred_at);
CREATE INDEX idx_clicks_zone_time ON clicks(zone_id, occurred_at);
CREATE INDEX idx_targeting_campaign ON targeting_rules(campaign_id);
