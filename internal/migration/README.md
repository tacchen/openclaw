# Database Migrations

This directory contains SQL migration scripts for the RSS Reader database.

## Migration Files

| File | Description | Date |
|------|-------------|------|
| 001_init_schema.sql | Initial database schema with all tables | 2026-03-23 |

## How to Apply Migrations

### Using psql directly
```bash
# Apply all migrations
psql -U postgres -d rss -f internal/migration/001_init_schema.sql

# Or with connection string from .env
psql $DATABASE_URL -f internal/migration/001_init_schema.sql
```

### Using docker-compose
```bash
# Connect to the database container
docker-compose exec db psql -U postgres -d rss -f /path/to/migration.sql
```

## Schema Overview

```
users
  ├── feeds (1:N)
  │     └── articles (1:N)
  │           └── article_tags (N:M)
  │                 └── tags
  └── tags (1:N)
```

## Tables

- **users**: User accounts with soft delete support
- **feeds**: RSS feed subscriptions per user
- **articles**: Feed articles with AI summaries
- **tags**: User-defined tags for articles
- **article_tags**: Many-to-many junction table

## Notes

- All foreign keys use CASCADE delete
- GORM soft delete is enabled on users table (deleted_at column)
- Composite unique constraints prevent duplicate feeds per user
