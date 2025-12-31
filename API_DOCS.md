# Privacy Social Backend API Documentation

## Auth
- **POST /users**: Create a new user.
  - Body: `{ "username": "...", "password": "...", "full_name": "...", "phone": "..." }`
  - Returns: `201 Created`
- **POST /users/login**: Login user.
  - Body: `{ "phone": "...", "password": "..." }`
  - Returns: `200 OK` with Access/Refresh tokens

## Stories
- **POST /stories**: Create a new story.
  - Headers: `Authorization: Bearer <token>`
  - Body: `{ "media_url": "...", "media_type": "image|video|text", "lat": 12.34, "lng": 56.78, "is_anonymous": bool, "caption": "..." }`
- **GET /feed**: Get stories nearby (Auto-expanding 5km -> 20km).
  - Query: `?lat=...&lng=...`
- **GET /stories/map**: Get stories for map view (Bounding Box).
  - Query: `?north=...&south=...&east=...&west=...`
- **GET /stories/connections**: Get stories from connected users (Global).

## Connections
- **POST /connections/request**: Send connection request.
  - Body: `{ "target_id": "uuid" }`
- **POST /connections/update**: Accept/Block request.
  - Body: `{ "target_id": "uuid", "status": "accepted|blocked" }`

## Chat (Locked)
- **GET /messages**: Get chat history.
  - Query: `?user_id=target_uuid`
  - **Restriction**: Returns `403 Forbidden` if not mutually connected.
- **GET /ws/chat**: WebSocket for real-time chat.

## Privacy & Activity
- **PUT /location/ghost-mode**: Toggle Ghost Mode.
  - Body: `{ "enabled": true|false }`
- **POST /location/panic**: Trigger Panic Mode (Delete all data).
  - Body: `{ "password": "..." }`
- **GET /activity/status**: Get user's activity/visibility status.
