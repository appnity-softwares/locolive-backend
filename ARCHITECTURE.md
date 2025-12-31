# System Architecture & Design

## Core Philosophy
**"Post daily to exist. Stop posting to disappear."**

The backend enforces a strict privacy-first, presence-based social model. Users must be active (posts/pings) to be visible. Inactivity leads to automatic "fading" and eventual invisibility from discovery surfaces.

## 1. Visibility Engine
The visibility system automates privacy based on `last_active_at`.

| State | Condition | Visibility Scope |
|-------|-----------|------------------|
| **Active** | Active < 24h | **Full**: Feed, Map, Crossings, Search |
| **Fading** | Active < 3 days | **Reduced**: Profile only, Conn. Feed |
| **Hidden** | Inactive > 3 days | **Invisible**: Search Only (Ghost) |

- **Feed Logic**: Radius expands (5km -> 10km -> 15km -> 20km) looking for *Active* users.
- **Map Logic**: Clusters strictly filter out *Hidden* users.
- **Crossings**: GPS intersections only recorded/shown if both users were *Active* at time of crossing.

## 2. Privacy & Safety
- **Ghost Mode**: `PUT /location/ghost-mode`. Immediate invisibility.
- **Panic Mode**: `POST /location/panic`. Nuclear option. Requires password. Deletes ALL data (Profile, Stories, Locations, Messages, Connections) and wipes Redis cache.
- **Chat Locking**: DM access strictly requires `status = 'accepted'` connection. No random DMs.
- **Anonymous Stories**: Option to post without identity linkage in Feed.

## 3. Technology Stack
- **Language**: Go (Golang) 1.23+
- **Framework**: Gin (HTTP), Gorilla (WS)
- **Database**: PostgreSQL + PostGIS (Spatial Data)
- **Caching**: Redis (User Profiles, Feeds, Rate Limits)
- **Auth**: JWT (Access/Refresh Tokens)

## 4. Key Workflows
### A. Story Feed Expansion
Algorithm aggressively caches "Nearby Stories" in Redis.
1. Check `stories:feed:{geohash}` cache.
2. If miss, Query PostGIS `ST_DWithin(5km)`.
3. If < 5 stories, Expand to 10km -> 20km.
4. Cache result with 5min TTL.

### B. Connection Handshake
1. **Request**: User A -> User B. (Spam Limit: 20/day).
2. **Pending**: User B sees request. Chat Locked.
3. **Accept**: User B accepts. Chat Unlocked.
4. **Block**: User B blocks. User A cannot retry.

## 5. Background Workers
- **CleanupWorker**: Runs daily. Hard deletes expired locations (>30d) and stories (>24h).
- **CrossingDetector**: Runs every minute. Matches recent location pings to find intersections.
