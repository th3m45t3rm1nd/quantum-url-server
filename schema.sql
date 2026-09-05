CREATE TABLE IF NOT EXISTS urls(
  id            BIGSERIAL PRIMARY KEY,
  code          TEXT NOT NULL UNIQUE, 
  original_url  TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS clicks (
    id         BIGSERIAL PRIMARY KEY,
    code       TEXT NOT NULL,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_agent TEXT,
    referrer   TEXT
);

CREATE INDEX IF NOT EXISTS idx_clicks_code ON clicks (code);

