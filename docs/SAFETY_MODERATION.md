# Safety, Trust & Moderation 🔥

## Feature Overview

**Goal:** App feels safe and trustworthy, ready for scale.

Complete safety infrastructure with blocking, reporting, shadow banning, and admin moderation.

---

## ✅ Implemented Features

### 1. Block Users
- ✅ Block via connection system
- ✅ Blocked users cannot message
- ✅ Blocked users cannot send requests
- ✅ Instant effect

**API:**
```bash
POST /connections/update
{
  "requester_id": "uuid",
  "status": "blocked"
}
```

---

### 2. Report System
- ✅ Report users
- ✅ Report stories
- ✅ Report reasons (spam/abuse/inappropriate/other)
- ✅ Optional description
- ✅ Timestamp tracking

**API:**
```bash
POST /reports
{
  "target_user_id": "uuid",
  "target_story_id": "uuid",  // optional
  "reason": "spam",
  "description": "Sending unsolicited messages"
}
```

**Report Reasons:**
- `spam` - Unwanted promotional content
- `abuse` - Harassment or bullying
- `inappropriate` - Offensive content
- `other` - Other violations

---

### 3. Shadow Banning
- ✅ Admin can shadow ban users
- ✅ User appears normal to themselves
- ✅ Content hidden from others
- ✅ Reduces spam without confrontation

**API:**
```bash
POST /admin/users/ban
{
  "user_id": "uuid",
  "ban": true
}
```

**Database:**
```sql
UPDATE users 
SET is_shadow_banned = true 
WHERE id = $1;
```

---

### 4. Profile Trust Indicators
- ✅ Trust level (0-100)
- ✅ Account age
- ✅ Verification status
- ✅ Report count tracking

**Trust Factors:**
- Account age (older = more trusted)
- Number of connections
- Report history (fewer = better)
- Content quality
- Engagement patterns

---

### 5. Admin Moderation
- ✅ View all reports
- ✅ Resolve reports
- ✅ Ban/unban users
- ✅ Delete users
- ✅ Delete stories
- ✅ View user statistics

**Admin Endpoints (8):**
```
GET    /admin/users          - List all users
POST   /admin/users/ban      - Ban/unban user
DELETE /admin/users/:id      - Delete user
GET    /admin/stats          - Platform statistics
GET    /admin/reports        - List reports
PUT    /admin/reports/:id/resolve - Resolve report
GET    /admin/stories        - List all stories
DELETE /admin/stories/:id    - Delete story
```

---

## 🔜 Recommended Enhancements

### 1. Rate Limiting (Redis)

**Implementation:**
```go
import "github.com/ulule/limiter/v3"
import "github.com/ulule/limiter/v3/drivers/store/redis"

// Rate limit configuration
rate := limiter.Rate{
    Period: 1 * time.Minute,
    Limit:  100,  // 100 requests per minute
}

// Redis store
store, _ := redis.NewStore(redisClient)
middleware := limiter.NewMiddleware(limiter.New(store, rate))

// Apply to routes
router.Use(middleware.Handler())
```

**Limits by Endpoint:**
| Endpoint | Limit | Window |
|----------|-------|--------|
| `/users/login` | 5 | 15 min |
| `/stories` | 10 | 1 hour |
| `/messages` | 100 | 1 min |
| `/location/ping` | 60 | 1 hour |

---

### 2. Fake GPS Detection

**Detection Methods:**

#### A. Velocity Check
```go
func detectImpossibleSpeed(oldLat, oldLng, newLat, newLng float64, timeDiff time.Duration) bool {
    distance := calculateDistance(oldLat, oldLng, newLat, newLng)
    speed := distance / timeDiff.Hours() // km/h
    
    // Human max speed ~800 km/h (airplane)
    if speed > 1000 {
        return true // Likely fake GPS
    }
    return false
}
```

#### B. Pattern Analysis
```go
// Detect perfect grid movements (GPS spoofing)
func detectGridPattern(locations []Location) bool {
    // Check if coordinates are too "perfect"
    // Real GPS has natural jitter
}
```

#### C. Geofencing
```go
// Detect impossible location jumps
func detectTeleportation(loc1, loc2 Location, timeDiff time.Duration) bool {
    distance := calculateDistance(loc1, loc2)
    if distance > 100 && timeDiff < 1*time.Minute {
        return true // Impossible jump
    }
    return false
}
```

**Action on Detection:**
- Flag user account
- Reduce trust score
- Notify admin
- Temporarily disable location features

---

### 3. Enhanced Trust Scoring

**Trust Score Algorithm:**
```go
func calculateTrustScore(user User) int {
    score := 50 // Base score
    
    // Account age (max +20)
    daysSinceCreation := time.Since(user.CreatedAt).Hours() / 24
    score += min(int(daysSinceCreation), 20)
    
    // Connections (max +15)
    score += min(user.ConnectionCount * 3, 15)
    
    // Reports against user (max -30)
    score -= user.ReportCount * 10
    
    // Verified (max +15)
    if user.IsVerified {
        score += 15
    }
    
    // Clamp to 0-100
    return max(0, min(100, score))
}
```

**Trust Levels:**
| Score | Level | Badge |
|-------|-------|-------|
| 0-20 | New | 🆕 |
| 21-40 | Basic | ⭐ |
| 41-60 | Trusted | ⭐⭐ |
| 61-80 | Verified | ⭐⭐⭐ |
| 81-100 | Elite | 👑 |

---

## Database Schema

### Reports Table
```sql
CREATE TABLE reports (
  id UUID PRIMARY KEY,
  reporter_id UUID REFERENCES users(id),
  target_user_id UUID REFERENCES users(id),
  target_story_id UUID REFERENCES stories(id),
  reason TEXT CHECK (reason IN ('spam', 'abuse', 'inappropriate', 'other')),
  description TEXT,
  is_resolved BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Users Table (Safety Fields)
```sql
ALTER TABLE users ADD COLUMN is_shadow_banned BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN trust_level INTEGER DEFAULT 50;
ALTER TABLE users ADD COLUMN report_count INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN is_verified BOOLEAN DEFAULT false;
```

---

## Safety Workflows

### Report Handling Flow
```
User Reports → Admin Dashboard → Review → Action
                                         ↓
                        Delete Content / Ban User / Dismiss
```

### Shadow Ban Flow
```
Admin Bans User → User sees normal app
                ↓
                Others don't see user's content
                ↓
                User eventually realizes and stops
```

### Block Flow
```
User A Blocks User B → Immediate effect
                     ↓
                     B cannot message A
                     B cannot see A's stories
                     B cannot send connection request
```

---

## Admin Dashboard Features

### 1. Report Queue
```json
{
  "reports": [
    {
      "id": "uuid",
      "reporter": "john_doe",
      "target": "spam_user",
      "reason": "spam",
      "description": "Sending promotional messages",
      "created_at": "2025-12-15T00:00:00Z",
      "is_resolved": false
    }
  ],
  "count": 15,
  "pending": 10
}
```

### 2. User Moderation
```json
{
  "user": {
    "id": "uuid",
    "username": "suspicious_user",
    "trust_level": 25,
    "report_count": 5,
    "is_shadow_banned": false,
    "created_at": "2025-12-01T00:00:00Z"
  },
  "actions": ["ban", "delete", "verify"]
}
```

### 3. Platform Statistics
```json
{
  "users": {
    "total": 10000,
    "new_24h": 150,
    "active_1h": 500,
    "banned": 25
  },
  "reports": {
    "total": 100,
    "pending": 15,
    "resolved": 85
  },
  "trust": {
    "avg_score": 65,
    "verified_users": 500
  }
}
```

---

## Safety Best Practices

### 1. Content Moderation
- ✅ User reports
- ✅ Admin review queue
- ✅ Quick action buttons
- 🔜 AI content filtering
- 🔜 Image moderation

### 2. User Protection
- ✅ Block functionality
- ✅ Report system
- ✅ Connection gating
- ✅ Privacy controls

### 3. Platform Health
- ✅ Shadow banning
- ✅ Trust scoring
- ✅ Admin tools
- 🔜 Rate limiting
- 🔜 Fake GPS detection

---

## Implementation Priority

### Phase 1: Already Done ✅
- [x] Block users
- [x] Report system
- [x] Shadow banning
- [x] Admin moderation
- [x] Trust levels

### Phase 2: Quick Wins (1-2 days)
- [ ] Rate limiting (Redis)
- [ ] Basic velocity checks
- [ ] Enhanced trust scoring

### Phase 3: Advanced (1 week)
- [ ] Fake GPS detection
- [ ] AI content moderation
- [ ] Automated trust scoring
- [ ] Pattern analysis

---

## Testing

### Test Report Flow
```bash
# 1. User reports another user
curl -X POST http://localhost:8080/reports \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "target_user_id":"uuid",
    "reason":"spam",
    "description":"Unwanted messages"
  }'

# 2. Admin views reports
curl http://localhost:8080/admin/reports?resolved=false \
  -H "Authorization: Bearer ADMIN_TOKEN"

# 3. Admin resolves report
curl -X PUT http://localhost:8080/admin/reports/REPORT_ID/resolve \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Test Shadow Ban
```bash
# Admin bans user
curl -X POST http://localhost:8080/admin/users/ban \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"user_id":"uuid","ban":true}'

# Verify user is banned
curl http://localhost:8080/admin/users \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

---

## Metrics to Track

### Safety Metrics
- Reports per day
- Resolution time
- Ban rate
- Trust score distribution
- Fake GPS detection rate

### Platform Health
- Active users
- Blocked users
- Shadow banned users
- Average trust score
- Report-to-user ratio

---

## Benefits

✅ **Safe Platform** - Multiple safety layers  
✅ **Admin Control** - Comprehensive moderation tools  
✅ **User Trust** - Trust scoring system  
✅ **Scalable** - Ready for growth  
✅ **Privacy-First** - User protection built-in  

---

**Status:** ✅ Core Features Implemented  
**Enhancements:** Rate limiting, GPS detection recommended  
**Admin Tools:** 8 endpoints fully functional  
**Trust System:** Basic implementation ready
