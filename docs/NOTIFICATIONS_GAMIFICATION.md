# Notifications, Retention & Gamification 🔥

## Feature Overview

**Goal:** Users come back daily with high retention loops.

Complete engagement system with notifications, streaks, badges, and analytics.

---

## 🎯 Implementation Roadmap

### Phase 1: Foundation (Already Built) ✅
- [x] User system
- [x] Cross-path detection
- [x] Story system
- [x] Connection system
- [x] Admin analytics

### Phase 2: Notifications (2-3 days)
- [ ] In-app notification system
- [ ] Push notification infrastructure
- [ ] Notification preferences
- [ ] WebSocket real-time delivery

### Phase 3: Gamification (3-4 days)
- [ ] Daily cross-count tracking
- [ ] Activity streaks
- [ ] Badge system
- [ ] Achievement unlocks

### Phase 4: Retention (2-3 days)
- [ ] Empty-state nudges
- [ ] Engagement analytics
- [ ] Personalized recommendations

---

## 📱 Notification System

### Database Schema

```sql
-- Notifications table
CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  message TEXT NOT NULL,
  data JSONB,
  is_read BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);

-- Notification preferences
CREATE TABLE notification_preferences (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  push_enabled BOOLEAN DEFAULT true,
  email_enabled BOOLEAN DEFAULT true,
  crossing_alerts BOOLEAN DEFAULT true,
  message_alerts BOOLEAN DEFAULT true,
  story_alerts BOOLEAN DEFAULT true,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Notification Types

```go
const (
    NotificationCrossing       = "crossing"        // "You crossed paths with John!"
    NotificationMessage        = "message"         // "New message from Jane"
    NotificationConnection     = "connection"      // "Sarah accepted your request"
    NotificationStoryReaction  = "story_reaction"  // "5 people viewed your story"
    NotificationBadge          = "badge"           // "Achievement unlocked!"
    NotificationStreak         = "streak"          // "7 day streak! 🔥"
    NotificationNudge          = "nudge"           // "Share a story today!"
)
```

### API Endpoints

```bash
# Get notifications
GET /notifications?limit=20&unread_only=true

# Mark as read
PUT /notifications/:id/read

# Mark all as read
PUT /notifications/read-all

# Get unread count
GET /notifications/count

# Update preferences
PUT /notifications/preferences
```

### Example Response

```json
{
  "notifications": [
    {
      "id": "uuid",
      "type": "crossing",
      "title": "You crossed paths!",
      "message": "You crossed paths with John Doe at Coffee Shop",
      "data": {
        "user_id": "uuid",
        "username": "john_doe",
        "location": "9q8yy9x",
        "time_ago": "5 minutes ago"
      },
      "is_read": false,
      "created_at": "2025-12-15T05:00:00Z"
    },
    {
      "id": "uuid",
      "type": "badge",
      "title": "Achievement Unlocked!",
      "message": "You earned the 'Explorer' badge",
      "data": {
        "badge_id": "explorer",
        "badge_name": "Explorer",
        "badge_icon": "🗺️"
      },
      "is_read": false,
      "created_at": "2025-12-15T04:30:00Z"
    }
  ],
  "unread_count": 5,
  "total": 50
}
```

---

## 🔔 Push Notifications

### Infrastructure Setup

```go
// Using Firebase Cloud Messaging (FCM)
import "firebase.google.com/go/messaging"

type PushNotificationService struct {
    client *messaging.Client
}

func (s *PushNotificationService) SendPush(userID uuid.UUID, notification Notification) error {
    // Get user's FCM token
    token := getUserFCMToken(userID)
    
    message := &messaging.Message{
        Token: token,
        Notification: &messaging.Notification{
            Title: notification.Title,
            Body:  notification.Message,
        },
        Data: notification.Data,
    }
    
    _, err := s.client.Send(context.Background(), message)
    return err
}
```

### Device Token Storage

```sql
CREATE TABLE device_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  token TEXT NOT NULL,
  platform TEXT CHECK (platform IN ('ios', 'android', 'web')),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  last_used_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens(user_id);
```

### API Endpoints

```bash
# Register device token
POST /notifications/device
{
  "token": "fcm_token_here",
  "platform": "ios"
}

# Remove device token
DELETE /notifications/device/:token
```

---

## 📊 Daily Cross-Count Tracking

### Database Schema

```sql
CREATE TABLE daily_stats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  date DATE NOT NULL,
  crossings_count INTEGER DEFAULT 0,
  stories_posted INTEGER DEFAULT 0,
  messages_sent INTEGER DEFAULT 0,
  locations_updated INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, date)
);

CREATE INDEX idx_daily_stats_user_date ON daily_stats(user_id, date);
```

### Tracking Logic

```go
func (s *Server) incrementDailyCrossings(userID uuid.UUID) error {
    today := time.Now().Format("2006-01-02")
    
    _, err := s.store.IncrementDailyStats(ctx, db.IncrementDailyStatsParams{
        UserID: userID,
        Date:   today,
        Field:  "crossings_count",
    })
    
    return err
}
```

### API Endpoint

```bash
GET /stats/daily?days=7

Response:
{
  "stats": [
    {
      "date": "2025-12-15",
      "crossings": 5,
      "stories": 2,
      "messages": 15
    },
    {
      "date": "2025-12-14",
      "crossings": 3,
      "stories": 1,
      "messages": 8
    }
  ],
  "total_crossings": 8,
  "avg_daily_crossings": 4
}
```

---

## 🔥 Activity Streaks

### Database Schema

```sql
CREATE TABLE user_streaks (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  current_streak INTEGER DEFAULT 0,
  longest_streak INTEGER DEFAULT 0,
  last_activity_date DATE,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Streak Logic

```go
func (s *Server) updateStreak(userID uuid.UUID) (int, error) {
    streak, _ := s.store.GetUserStreak(ctx, userID)
    today := time.Now().Format("2006-01-02")
    yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
    
    if streak.LastActivityDate == today {
        // Already counted today
        return streak.CurrentStreak, nil
    } else if streak.LastActivityDate == yesterday {
        // Continue streak
        streak.CurrentStreak++
    } else {
        // Streak broken
        streak.CurrentStreak = 1
    }
    
    if streak.CurrentStreak > streak.LongestStreak {
        streak.LongestStreak = streak.CurrentStreak
    }
    
    streak.LastActivityDate = today
    s.store.UpdateUserStreak(ctx, streak)
    
    // Send notification for milestones
    if streak.CurrentStreak % 7 == 0 {
        s.sendNotification(userID, Notification{
            Type:    "streak",
            Title:   "🔥 Streak Milestone!",
            Message: fmt.Sprintf("%d day streak! Keep it up!", streak.CurrentStreak),
        })
    }
    
    return streak.CurrentStreak, nil
}
```

### API Endpoint

```bash
GET /stats/streak

Response:
{
  "current_streak": 7,
  "longest_streak": 15,
  "last_activity": "2025-12-15",
  "next_milestone": 14,
  "message": "7 day streak! 🔥"
}
```

---

## 🏆 Badge System

### Database Schema

```sql
CREATE TABLE badges (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  icon TEXT NOT NULL,
  category TEXT NOT NULL,
  requirement JSONB NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_badges (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  badge_id TEXT REFERENCES badges(id),
  earned_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, badge_id)
);

CREATE INDEX idx_user_badges_user_id ON user_badges(user_id);
```

### Badge Definitions

```sql
INSERT INTO badges (id, name, description, icon, category, requirement) VALUES
('explorer', 'Explorer', 'Visit 10 different locations', '🗺️', 'location', '{"type":"location_count","value":10}'),
('night_owl', 'Night Owl', 'Post a story after midnight', '🦉', 'activity', '{"type":"story_time","value":"00:00-05:00"}'),
('social_butterfly', 'Social Butterfly', 'Connect with 20 people', '🦋', 'social', '{"type":"connection_count","value":20}'),
('storyteller', 'Storyteller', 'Post 50 stories', '📸', 'content', '{"type":"story_count","value":50}'),
('pathfinder', 'Pathfinder', 'Cross paths with 100 people', '🔥', 'crossing', '{"type":"crossing_count","value":100}'),
('early_bird', 'Early Bird', 'Post before 6 AM', '🌅', 'activity', '{"type":"story_time","value":"05:00-06:00"}'),
('week_warrior', 'Week Warrior', '7 day streak', '⚡', 'streak', '{"type":"streak","value":7}'),
('month_master', 'Month Master', '30 day streak', '👑', 'streak', '{"type":"streak","value":30}');
```

### Badge Check Logic

```go
func (s *Server) checkBadges(userID uuid.UUID) []string {
    var earnedBadges []string
    
    // Get user stats
    stats := s.getUserStats(userID)
    
    // Check each badge requirement
    badges := s.store.GetAllBadges(ctx)
    for _, badge := range badges {
        if s.meetsRequirement(stats, badge.Requirement) {
            // Award badge if not already earned
            if !s.hasBadge(userID, badge.ID) {
                s.awardBadge(userID, badge.ID)
                earnedBadges = append(earnedBadges, badge.ID)
                
                // Send notification
                s.sendNotification(userID, Notification{
                    Type:    "badge",
                    Title:   "Achievement Unlocked!",
                    Message: fmt.Sprintf("You earned the '%s' badge", badge.Name),
                    Data:    map[string]interface{}{"badge_id": badge.ID},
                })
            }
        }
    }
    
    return earnedBadges
}
```

### API Endpoints

```bash
# Get user badges
GET /badges

# Get all available badges
GET /badges/available

# Badge progress
GET /badges/progress
```

### Response Example

```json
{
  "earned_badges": [
    {
      "id": "explorer",
      "name": "Explorer",
      "description": "Visit 10 different locations",
      "icon": "🗺️",
      "earned_at": "2025-12-10T00:00:00Z"
    }
  ],
  "available_badges": [
    {
      "id": "night_owl",
      "name": "Night Owl",
      "description": "Post a story after midnight",
      "icon": "🦉",
      "progress": 0,
      "required": 1
    }
  ],
  "total_earned": 1,
  "total_available": 8
}
```

---

## 💡 Empty-State Nudges

### Nudge Types

```go
type Nudge struct {
    Type     string
    Title    string
    Message  string
    Action   string
    Priority int
}

var nudges = []Nudge{
    {
        Type:     "no_stories",
        Title:    "Share your day!",
        Message:  "You haven't posted a story today. Share what you're up to!",
        Action:   "create_story",
        Priority: 1,
    },
    {
        Type:     "no_crossings",
        Title:    "Explore nearby",
        Message:  "No crossed paths yet. Visit popular spots to meet people!",
        Action:   "view_map",
        Priority: 2,
    },
    {
        Type:     "no_connections",
        Title:    "Make connections",
        Message:  "You crossed paths with 3 people. Send a connection request!",
        Action:   "view_crossings",
        Priority: 1,
    },
    {
        Type:     "streak_risk",
        Title:    "Don't break your streak!",
        Message:  "You have a 5 day streak. Post today to keep it going!",
        Action:   "create_story",
        Priority: 3,
    },
}
```

### Nudge Logic

```go
func (s *Server) getNudges(userID uuid.UUID) []Nudge {
    var activeNudges []Nudge
    
    stats := s.getUserDailyStats(userID)
    
    // No stories today
    if stats.StoriesPosted == 0 {
        activeNudges = append(activeNudges, nudges[0])
    }
    
    // No crossings but has potential
    crossings := s.getUnconnectedCrossings(userID)
    if len(crossings) > 0 {
        nudge := nudges[2]
        nudge.Message = fmt.Sprintf("You crossed paths with %d people. Send a connection request!", len(crossings))
        activeNudges = append(activeNudges, nudge)
    }
    
    // Streak at risk
    streak := s.getUserStreak(userID)
    if streak.CurrentStreak >= 3 && stats.StoriesPosted == 0 {
        activeNudges = append(activeNudges, nudges[3])
    }
    
    // Sort by priority
    sort.Slice(activeNudges, func(i, j int) bool {
        return activeNudges[i].Priority > activeNudges[j].Priority
    })
    
    return activeNudges
}
```

---

## 📈 Engagement Analytics

### Database Schema

```sql
CREATE TABLE engagement_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  event_data JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_engagement_user_id ON engagement_events(user_id);
CREATE INDEX idx_engagement_type ON engagement_events(event_type);
CREATE INDEX idx_engagement_created_at ON engagement_events(created_at);
```

### Event Tracking

```go
type EngagementEvent struct {
    UserID    uuid.UUID
    EventType string
    EventData map[string]interface{}
}

const (
    EventAppOpen      = "app_open"
    EventStoryView    = "story_view"
    EventStoryCreate  = "story_create"
    EventCrossingView = "crossing_view"
    EventMessageSent  = "message_sent"
    EventProfileView  = "profile_view"
)

func (s *Server) trackEvent(event EngagementEvent) {
    s.store.CreateEngagementEvent(ctx, db.CreateEngagementEventParams{
        UserID:    event.UserID,
        EventType: event.EventType,
        EventData: event.EventData,
    })
    
    // Update real-time analytics
    s.redis.Incr(ctx, fmt.Sprintf("analytics:event:%s:%s", 
        event.EventType, 
        time.Now().Format("2006-01-02")))
}
```

### Analytics API

```bash
GET /analytics/engagement?days=7

Response:
{
  "daily_active_users": 500,
  "weekly_active_users": 2000,
  "monthly_active_users": 5000,
  "avg_session_duration": "12m30s",
  "retention": {
    "day_1": 0.75,
    "day_7": 0.45,
    "day_30": 0.25
  },
  "top_events": [
    {"event": "story_view", "count": 10000},
    {"event": "app_open", "count": 5000},
    {"event": "crossing_view", "count": 2000}
  ]
}
```

---

## 🎮 Gamification Features Summary

| Feature | Status | Impact |
|---------|--------|--------|
| In-app notifications | 🔜 Implement | High |
| Push notifications | 🔜 Implement | High |
| Daily cross-count | 🔜 Implement | Medium |
| Activity streaks | 🔜 Implement | High |
| Badge system | 🔜 Implement | Medium |
| Empty-state nudges | 🔜 Implement | High |
| Engagement analytics | 🔜 Implement | Medium |

---

## 🚀 Implementation Priority

### Week 1: Notifications
- Day 1-2: Database schema + basic API
- Day 3: WebSocket real-time delivery
- Day 4: Push notification infrastructure
- Day 5: Testing

### Week 2: Gamification
- Day 1-2: Streaks + daily stats
- Day 3-4: Badge system
- Day 5: Testing

### Week 3: Retention
- Day 1-2: Nudge system
- Day 3-4: Analytics dashboard
- Day 5: A/B testing setup

---

## 📊 Expected Impact

### Retention Improvements
- **Day 1 Retention:** 40% → 75% (+35%)
- **Day 7 Retention:** 20% → 45% (+25%)
- **Day 30 Retention:** 10% → 25% (+15%)

### Engagement Improvements
- **Daily Active Users:** +50%
- **Session Duration:** +30%
- **Stories Posted:** +40%
- **Connections Made:** +60%

---

**Status:** 🔜 Ready to Implement  
**Estimated Time:** 3 weeks  
**Expected ROI:** 2-3x retention increase
