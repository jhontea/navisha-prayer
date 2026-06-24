-- migrations/001_initial_schema.sql

-- Prayer times cache table
-- Rows are unique per (lat_bucket, lon_bucket, date, method) to avoid
-- recalculating for the same effective location.
CREATE TABLE IF NOT EXISTS prayer_cache (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    lat             REAL    NOT NULL,
    lon             REAL    NOT NULL,
    date            TEXT    NOT NULL,    -- "DD-MM-YYYY"
    method          INTEGER NOT NULL DEFAULT 3,
    fajr            TEXT    NOT NULL,
    sunrise         TEXT    NOT NULL,
    dhuhr           TEXT    NOT NULL,
    asr             TEXT    NOT NULL,
    maghrib         TEXT    NOT NULL,
    isha            TEXT    NOT NULL,
    raw_response    TEXT,                -- raw JSON from Aladhan
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(lat, lon, date, method)
);

-- Index for fast lookups
CREATE INDEX IF NOT EXISTS idx_prayer_cache_date ON prayer_cache(date);
CREATE INDEX IF NOT EXISTS idx_prayer_cache_location ON prayer_cache(lat, lon);

-- User settings key-value store
CREATE TABLE IF NOT EXISTS user_settings (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    key             TEXT    NOT NULL UNIQUE,
    value           TEXT    NOT NULL,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Fasting schedule preferences
CREATE TABLE IF NOT EXISTS fasting_schedule (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    type            TEXT    NOT NULL CHECK(type IN ('ramadan', 'syawal', 'sawm', 'none')),
    enabled         BOOLEAN DEFAULT 0,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(type)
);

-- Default settings seed data
INSERT OR IGNORE INTO user_settings (key, value) VALUES ('location.lat',       '3.1390');
INSERT OR IGNORE INTO user_settings (key, value) VALUES ('location.lon',     '101.6869');
INSERT OR IGNORE INTO user_settings (key, value) VALUES ('location.city',    'Kuala Lumpur');
INSERT OR IGNORE INTO user_settings (key, value) VALUES ('location.country', 'MY');
INSERT OR IGNORE INTO user_settings (key, value) VALUES ('location.source',  'default');
INSERT OR IGNORE INTO user_settings (key, value) VALUES ('calculation.method', '3');
INSERT OR IGNORE INTO fasting_schedule (type, enabled) VALUES ('ramadan', 0);
INSERT OR IGNORE INTO fasting_schedule (type, enabled) VALUES ('syawal',  0);
INSERT OR IGNORE INTO fasting_schedule (type, enabled) VALUES ('sawm',    0);
