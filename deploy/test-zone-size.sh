#!/bin/bash
TOKEN=$(curl -s -X POST https://api.otexads.com/api/auth/login -H 'Content-Type: application/json' --data @/tmp/login2.json | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
echo "Token length: ${#TOKEN}"
echo "--- Zone with size field ---"
curl -s -X POST https://api.otexads.com/api/v1/zones -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" --data @/tmp/zone-with-size.json
echo ""
