CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  profile JSONB NOT NULL DEFAULT '{}',
  password_hash TEXT NOT NULL,
  version BIGINT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS venues (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  district TEXT NOT NULL,
  address TEXT NOT NULL,
  latitude DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
  longitude DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
  distance_bucket INTEGER NOT NULL CHECK (distance_bucket BETWEEN 0 AND 5),
  approved BOOLEAN NOT NULL DEFAULT false,
  version BIGINT NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  organizer_id TEXT NOT NULL REFERENCES users(id),
  venue_id TEXT NOT NULL REFERENCES venues(id),
  title TEXT NOT NULL,
  sessions JSONB NOT NULL,
  price_minor BIGINT NOT NULL CHECK (price_minor >= 0),
  currency CHAR(3) NOT NULL,
  capacity INTEGER NOT NULL CHECK (capacity > 0),
  status TEXT NOT NULL CHECK (status IN ('draft','review','published','cancelled','completed')),
  registered INTEGER NOT NULL DEFAULT 0 CHECK (registered >= 0),
  waitlisted INTEGER NOT NULL DEFAULT 0 CHECK (waitlisted >= 0),
  version BIGINT NOT NULL DEFAULT 1,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS events_status_updated_idx ON events(status, updated_at DESC);
CREATE TABLE IF NOT EXISTS invitations (
  id TEXT PRIMARY KEY,
  sender_id TEXT NOT NULL REFERENCES users(id),
  recipient_id TEXT NOT NULL REFERENCES users(id),
  event_id TEXT NOT NULL REFERENCES events(id),
  window_start TIMESTAMPTZ NOT NULL,
  window_end TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  idempotency_key TEXT NOT NULL,
  UNIQUE (recipient_id, event_id, idempotency_key)
);
CREATE TABLE IF NOT EXISTS registrations (
  event_id TEXT NOT NULL REFERENCES events(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  status TEXT NOT NULL,
  position INTEGER NOT NULL DEFAULT 0,
  version BIGINT NOT NULL DEFAULT 1,
  PRIMARY KEY(event_id, user_id)
);
CREATE TABLE IF NOT EXISTS audit_entries (
  id TEXT PRIMARY KEY,
  actor_id TEXT NOT NULL,
  source TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  before_state JSONB NOT NULL,
  after_state JSONB NOT NULL,
  reason TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
