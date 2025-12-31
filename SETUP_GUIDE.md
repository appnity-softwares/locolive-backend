# Backend Setup & Testing Guide

## Prerequisites
- Go 1.23+
- Docker & Docker Compose
- Make

## Quick Start
1. **Start Infrastructure** (Postgres, Redis):
   ```bash
   make network
   make postgres
   make redis
   make createdb
   ```

2. **Run Migrations**:
   ```bash
   make migrateup
   ```

3. **Start Server**:
   ```bash
   make server
   # Runs on localhost:8080
   ```

## Configuration
Set environment variables in `app.env`:
```ini
DB_SOURCE=postgresql://postgres:password@localhost:5432/privacy_social?sslmode=disable
REDIS_ADDRESS=localhost:6379
SERVER_ADDRESS=0.0.0.0:8080
TOKEN_SYMMETRIC_KEY=YOUR_32_BYTE_SECRET_KEY
```

## Testing
Run the comprehensive test suite (Unit + Integration):
```bash
make test
# Expected: PASS (Coverage > 80% on core API)
```
*Note: Unit tests use `gomock` and bypass Rate Limits in Test Mode.*

## Manual Verification (Curl)
**1. Create User:**
```bash
curl -X POST http://localhost:8080/users \
  -d '{"username":"test","password":"password123","phone":"+1234567890","full_name":"Test User"}'
```

**2. Login:**
```bash
curl -X POST http://localhost:8080/users/login \
  -d '{"phone":"+1234567890","password":"password123"}'
```

**3. Check Activity:**
```bash
curl http://localhost:8080/activity/status -H "Authorization: Bearer <TOKEN>"
```
