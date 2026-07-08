# AdNet - Ad Network Platform

A high-performance ad network platform built with Go, PostgreSQL, and Redis. Inspired by PropellerAds architecture.

## Architecture

The system follows a dual-path architecture:

- **Ad Serving Path**: Ultra-low latency (<15ms p99), Redis-only reads, stateless horizontal scaling
- **Dashboard Path**: Standard CRUD operations, direct Postgres access with auth

### Services

- **adserve**: Public ad delivery endpoint (Redis-only, stateless)
- **api**: Dashboard REST API for advertisers/publishers/admins (Postgres + auth)
- **eventd**: Event consumer (impressions/clicks/conversions) → Postgres batch writes
- **reconciler**: Background worker for budget reconciliation and campaign state sync
- **billingd**: M-Pesa integration for top-ups and payouts
- **worker**: Generic cron runner (reports, emails, moderation)

## Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15 (with partitioning)
- **Cache**: Redis 7
- **Auth**: JWT (access + refresh tokens)
- **HTTP**: gorilla/mux router

## Project Structure

```
adnet/
├── cmd/                    # Service binaries
│   ├── adserve/           # Ad serving endpoint
│   ├── api/               # Dashboard API
│   ├── eventd/            # Event consumer
│   ├── reconciler/        # Budget reconciliation
│   ├── billingd/          # M-Pesa billing
│   ├── worker/            # Cron jobs
│   └── seed-redis/        # Redis seeding utility
├── internal/
│   ├── auth/              # JWT & password hashing
│   ├── campaign/          # Campaign domain logic
│   ├── budget/            # Budget pacing logic
│   ├── freqcap/           # Frequency capping
│   ├── fraud/             # Anti-bot detection
│   ├── billing/           # M-Pesa Daraja client
│   ├── events/            # Event schemas
│   ├── store/
│   │   ├── postgres/      # Postgres queries
│   │   └── redis/         # Redis client wrappers
│   ├── geo/               # IP → geo lookup
│   └── mw/                # HTTP middleware
├── pkg/
│   └── httpx/             # Shared HTTP helpers
├── migrations/            # SQL migrations
└── deploy/
    ├── docker-compose.yml
    ├── Dockerfile.api
    └── Dockerfile.adserve
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Using Docker Compose

1. Start the infrastructure:
```bash
cd deploy
docker-compose up -d
```

This starts:
- PostgreSQL on port 5432
- Redis on port 6379
- API service on port 8080
- Adserve service on port 8081

2. Seed Redis with test data:
```bash
cd ..
go run cmd/seed-redis/main.go
```

3. Test the ad serving endpoint:
```bash
curl 'http://localhost:8081/serve?zone=550e8400-e29b-41d4-a716-446655440000&format=banner'
```

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Start PostgreSQL and Redis (or use docker-compose for just those):
```bash
cd deploy
docker-compose up postgres redis -d
```

3. Run migrations:
```bash
psql postgres://postgres:postgres@localhost:5432/adnet -f migrations/000001_init.up.sql
```

4. Run the API service:
```bash
POSTGRES_URL=postgres://postgres:postgres@localhost:5432/adnet?sslmode=disable \
REDIS_URL=localhost:6379 \
JWT_SECRET=dev-secret \
PORT=8080 \
go run cmd/api/main.go
```

5. Run the adserve service:
```bash
REDIS_URL=localhost:6379 \
JWT_SECRET=dev-secret \
PORT=8081 \
go run cmd/adserve/main.go
```

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register new account
- `POST /api/auth/login` - Login and get tokens
- `POST /api/auth/refresh` - Refresh access token

### Campaigns (Advertiser)

- `GET /api/v1/campaigns` - List campaigns
- `POST /api/v1/campaigns` - Create campaign
- `GET /api/v1/campaigns/:id` - Get campaign details
- `PATCH /api/v1/campaigns/:id` - Update campaign

### Zones (Publisher)

- `GET /api/v1/sites` - List sites
- `POST /api/v1/sites` - Create site
- `POST /api/v1/sites/:id/zones` - Create zone
- `GET /api/v1/zones/:id` - Get zone details

### Ad Serving

- `GET /serve?zone={zone_id}&format={format}` - Get ad for display
- `GET /click?token={token}` - Handle ad click

## Redis Key Schema

- `campaign:{zone_id}:candidates` - Sorted set of campaign IDs by eCPM
- `campaign:{id}:meta` - Hash with campaign metadata
- `campaign:{id}:spend:today` - Daily spend counter
- `campaign:{id}:budget:daily` - Daily budget limit
- `campaign:{id}:budget:total_remaining` - Total remaining budget
- `freq:{user_hash}:{campaign_id}` - Frequency cap counter
- `zone:{id}:meta` - Zone metadata
- `fraud:ip_block` - Set of blocked IPs
- `fraud:rate:{ip}` - Rate limit counter

## Postgres Schema

Key tables:
- `accounts` - User accounts (advertisers, publishers, admins)
- `campaigns` - Ad campaigns
- `creatives` - Ad creatives
- `targeting_rules` - Campaign targeting rules
- `sites` - Publisher sites
- `zones` - Ad zones on sites
- `impressions` - Impression events (partitioned)
- `clicks` - Click events (partitioned)
- `conversions` - Conversion events (partitioned)
- `wallets` - Account wallets
- `transactions` - Financial transactions
- `payout_requests` - Publisher payout requests

## Development Roadmap

### Phase 1 - Foundation ✅
- [x] Project structure
- [x] Postgres schema
- [x] Auth system (JWT)
- [x] API service (campaigns, zones)
- [x] Basic adserve service
- [x] Redis client wrappers
- [x] Docker compose setup

### Phase 2 - Real Pipeline ✅
- [x] Redis Streams wiring (adserve → eventd → Postgres)
- [x] Campaign update propagation (pub/sub)
- [x] Budget counters + daily reset
- [x] Reconciler service

### Phase 3 - Money ✅
- [x] Wallet system
- [x] M-Pesa top-up flow
- [x] Spend deduction
- [x] Payout requests + B2C disbursement

### Phase 4 - Scale & Safety ✅
- [x] Additional ad formats (native, popunder, in-page push, interstitial, push)
- [x] Inline fraud checks
- [x] Creative moderation queue
- [x] Frequency capping

### Phase 5 - Polish
- [ ] Frontend integration (designs ready, to be cloned)
- [ ] Async fraud analysis
- [ ] Reporting/analytics endpoints
- [ ] Dayparting
- [ ] Admin dashboards

## License

MIT
