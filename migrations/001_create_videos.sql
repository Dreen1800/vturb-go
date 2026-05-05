CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS videos (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title           TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'pending',
  embed_token     TEXT NOT NULL UNIQUE,
  manifest_url    TEXT,
  thumbnail_url   TEXT,
  duration_sec    INTEGER,
  stream_uid      TEXT,
  stream_status   TEXT,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_videos_embed_token ON videos(embed_token);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
