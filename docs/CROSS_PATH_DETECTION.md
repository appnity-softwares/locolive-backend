# Cross-Path Detection 🔥

## Feature Overview

**Goal:** Unique differentiation - the "You crossed paths" magic moment.

Detects when two users were at the same location at the same time, creating serendipitous connection opportunities.

---

## How It Works

### 1. Location Tracking
- Users passively update location via `POST /location/ping`
- Location converted to **geohash** (7-char precision ~76m)
- Time rounded to **10-minute buckets**
- Data stored for 24 hours

### 2. Crossing Detection (Background Worker)
- Runs every **5 minutes**
- Looks back 10-15 minutes
- Finds users with:
  - **Same geohash** (within ~76m)
  - **Same time bucket** (within 10 minutes)
- Creates crossing records

### 3. 24-Hour Validity
- Crossings expire after 24 hours
- Only recent crossings shown
- Keeps data fresh and relevant

### 4. Ranking
- **By recency:** Most recent crossings first
- **By time:** "just now", "5 minutes ago", "2 hours ago"

---

## API Endpoint

### Get Cross-Path Suggestions

**Endpoint:** `GET /crossings`

**Request:**
```bash
curl -X GET "http://localhost:8080/crossings?limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "crossings": [
    {
      "id": "uuid",
      "username": "john_doe",
      "full_name": "John Doe",
      "avatar_url": "https://...",
      "location_center": "9q8yy9x",
      "occurred_at": "2025-12-15T04:30:00Z",
      "time_ago": "5 minutes ago"
    },
    {
      "id": "uuid",
      "username": "jane_smith",
      "full_name": "Jane Smith",
      "avatar_url": null,
      "location_center": "9q8yy9x",
      "occurred_at": "2025-12-15T03:45:00Z",
      "time_ago": "1 hour ago"
    }
  ],
  "count": 2,
  "message": "You crossed paths with 2 people!"
}
```

---

## UI Messages

Dynamic messages based on crossing count:

| Count | Message |
|-------|---------|
| 0 | "No crossed paths yet. Keep exploring!" |
| 1 | "You crossed paths with 1 person!" |
| 2-4 | "You crossed paths with X people!" |
| 5+ | "Wow! You crossed paths with X people!" |

---

## Time Formatting

User-friendly relative time:

| Duration | Display |
|----------|---------|
| < 1 min | "just now" |
| 1-59 min | "X minutes ago" |
| 1-23 hours | "X hours ago" |
| 24 hours | "today" |

---

## Technical Implementation

### Database Schema
```sql
CREATE TABLE crossings (
  id UUID PRIMARY KEY,
  user_id_1 UUID REFERENCES users(id),
  user_id_2 UUID REFERENCES users(id),
  location_center TEXT,  -- geohash
  occurred_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Detection Logic
```sql
-- Find potential crossings
SELECT 
  l1.user_id as user1,
  l2.user_id as user2,
  l1.geohash as location,
  l1.time_bucket
FROM locations l1
JOIN locations l2 ON 
  l1.geohash = l2.geohash AND
  l1.time_bucket = l2.time_bucket AND
  l1.user_id < l2.user_id
WHERE l1.created_at BETWEEN $1 AND $2;
```

### Background Worker
```go
// Runs every 5 minutes
ticker := time.NewTicker(5 * time.Minute)

// Detects crossings from last 10-15 minutes
minTime := now.Add(-15 * time.Minute)
maxTime := now
```

---

## Privacy Features

### No Exact Locations
- Only geohash stored (~76m precision)
- No raw GPS coordinates
- Location fuzzing built-in

### Time Bucketing
- 10-minute intervals
- Prevents precise timing tracking
- "Around 3:30 PM" not "3:32:47 PM"

### Ephemeral Data
- Locations expire in 24 hours
- Crossings expire in 24 hours
- Auto-cleanup every 10 minutes

---

## Use Cases

### 1. Serendipitous Connections
"You crossed paths with Sarah at the coffee shop this morning!"

### 2. Event Discovery
Multiple crossings at same location = popular event

### 3. Friend Suggestions
Crossed paths multiple times = potential friend

### 4. Safety Verification
"You were near 5 other users at the concert"

---

## Frontend Integration

### Display Crossings List
```javascript
const response = await fetch('/crossings?limit=20', {
  headers: { 'Authorization': `Bearer ${token}` }
});

const data = await response.json();

// Show magic moment
if (data.count > 0) {
  showNotification(data.message, 'magic');
}

// Render list
data.crossings.forEach(crossing => {
  renderCrossingCard({
    name: crossing.full_name,
    username: crossing.username,
    avatar: crossing.avatar_url,
    timeAgo: crossing.time_ago,
    location: crossing.location_center
  });
});
```

### Empty State
```javascript
if (data.count === 0) {
  showEmptyState("No crossed paths yet. Keep exploring!");
}
```

---

## Performance

### Background Processing
- Async detection (doesn't block API)
- Runs every 5 minutes
- Processes in 2-minute timeout

### Query Optimization
- Indexed on geohash + time_bucket
- Only looks back 15 minutes
- Efficient JOIN on indexed columns

### Scalability
- O(n²) within geohash bucket
- Geohash pre-filtering reduces n dramatically
- Typical: 5-10 users per bucket

---

## Configuration

```go
const (
    crossingDetectionInterval = 5 * time.Minute
    crossingLookbackWindow    = 15 * time.Minute
    crossingValidityPeriod    = 24 * time.Hour
    defaultCrossingLimit      = 20
    maxCrossingLimit          = 50
)
```

---

## Testing

### Simulate Crossing
```bash
# User 1 updates location
curl -X POST http://localhost:8080/location/ping \
  -H "Authorization: Bearer TOKEN1" \
  -d '{"latitude":37.7749,"longitude":-122.4194}'

# User 2 updates same location (within 5 minutes)
curl -X POST http://localhost:8080/location/ping \
  -H "Authorization: Bearer TOKEN2" \
  -d '{"latitude":37.7749,"longitude":-122.4194}'

# Wait 5 minutes for background worker

# Check crossings
curl http://localhost:8080/crossings \
  -H "Authorization: Bearer TOKEN1"
```

---

## Future Enhancements

- [ ] Crossing notifications (push/WebSocket)
- [ ] "Crossed paths X times" counter
- [ ] Location labels ("Coffee Shop", "Park")
- [ ] Mutual interest matching
- [ ] "Near miss" detection (almost crossed)

---

## Benefits

✅ **Unique Feature** - Differentiates from competitors  
✅ **Privacy-First** - No exact locations exposed  
✅ **Async Processing** - Doesn't slow down API  
✅ **Scalable** - Geohash pre-filtering  
✅ **Magic Moments** - Serendipitous connections  

---

**Status:** ✅ Fully Implemented  
**Background Worker:** Running every 5 minutes  
**API Endpoint:** `GET /crossings`  
**Privacy:** Geohash + Time bucketing
