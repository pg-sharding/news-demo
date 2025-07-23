# SPQR Demo - News Aggregator

A simple demonstration of PostgreSQL sharding using SPQR (Stateless Postgres Query Router).

This app collects news from RSS feeds and automatically distributes articles across PostgreSQL shards.

## 🚀 Quick Start

```bash
make demo
```

That's it! You now have:
- 2 PostgreSQL shards running
- SPQR router distributing queries automatically  
- 30+ news articles loaded and distributed across shards
- API server ready for testing

## 🧪 Test the Sharding

```bash
# See how data is distributed across shards
make test

# Test the API endpoints  
make demo-api
```

## 📊 What You'll See

1. **Automatic Distribution**: Articles are distributed across shards based on their ID hash
2. **Transparent Querying**: The API queries through SPQR, which routes to the correct shard
3. **Live Data**: Real news articles from Hacker News, Habr, The Verge, and Wired
4. **Simple Validation**: Easy commands to verify everything is working

## 🔧 Architecture

```
[RSS Parser] → [SPQR Router] → [Shard 1 & Shard 2]
                     ↑
              [API Server]
```

## 🛠️ Manual Testing

Connect directly to SPQR router:
```bash
PGPASSWORD=12345678 psql "host=localhost user=user1 dbname=db1 port=16432"
```

Or check individual shards:
```bash
# Shard 1
PGPASSWORD=12345678 psql "host=localhost user=user1 dbname=db1 port=5550"

# Shard 2  
PGPASSWORD=12345678 psql "host=localhost user=user1 dbname=db1 port=5551"
```

Try queries like:
```sql
-- See all articles
SELECT COUNT(*) FROM articles;

-- Get a specific article by ID  
SELECT * FROM articles WHERE id = 2083415601;

-- See which articles are on this shard
SELECT id, left(title, 50) as title FROM articles LIMIT 5;
```

## 🧹 Cleanup

```bash
make clean
```

---

## About SPQR

This demo showcases [SPQR (Stateless Postgres Query Router)](https://github.com/pg-sharding/spqr), which provides:

- **Transparent Sharding**: Applications connect as if to a single PostgreSQL instance
- **Automatic Routing**: Queries are routed to the correct shard based on distribution keys
- **Easy Setup**: No application changes required for existing PostgreSQL apps

