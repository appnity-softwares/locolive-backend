# Production Deployment Guide

## Pre-Deployment Checklist

### ✅ Code Quality
- [x] All tests passing (6/6)
- [x] No linting errors
- [x] Clean build
- [x] No unused imports
- [x] Code formatted

### ✅ Configuration
- [x] Environment variables documented
- [x] Database migrations ready
- [x] Docker Compose configured
- [x] Connection pooling optimized

### ✅ Security
- [x] JWT secret configured
- [x] Password hashing (bcrypt)
- [x] SQL injection protected
- [x] CORS configurable

### ✅ Performance
- [x] Redis caching enabled
- [x] Database indexes created
- [x] Gzip compression enabled
- [x] Connection pooling configured

### ✅ Monitoring
- [x] Structured logging (zerolog)
- [x] Request logging
- [x] Error tracking
- [x] Graceful shutdown

---

## Deployment Steps

### 1. Environment Setup
```bash
# Copy environment template
cp app.env.example app.env

# Update with production values
nano app.env
```

### 2. Database Setup
```bash
# Start PostgreSQL + Redis
docker-compose up -d

# Run migrations
make migrateup
```

### 3. Build Application
```bash
# Build binary
go build -o bin/server ./cmd/server

# Or use Docker
docker build -t privacy-social-backend .
```

### 4. Start Server
```bash
# Direct
./bin/server

# Or with Docker
docker run -p 8080:8080 privacy-social-backend
```

### 5. Verify Deployment
```bash
# Health check
curl http://localhost:8080/users

# Create test user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"phone":"+1234567890","username":"test","full_name":"Test","password":"test123"}'
```

---

## Production Environment Variables

```env
# Database
DB_DRIVER=postgres
DB_SOURCE=postgresql://user:password@host:5432/dbname?sslmode=require

# Redis
REDIS_ADDRESS=redis:6379

# JWT
TOKEN_SYMMETRIC_KEY=your-32-character-secret-key-here
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h

# Server
SERVER_ADDRESS=0.0.0.0:8080
GIN_MODE=release
```

---

## Monitoring

### Logs
```bash
# View logs
tail -f /var/log/privacy-social-backend.log

# Or with Docker
docker logs -f privacy-social-backend
```

### Metrics
- Response times in logs
- Cache hit rates via X-Cache header
- Database connection pool stats

---

## Backup Strategy

### Database
```bash
# Daily backup
pg_dump privacy_social > backup_$(date +%Y%m%d).sql

# Restore
psql privacy_social < backup_20251215.sql
```

### Redis
```bash
# Save snapshot
docker exec privacy_social_redis redis-cli SAVE

# Copy RDB file
docker cp privacy_social_redis:/data/dump.rdb ./backup/
```

---

## Scaling Checklist

When you reach capacity:

1. **Add Read Replicas** (10K users)
   - 1 primary + 2 replicas
   - Cost: +$100/month

2. **Redis Cluster** (15K users)
   - 3 master + 3 replica
   - Cost: +$50/month

3. **Load Balancer** (20K users)
   - Nginx or cloud LB
   - Multiple app instances
   - Cost: +$20/month

4. **CDN** (50K users)
   - CloudFlare or AWS CloudFront
   - Cost: +$50/month

---

## Troubleshooting

### Server Won't Start
```bash
# Check logs
tail -f server.log

# Verify database connection
psql $DB_SOURCE

# Check Redis
redis-cli PING
```

### Slow Performance
```bash
# Check cache hit rate
curl -I http://localhost:8080/feed?latitude=37.7&longitude=-122.4

# Monitor database
docker exec privacy_social_db psql -U postgres -c "SELECT * FROM pg_stat_activity;"
```

### High Memory Usage
```bash
# Check connection pool
# Reduce MaxOpenConns in main.go

# Check Redis memory
docker exec privacy_social_redis redis-cli INFO memory
```

---

## Support

- **Documentation:** See README.md
- **API Reference:** See API_TESTING.md
- **Performance:** See PERFORMANCE_REPORT.md
- **Scaling:** See SCALABILITY_GUIDE.md

---

**Status:** Ready for Production ✅  
**Last Updated:** December 15, 2025
