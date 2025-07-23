#!/bin/bash

echo "🌐 Testing API endpoints..."

# Set password for PostgreSQL connections
export PGPASSWORD=12345678

# Start API server in background
./apiserver &
API_PID=$!
sleep 2

echo "Testing article retrieval through SPQR:"

# Get first few article IDs from database
ARTICLE_IDS=$(psql "host=localhost user=user1 dbname=db1 port=16432" -t -c "SELECT id FROM articles LIMIT 3;" 2>/dev/null | xargs)

for id in $ARTICLE_IDS; do
    echo ""
    echo "Getting article $id:"
    curl -s "http://localhost:8080/article/$id" | jq -r '.title // "Article not found"' 2>/dev/null || echo "Could not retrieve article $id"
done

echo ""
echo "✅ API test complete"

# Clean up
kill $API_PID 2>/dev/null

# Unset password
unset PGPASSWORD
