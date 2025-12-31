CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE user_role AS ENUM ('user', 'moderator', 'admin');
CREATE TYPE connection_status AS ENUM ('pending', 'accepted', 'blocked');
CREATE TYPE story_availability AS ENUM ('public', 'connections', 'close_friends');

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  phone varchar NOT NULL UNIQUE,
  password_hash varchar NOT NULL,
  username varchar NOT NULL UNIQUE,
  full_name varchar NOT NULL,
  avatar_url varchar,
  bio text,
  role user_role NOT NULL DEFAULT 'user',
  trust_level int NOT NULL DEFAULT 10,
  is_verified boolean NOT NULL DEFAULT false,
  is_shadow_banned boolean NOT NULL DEFAULT false,
  last_active_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT (now())
);

-- Ephemeral locations for discovery
-- Strictly deleted after 24h
CREATE TABLE locations (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  geohash varchar NOT NULL,
  geom geometry(Point, 4326) NOT NULL,
  time_bucket timestamptz NOT NULL, -- e.g. truncated to nearest 10m
  created_at timestamptz NOT NULL DEFAULT (now()),
  expires_at timestamptz NOT NULL
);

CREATE INDEX idx_locations_geom ON locations USING GIST (geom);
CREATE INDEX idx_locations_geohash ON locations (geohash);
CREATE INDEX idx_locations_expires_at ON locations (expires_at);

CREATE TABLE stories (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  media_url varchar NOT NULL,
  media_type varchar NOT NULL, -- image, video
  thumbnail_url varchar,
  caption text,
  geohash varchar NOT NULL,
  geom geometry(Point, 4326) NOT NULL,
  visibility story_availability NOT NULL DEFAULT 'public',
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX idx_stories_geom ON stories USING GIST (geom);
CREATE INDEX idx_stories_expires_at ON stories (expires_at);

-- Crossings between users (detected by worker)
CREATE TABLE crossings (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id_1 uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_id_2 uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  location_center varchar NOT NULL, -- geohash
  occurred_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX idx_crossings_occurred_at ON crossings (occurred_at);
CREATE INDEX idx_crossings_participants ON crossings (user_id_1, user_id_2);

-- Connections between users
CREATE TABLE connections (
  requester_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  target_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status connection_status NOT NULL DEFAULT 'pending',
  created_at timestamptz NOT NULL DEFAULT (now()),
  updated_at timestamptz NOT NULL DEFAULT (now()),
  PRIMARY KEY (requester_id, target_id)
);

CREATE INDEX idx_connections_status ON connections (status);

-- Persistent messages (only allowed if connection accepted)
CREATE TABLE messages (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  sender_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  receiver_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  content text NOT NULL,
  is_read boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX idx_messages_conversation ON messages (sender_id, receiver_id);

-- Sessions for Refresh Token
CREATE TABLE sessions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token varchar NOT NULL,
  user_agent varchar NOT NULL,
  client_ip varchar NOT NULL,
  is_blocked boolean NOT NULL DEFAULT false,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT (now())
);
