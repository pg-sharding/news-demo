#!/bin/bash

echo "Setting up SPQR demo..."

# Set password for PostgreSQL connections
export PGPASSWORD=12345678

# Wait a bit more to ensure SPQR is fully ready
sleep 3

# The table creation and sharding setup should already be done by init.sql
# Let's verify the table exists by connecting through SPQR
echo "Verifying table setup..."
TABLE_EXISTS=$(psql "host=localhost port=16432 user=user1 dbname=db1" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'articles';" 2>/dev/null | xargs)

if [[ "$TABLE_EXISTS" -eq 0 ]]; then
    echo "Creating articles table via SPQR..."
    psql "host=localhost port=16432 user=user1 dbname=db1" -c "CREATE TABLE articles (id SERIAL PRIMARY KEY, url TEXT UNIQUE, title TEXT, description TEXT);" >/dev/null 2>&1
fi

echo "✅ Database setup complete"

# Load sample data using our RSS parser
echo "Loading sample data..."
./rssparser
echo "✅ Sample data loaded and distributed across shards"

# Unset password
unset PGPASSWORD
