#!/bin/bash
# Fix Paystack secret key in .env and restart API

cd /opt/adnet/deploy

# Remove all PAYSTACK_SECRET_KEY lines
sed -i '/PAYSTACK_SECRET_KEY/d' .env

# Add the correct Paystack secret key
echo "PAYSTACK_SECRET_KEY=sk_test_a574bc3226c33b06404a8f35bf3e663b6f17d961" >> .env

# Restart the API service
docker compose restart api

echo "Paystack secret key updated and API restarted"
