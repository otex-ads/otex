#!/bin/bash
API=https://api.otexads.com
EMAIL="pub-size-test-$(date +%s)@test.com"

# Register publisher
REG=$(curl -s -X POST $API/api/auth/register -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"Test1234!\",\"account_type\":\"publisher\",\"company_name\":\"Debug Pub\"}")
TOKEN=$(echo "$REG" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
echo "Publisher token length: ${#TOKEN}"

# Create site
echo "--- Create site ---"
SITE=$(curl -s -X POST $API/api/v1/sites -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" -d '{"domain":"debug-site-size.com"}')
echo "$SITE"
SITE_ID=$(echo "$SITE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Site ID: $SITE_ID"

# Create zone WITH size field (what the UI sends)
echo "--- Create zone with size field ---"
curl -s -X POST $API/api/v1/zones -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"name\":\"Debug Zone With Size\",\"siteId\":\"$SITE_ID\",\"format\":\"banner\",\"size\":\"300x250\",\"floor_price_cents\":0}"
echo ""
