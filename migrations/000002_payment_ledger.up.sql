-- PAYMENT & LEDGER INFRASTRUCTURE

-- Transfer Recipients (publisher payout details stored from Paystack)
CREATE TABLE IF NOT EXISTS transfer_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    recipient_code TEXT NOT NULL,           -- RCP_xxxx from Paystack
    type TEXT NOT NULL DEFAULT 'mobile_money', -- mobile_money, bank
    name TEXT NOT NULL,
    phone TEXT,
    email TEXT,
    bank_code TEXT,                         -- e.g. MPESA
    currency TEXT NOT NULL DEFAULT 'KES',
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transfer_recipients_account ON transfer_recipients(account_id);

-- Add Paystack-specific fields to transactions table
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'completed';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS paystack_reference TEXT;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS metadata JSONB;

CREATE INDEX idx_transactions_reference ON transactions(reference);
CREATE INDEX idx_transactions_paystack_ref ON transactions(paystack_reference);
CREATE INDEX idx_transactions_status ON transactions(status);

-- Add Paystack fields to payout_requests table
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS paystack_transfer_code TEXT;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS paystack_reference TEXT;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS recipient_code TEXT;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS failure_reason TEXT;

CREATE INDEX idx_payout_requests_status ON payout_requests(status);
CREATE INDEX idx_payout_requests_paystack_ref ON payout_requests(paystack_reference);

-- Paystack webhook events log (idempotency & audit)
CREATE TABLE IF NOT EXISTS paystack_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,              -- charge.success, transfer.success, etc.
    paystack_id TEXT,                      -- Paystack event ID for dedup
    reference TEXT,
    amount_cents BIGINT,
    currency TEXT DEFAULT 'KES',
    status TEXT,
    raw_payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_paystack_events_dedup ON paystack_events(paystack_id) WHERE paystack_id IS NOT NULL;
CREATE INDEX idx_paystack_events_type ON paystack_events(event_type);
CREATE INDEX idx_paystack_events_reference ON paystack_events(reference);

-- Revenue ledger for publisher earnings from ad events
CREATE TABLE IF NOT EXISTS revenue_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publisher_id UUID NOT NULL REFERENCES accounts(id),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    zone_id UUID NOT NULL,
    event_type TEXT NOT NULL CHECK (event_type IN ('impression', 'click', 'conversion')),
    gross_amount_cents BIGINT NOT NULL,     -- total cost from advertiser
    publisher_share_cents BIGINT NOT NULL,  -- 70% to publisher
    platform_fee_cents BIGINT NOT NULL,     -- 30% platform fee
    settled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_revenue_ledger_publisher ON revenue_ledger(publisher_id, settled);
CREATE INDEX idx_revenue_ledger_campaign ON revenue_ledger(campaign_id);
CREATE INDEX idx_revenue_ledger_created ON revenue_ledger(created_at);
