#!/bin/bash

echo "📊 Data Distribution Across Shards:"
echo "=================================="

# Set password for PostgreSQL connections
export PGPASSWORD=12345678

SHARD1_COUNT=$(psql "host=localhost port=5550 user=user1 dbname=db1" -t -c 'SELECT COUNT(*) FROM articles;' 2>/dev/null | xargs)
SHARD2_COUNT=$(psql "host=localhost port=5551 user=user1 dbname=db1" -t -c 'SELECT COUNT(*) FROM articles;' 2>/dev/null | xargs)

echo "Shard 1: $SHARD1_COUNT articles"
echo "Shard 2: $SHARD2_COUNT articles"
echo ""

if [[ "$SHARD1_COUNT" -gt 0 && "$SHARD2_COUNT" -gt 0 ]]; then
    echo "✅ Data is distributed across both shards!"
else
    echo "⚠️  Data might not be properly distributed"
fi

echo ""
echo "🔍 Sample data from each shard:"
echo "Shard 1 sample:"
psql "host=localhost port=5550 user=user1 dbname=db1" -c "SELECT id, left(title, 50) as title FROM articles LIMIT 2;" 2>/dev/null

echo ""
echo "Shard 2 sample:"
psql "host=localhost port=5551 user=user1 dbname=db1" -c "SELECT id, left(title, 50) as title FROM articles LIMIT 2;" 2>/dev/null

# Unset password
unset PGPASSWORD
