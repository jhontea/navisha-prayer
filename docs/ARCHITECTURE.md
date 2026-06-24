# Navisha Prayer — Architecture Document

> **Version:** 1.0.0  
> **Date:** 2026-02-01  
> **Status:** Design

---

## Table of Contents

1. [Overview](#1-overview)
2. [System Architecture](#2-system-architecture)
3. [Project Folder Structure](#3-project-folder-structure)
4. [Backend Architecture (Go + Gin)](#4-backend-architecture-go--gin)
   - 4.1 [Key Interfaces & Types](#41-key-interfaces--types)
   - 4.2 [API Endpoint Specification](#42-api-endpoint-specification)
   - 4.3 [Database Schema (SQLite)](#43-database-schema-sqlite)
   - 4.4 [Configuration Management](#44-configuration-management)
   - 4.5 [External API Integration](#45-external-api-integration)
5. [Frontend Architecture (Next.js + TypeScript + Tailwind)](#5-frontend-architecture-nextjs--typescript--tailwind)
   - 5.1 [Component Tree](#51-component-tree)
   - 5.2 [State Management](#52-state-management)
   - 5.3 [Data Fetching](#53-data-fetching)
6. [Deployment (Docker)](#6-deployment-docker)
7. [README / Setup Instructions](#7-readme--setup-instructions)

---

## 1. Overview

**Navisha Prayer** is a modern Islamic web application that provides:

| Feature | Description |
|---------|-------------|
| **Prayer Times** | 5 daily prayers (Fajr, Dhuhr, Asr, Maghrib, Isha) + Sunrise, calculated from the user's GPS coordinates or a city search |
| **Fasting Schedule** | Ramadan, Syawal (6-day fast), and general Sawm fasting with live countdowns to Iftar and Suhoor |
| **Eid al-Fitr Countdown** | Real-time countdown timer to the next Eid al-Fitr, sourced from the Aladhan Hijri calendar API |

### Technology Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22+ with Gin HTTP framework |
| Frontend | Next.js 16 (App Router), TypeScript, Tailwind CSS v3 |
| Database | SQLite (local, file-based) |
| External API | [Aladhan Prayer Times API](https://aladhan.com/prayer-times-api) |
| Deployment | Docker + Docker Compose (multi-stage build) |

---

## 2. System Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                        Docker Network                            │
│                                                                  │
│  ┌─────────────────────┐         ┌────────────────────────────┐  │
│  │   Frontend (Next.js) │         │     Backend (Go + Gin)     │  │
│  │   Port: 3000         │────────▶│     Port: 8080              │  │
│  │                      │  HTTP   │                            │  │
│  │  - React Components  │         │  - REST API Handlers       │  │
│  │  - Tailwind CSS      │         │  - Business Logic          │  │
│  │  - SWR / Fetch       │         │  - Aladhan API Client      │  │
│  └─────────────────────┘         │  - SQLite via GORM         │  │
│                                   └──────────┬─────────────────┘  │
│                                              │                    │
│                                   ┌──────────▼─────────────────┐  │
│                                   │      SQLite Database        │  │
│                                   │      (Docker Volume)        │  │
│                                   └────────────────────────────┘  │
│                                                                   │
│                                              │                    │
│                                   ┌──────────▼─────────────────┐  │
│                                   │   Aladhan API (External)    │  │
│                                   │   api.aladhan.com           │  │
│                                   └────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

**Data Flow:**

1. User opens the browser → Next.js frontend served from the Next.js server (SSR).
2. Frontend calls `GET /api/v1/config/location` to get the stored location.
3. Frontend calls `GET /api/v1/prayer-times/today` for prayer data.
4. Backend checks SQLite cache first; if stale/miss, fetches from Aladhan API.
5. Backend returns JSON response; frontend renders with live countdown timers.
6. User can update location via city search or GPS → `POST /api/v1/config/location`.

---

## 3. Project Folder Structure

```
navisha-prayer/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                  # Application entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go                # Config loading (env vars, defaults)
│   │   ├── handler/
│   │   │   ├── prayer.go                # Prayer times HTTP handlers
│   │   │   ├── fasting.go               # Fasting schedule handlers
│   │   │   ├── eid.go                   # Eid countdown handlers
│   │   │   ├── location.go              # Location config handlers
│   │   │   └── health.go                # Health check handler
│   │   ├── service/
│   │   │   ├── prayer.go                # Prayer times business logic
│   │   │   ├── fasting.go               # Fasting calculation logic
│   │   │   ├── eid.go                   # Eid calculation logic
│   │   │   └── location.go              # Location resolution logic
│   │   ├── repository/
│   │   │   ├── prayer_cache.go          # Prayer times cache in SQLite
│   │   │   ├── settings.go              # User settings persistence
│   │   │   └── models.go                # GORM models / table schemas
│   │   ├── client/
│   │   │   └── aladhan.go               # Aladhan API HTTP client
│   │   └── middleware/
│   │       ├── cors.go                  # CORS middleware
│   │       ├── logger.go                # Request logging middleware
│   │       └── error.go                 # Error recovery / JSON error responses
│   ├── migrations/
│   │   └── 001_initial_schema.sql       # Initial DB migration
│   ├── data/
│   │   └── navisha.db                   # SQLite database file (runtime)
│   ├── .env.example                     # Example environment file
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx               # Root layout (fonts, metadata)
│   │   │   ├── page.tsx                 # Home page (main dashboard)
│   │   │   ├── globals.css              # Global Tailwind imports
│   │   │   └── settings/
│   │   │       └── page.tsx             # Settings page (location config)
│   │   ├── components/
│   │   │   ├── prayer/
│   │   │   │   ├── PrayerCard.tsx       # Single prayer time display
│   │   │   │   ├── PrayerList.tsx       # List of all 5 prayers + sunrise
│   │   │   │   └── NextPrayer.tsx       # Highlight next upcoming prayer
│   │   │   ├── fasting/
│   │   │   │   ├── FastingCard.tsx      # Suhoor/Iftar countdown card
│   │   │   │   ├── FastingToggle.tsx    # Enable/disable fasting mode
│   │   │   │   └── RamadanBanner.tsx    # Ramadan-specific banner
│   │   │   ├── eid/
│   │   │   │   └── EidCountdown.tsx     # Eid al-Fitr countdown timer
│   │   │   ├── location/
│   │   │   │   ├── LocationSearch.tsx   # City search input
│   │   │   │   ├── LocationDisplay.tsx  # Current location display
│   │   │   │   └── GpsButton.tsx        # "Use my location" button
│   │   │   └── ui/
│   │   │       ├── Header.tsx           # App header / navigation
│   │   │       ├── Footer.tsx           # App footer
│   │   │       ├── LoadingSpinner.tsx   # Loading state
│   │   │       └── ErrorMessage.tsx     # Error display
│   │   ├── hooks/
│   │   │   ├── usePrayerTimes.ts        # SWR hook for prayer data
│   │   │   ├── useFasting.ts            # SWR hook for fasting data
│   │   │   ├── useEidCountdown.ts       # SWR hook for Eid data
│   │   │   ├── useLocation.ts           # SWR hook for location config
│   │   │   └── useCountdown.ts          # Generic countdown timer hook
│   │   ├── lib/
│   │   │   ├── api.ts                   # API client (fetch wrapper)
│   │   │   └── utils.ts                 # Utility functions (time formatting, etc.)
│   │   └── types/
│   │       ├── prayer.ts                # Prayer time type definitions
│   │       ├── fasting.ts               # Fasting type definitions
│   │       ├── location.ts              # Location type definitions
│   │       └── api.ts                   # API response type definitions
│   ├── public/
│   │   └── icons/                       # App icons, favicon
│   ├── .env.local.example               # Example frontend env file
│   ├── next.config.ts                   # Next.js configuration
│   ├── tailwind.config.ts               # Tailwind CSS configuration
│   ├── tsconfig.json
│   ├── package.json
│   └── Dockerfile
│
├── docker-compose.yml                   # Multi-service orchestration
├── .gitignore
├── LICENSE
├── Makefile                             # Common dev commands
└── README.md                            # Setup and usage instructions
```

---

## 4. Backend Architecture (Go + Gin)

### 4.1 Key Interfaces & Types

```go
// ── internal/config/config.go ──────────────────────────────────

package config

// Config holds all application configuration loaded from environment variables.
type Config struct {
    Server   ServerConfig
    Aladhan  AladhanConfig
    Database DatabaseConfig
}

type ServerConfig struct {
    Port         string // default ":8080"
    Environment  string // "development" | "production"
    AllowOrigins string // CORS allowed origins, e.g. "http://localhost:3000"
}

type AladhanConfig struct {
    BaseURL string // "https://api.aladhan.com/v1"
    APIKey  string // optional; Aladhan free tier may not require one
}

type DatabaseConfig struct {
    Path string // default "./data/navisha.db"
}
```

```go
// ── internal/client/aladhan.go ─────────────────────────────────

package client

// AladhanClient defines the interface for the external Aladhan API client.
type AladhanClient interface {
    GetPrayerTimes(ctx context.Context, date string, lat, lon float64, method int) (*AladhanResponse, error)
    GetHijriCalendar(ctx context.Context, lat, lon float64, method int) (*HijriCalendarResponse, error)
    GetCities(ctx context.Context, query string) (*CitySearchResponse, error)
}

// AladhanResponse is the top-level JSON response from the Aladhan API.
type AladhanResponse struct {
    Code   int             `json:"code"`
    Status string          `json:"status"`
    Data   AladhanData     `json:"data"`
}

type AladhanData struct {
    Timings Timings        `json:"timings"`
    Date    AladhanDate    `json:"date"`
}

type Timings struct {
    Fajr    string `json:"Fajr"`
    Sunrise string `json:"Sunrise"`
    Dhuhr   string `json:"Dhuhr"`
    Asr     string `json:"Asr"`
    Maghrib string `json:"Maghrib"`
    Isha    string `json:"Isha"`
}

type AladhanDate struct {
    Hijri HijriDate `json:"hijri"`
    Gregorian GregorianDate `json:"gregorian"`
}

type HijriDate struct {
    Date  string `json:"date"`
    Month HijriMonth `json:"month"`
}

type HijriMonth struct {
    Number int `json:"number"`
    En     string `json:"en"`
}

type GregorianDate struct {
    Date string `json:"date"`
}

// CitySearchResponse is used for city autocomplete.
type CitySearchResponse struct {
    Code    int      `json:"code"`
    Status  string   `json:"status"`
    Data    []CityEntry `json:"data"`
}

type CityEntry struct {
    Name    string `json:"name"`
    Country string `json:"country"`
    Lat     float64 `json:"latitude,string"`
    Lon     float64 `json:"longitude,string"`
}
```

```go
// ── internal/repository/models.go ──────────────────────────────

package repository

import "time"

// PrayerCache stores fetched prayer times to avoid repeated API calls.
type PrayerCache struct {
    ID          uint      `gorm:"primaryKey"`
    Lat         float64   `gorm:"not null"`
    Lon         float64   `gorm:"not null"`
    Date        string    `gorm:"not null;index"` // "DD-MM-YYYY"
    Method      int       `gorm:"not null"`
    Fajr        string    `gorm:"not null"`
    Sunrise     string    `gorm:"not null"`
    Dhuhr       string    `gorm:"not null"`
    Asr         string    `gorm:"not null"`
    Maghrib     string    `gorm:"not null"`
    Isha        string    `gorm:"not null"`
    RawResponse string    `gorm:"type:text"`      // raw JSON for debugging
    CreatedAt   time.Time `gorm:"autoCreatedAt"`
    
    // Composite unique constraint on (lat_hash, lon_hash, date, method)
}

func (PrayerCache) TableName() string { return "prayer_cache" }

// UserSettings persists user preferences.
type UserSettings struct {
    ID              uint      `gorm:"primaryKey"`
    Key             string    `gorm:"uniqueIndex;not null"`
    Value           string    `gorm:"not null"`
    UpdatedAt       time.Time `gorm:"autoUpdatedAt"`
}

func (UserSettings) TableName() string { return "user_settings" }

// FastingSchedule stores fasting preferences.
type FastingSchedule struct {
    ID        uint      `gorm:"primaryKey"`
    Type      string    `gorm:"not null"` // "ramadan" | "syawal" | "sawm" | "none"
    Enabled   bool      `gorm:"default:false"`
    CreatedAt time.Time `gorm:"autoCreatedAt"`
    UpdatedAt time.Time `gorm:"autoUpdatedAt"`
}

func (FastingSchedule) TableName() string { return "fasting_schedule" }
```

```go
// ── internal/service/prayer.go ─────────────────────────────────

package service

// PrayerService defines the prayer time business logic interface.
type PrayerService interface {
    GetTodayPrayerTimes(ctx context.Context) (*PrayerTimesResponse, error)
    GetPrayerTimesByDate(ctx context.Context, date string) (*PrayerTimesResponse, error)
    GetMonthlyPrayerTimes(ctx context.Context, year, month int) ([]PrayerTimesResponse, error)
    RefreshCache(ctx context.Context) error
}

// PrayerTimesResponse is the canonical API response for prayer times.
type PrayerTimesResponse struct {
    Date    DateInfo      `json:"date"`
    Lat     float64       `json:"latitude"`
    Lon     float64       `json:"longitude"`
    Method  int           `json:"method"`
    Prayers []PrayerEntry `json:"prayers"`
}

type DateInfo struct {
    Hijri   string `json:"hijri"`
    HijriMonth int  `json:"hijri_month_number"`
    Gregorian string `json:"gregorian"`
}

type PrayerEntry struct {
    Name    string `json:"name"`
    Time    string `json:"time"`    // "HH:MM" in 24h format, in the local timezone
    IsNext  bool   `json:"is_next"` // true if this is the next upcoming prayer
}
```

```go
// ── internal/service/fasting.go ────────────────────────────────

package service

// FastingService defines fasting-related business logic.
type FastingService interface {
    GetCurrentFasting(ctx context.Context) (*FastingResponse, error)
    UpdateFastingType(ctx context.Context, fastingType string) error
}

// FastingResponse contains countdown info for Suhoor and Iftar.
type FastingResponse struct {
    IsFasting     bool   `json:"is_fasting"`
    FastingType   string `json:"fasting_type"`    // "ramadan" | "syawal" | "sawm" | "none"
    SuhoorTime    string `json:"suhoor_time"`     // "HH:MM"
    IftarTime     string `json:"iftar_time"`      // "HH:MM"
    SuhoorIn      string `json:"suhoor_in"`       // e.g. "2h 15m"
    IftarIn       string `json:"iftar_in"`        // e.g. "5h 30m"
    SuhoorSeconds int    `json:"suhoor_seconds"`  // seconds until suhoor
    IftarSeconds  int    `json:"iftar_seconds"`   // seconds until iftar
}
```

```go
// ── internal/service/eid.go ────────────────────────────────────

package service

// EidService defines Eid al-Fitr countdown business logic.
type EidService interface {
    GetEidAlFitrCountdown(ctx context.Context) (*EidCountdownResponse, error)
}

type EidCountdownResponse struct {
    IsToday       bool   `json:"is_islamic_today"`
    HijriDate     string `json:"hijri_date"`
    EidDate       string `json:"eid_date"`        // ISO 8601
    EidDateHijri  string `json:"eid_date_hijri"`  // e.g. "1-10-1446"
    DaysRemaining int    `json:"days_remaining"`
    Message       string `json:"message"`         // e.g. "Eid Mubarak!" or countdown text
}
```

```go
// ── internal/service/location.go ──────────────────────────────

package service

// LocationService handles location resolution.
type LocationService interface {
    GetLocation(ctx context.Context) (*LocationResponse, error)
    SetLocationGPS(ctx context.Context, lat, lon float64) error
    SetLocationCity(ctx context.Context, cityName, countryCode string) error
    SearchCities(ctx context.Context, query string) ([]CityResult, error)
}

type LocationResponse struct {
    Lat       float64 `json:"latitude"`
    Lon       float64 `json:"longitude"`
    City      string  `json:"city"`
    Country   string  `json:"country"`
    Source    string  `json:"source"` // "gps" | "city_search" | "default"
}

type CityResult struct {
    Name    string  `json:"name"`
    Country string  `json:"country"`
    Lat     float64 `json:"latitude"`
    Lon     float64 `json:"longitude"`
}
```

### 4.2 API Endpoint Specification

All endpoints are prefixed with `/api/v1`.

#### Prayer Times

| Method | Endpoint | Description | Query Params |
|--------|----------|-------------|--------------|
| `GET` | `/api/v1/prayer-times/today` | Today's prayer times | — |
| `GET` | `/api/v1/prayer-times/date/:date` | Prayer times for a date | `date` = `DD-MM-YYYY` |
| `GET` | `/api/v1/prayer-times/monthly` | Monthly prayer times | `month` (1-12), `year` (e.g. 2026) |
| `GET` | `/api/v1/prayer-times/next` | Next upcoming prayer only | — |

**Response** (`GET /api/v1/prayer-times/today`):
```json
{
  "code": 200,
  "status": "ok",
  "data": {
    "date": {
      "hijri": "23-08-1446",
      "hijri_month_number": 8,
      "gregorian": "23-02-2026"
    },
    "latitude": 3.1390,
    "longitude": 101.6869,
    "method": 3,
    "prayers": [
      { "name": "Fajr",    "time": "05:58", "is_next": false },
      { "name": "Sunrise", "time": "07:18", "is_next": false },
      { "name": "Dhuhr",   "time": "13:21", "is_next": true  },
      { "name": "Asr",     "time": "16:42", "is_next": false },
      { "name": "Maghrib", "time": "19:19", "is_next": false },
      { "name": "Isha",    "time": "20:34", "is_next": false }
    ]
  }
}
```

#### Fasting

| Method | Endpoint | Description | Params |
|--------|----------|-------------|--------|
| `GET` | `/api/v1/fasting/status` | Current fasting status & countdown | — |
| `POST` | `/api/v1/fasting/type` | Update fasting type | Body: `{"type": "ramadan"}` |
| `GET` | `/api/v1/fasting/schedule` | Today's Suhoor/Iftar times | — |

**Response** (`GET /api/v1/fasting/status`):
```json
{
  "code": 200,
  "status": "ok",
  "data": {
    "is_fasting": true,
    "fasting_type": "ramadan",
    "suhoor_time": "05:58",
    "iftar_time": "19:19",
    "suhoor_in": "2h 15m",
    "suhoor_seconds": 8100,
    "iftar_in": "5h 30m",
    "iftar_seconds": 19800
  }
}
```

#### Eid al-Fitr

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/eid/countdown` | Eid al-Fitr countdown |

**Response:**
```json
{
  "code": 200,
  "status": "ok",
  "data": {
    "is_islamic_today": false,
    "hijri_date": "23-08-1446",
    "eid_date": "2026-03-20T00:00:00+08:00",
    "eid_date_hijri": "1-10-1446",
    "days_remaining": 25,
    "message": "25 days until Eid al-Fitr"
  }
}
```

#### Location Configuration

| Method | Endpoint | Description | Params |
|--------|----------|-------------|--------|
| `GET` | `/api/v1/config/location` | Get current location | — |
| `POST` | `/api/v1/config/location/gps` | Set location via GPS | Body: `{"latitude": 3.139, "longitude": 101.6869}` |
| `POST` | `/api/v1/config/location/city` | Set location via city name | Body: `{"city": "Kuala Lumpur", "country": "MY"}` |
| `GET` | `/api/v1/config/location/search` | Search cities | `q` = search query |

### 4.3 Database Schema (SQLite)

Schema is managed via GORM auto-migration on startup; raw SQL is kept in `migrations/` for reference.

```sql
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
```

**Entity Relationship Diagram:**

```
┌─────────────────────┐
│    prayer_cache     │
├─────────────────────┤
│ PK  id              │
│     lat             │
│     lon             │
│     date            │     ┌──────────────────────┐
│     method          │     │    user_settings      │
│     fajr            │     ├──────────────────────┤
│     sunrise         │     │ PK  id               │
│     dhuhr           │     │ UQ  key              │
│     asr             │     │     value            │
│     maghrib         │     │     updated_at       │
│     isha            │     └──────────────────────┘
│     raw_response    │
│     created_at      │     ┌──────────────────────┐
└─────────────────────┘     │  fasting_schedule     │
                            ├──────────────────────┤
                            │ PK  id               │
                            │ UQ  type             │
                            │     enabled          │
                            │     created_at       │
                            │     updated_at       │
                            └──────────────────────┘
```

### 4.4 Configuration Management

Configuration is loaded via environment variables with sensible defaults.

```env
# backend/.env.example

# Server
NAVISHA_SERVER_PORT=:8080
NAVISHA_SERVER_ENV=development
NAVISHA_SERVER_ALLOW_ORIGINS=http://localhost:3000

# Aladhan API
NAVISHA_ALADHAN_BASE_URL=https://api.aladhan.com/v1
NAVISHA_ALADHAN_API_KEY=

# Database
NAVISHA_DATABASE_PATH=./data/navisha.db

# Defaults (used if no user location set)
NAVISHA_DEFAULT_LAT=3.1390
NAVISHA_DEFAULT_LON=101.6869
NAVISHA_DEFAULT_METHOD=3
```

**Priority order for configuration resolution:**

1. Environment variables (highest priority)
2. `.env` file (loaded via `godotenv` in development)
3. Hard-coded defaults (lowest priority)

**Calculation Methods Map** (Aladhan method numbers):

| Code | Method |
|------|--------|
| 0 | Jafari (Shia) |
| 1 | University of Islamic Sciences, Karachi |
| 2 | Islamic Society of North America (ISNA) |
| 3 | Muslim World League (MWL) ← **default** |
| 4 | Umm Al-Qura University, Makkah |
| 5 | Egyptian General Authority of Survey |
| 7 | Institute of Geophysics, University of Tehran |
| 8 | Gulf Region |
| 9 | Kuwait |
| 10 | Qatar |
| 11 | Majlis Ugama Islam Singapura (MUIS) |
| 12 | Union Organization Islamic de France |
| 13 | Diyanet İşleri Başkanlığı, Turkey |
| 14 | Spiritual Administration of Muslims of Russia |
| 15 | Moonsighting Committee |

### 4.5 External API Integration

**Aladhan API Client Design:**

```
┌──────────────────┐      HTTP GET       ┌─────────────────────┐
│  Internal Gin     │ ──────────────────▶ │  api.aladhan.com/v1 │
│  Handler          │ ◀────────────────── │                     │
│                   │    JSON response    │  Endpoints used:    │
│                   │                     │  - /timings/:date   │
│                   │                     │  - /hijriCalendar   │
│                   │                     │  - /cityInfo        │
└──────────────────┘                      └─────────────────────┘
```

**Request Flow with Cache:**

```
Handler → Service → Repository (check cache)
                          │
               ┌──────────┴──────────┐
               │                     │
          Cache HIT              Cache MISS
               │                     │
               ▼                     ▼
         Return cached         Client → Aladhan API
         prayer times          Store in cache
                                     │
                                     ▼
                               Return fresh data
```

**Error Handling:**

- Aladhan API timeout: 15-second HTTP client timeout
- On failure, return stale cache with `X-Cache-Stale: true` header
- If no cache exists, return `503 Service Unavailable` with descriptive message

---

## 5. Frontend Architecture (Next.js + TypeScript + Tailwind)

### 5.1 Component Tree

```
app/layout.tsx (Root Layout)
│
├── app/page.tsx (Home Page / Dashboard)
│   ├── Header
│   │   ├── LocationDisplay
│   │   └── SettingsLink
│   │
│   ├── MainDashboard
│   │   ├── NextPrayerSticky
│   │   │   ├── PrayerCard (highlighted)
│   │   │   └── useCountdown (hook)
│   │   │
│   │   ├── PrayerList
│   │   │   ├── PrayerCard × 6 (Fajr, Sunrise, Dhuhr, Asr, Maghrib, Isha)
│   │   │   └── PrayerCard (styled variant for next/upcoming)
│   │   │
│   │   ├── FastingSection
│   │   │   ├── RamadanBanner (conditional, Ramadan only)
│   │   │   ├── FastingCard
│   │   │   │   ├── SuhoorCountdown
│   │   │   │   ├── IftarCountdown   
│   │   │   │   └── ProgressBar (time between suhoor→iftar)
│   │   │   └── FastingToggle
│   │   │       └── ToggleButton (toggle each type)
│   │   │
│   │   └── EidSection
│   │       └── EidCountdown
│   │           ├── DaysCounter (large number)
│   │           └── MessageText
│   │
│   └── Footer
│
├── app/settings/page.tsx (Settings Page)
│   ├── Header (with back button)
│   ├── LocationSettings
│   │   ├── LocationDisplay (read-only current)
│   │   ├── LocationSearch
│   │   │   ├── SearchInput
│   │   │   └── SearchResults (dropdown)
│   │   ├── GpsButton
│   │   └── CalculationMethodSelect
│   └── FastingSettings
│       └── FastingToggle (per type)
│
└── app/globals.css (Tailwind imports + custom theme)
```

**Component Responsibility Table:**

| Component | Purpose | Server/Client |
|-----------|---------|---------------|
| `PrayerCard` | Display a single prayer time with icon and countdown | Client |
| `PrayerList` | Render all 6 prayer cards with next-prayer highlighting | Client |
| `NextPrayerSticky` | Sticky bar showing the very next prayer at a glance | Client |
| `FastingCard` | Suhoor/Iftar times with live countdown, shows progress through fasting day | Client |
| `RamadanBanner` | Special decorative banner during Ramadan month | Server |
| `FastingToggle` | Toggle switches for Ramadan/Syawal/Sawm fasting | Client |
| `EidCountdown` | Large countdown to Eid al-Fitr with day counter | Client |
| `LocationSearch` | City autocomplete search with API-backed results | Client |
| `LocationDisplay` | Shows current location (city, country) | Server |
| `GpsButton` | Browser Geolocation API trigger | Client |
| `CalculationMethodSelect` | Dropdown to select calculation method | Client |

**Visual Layout (Dashboard):**

```
┌──────────────────────────────────────────────────────┐
│  ☰ Navisha Prayer          Kuala Lumpur, MY    ⚙    │  ← Header
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │ 🕌 Next Prayer: Dhuhr at 01:21 PM             │  │  ← NextPrayerSticky
│  │                     in 2h 15m                 │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐             │
│  │ Fajr     │ │ Sunrise  │ │ Dhuhr ◀  │             │  ← PrayerList
│  │ 05:58 AM │ │ 07:18 AM │ │ 01:21 PM │             │     (next prayer
│  │ ✓ past   │ │ ✓ past   │ │ NEXT     │             │      is highlighted)
│  └──────────┘ └──────────┘ └──────────┘             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐             │
│  │ Asr      │ │ Maghrib  │ │ Isha     │             │
│  │ 04:42 PM │ │ 07:19 PM │ │ 08:34 PM │             │
│  │ 5h 30m   │ │ 7h 15m   │ │ 9h 30m   │             │
│  └──────────┘ └──────────┘ └──────────┘             │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │ 🌙 Ramadan Mubarak!                           │  │  ← RamadanBanner
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │ 🕐 Fasting Schedule                           │  │  ← FastingCard
│  │ Suhoor ends: 05:58 AM  │  in 2h 15m          │  │
│  │ Iftar at:    07:19 PM  │  in 11h 36m ████████│  │
│  │ ☑ Ramadan  □ Syawal  □ Sawm (Mon/Thu)        │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │ 🌙 Eid al-Fitr Countdown                     │  │  ← EidCountdown
│  │              25 DAYS                          │  │
│  │         March 20, 2026                        │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  © 2026 Navisha Prayer · Made with ☪                │  ← Footer
└──────────────────────────────────────────────────────┘
```

### 5.2 State Management

Navisha Prayer uses a **lightweight state management approach** — no Redux/Zustand needed.

| Concern | Solution |
|---------|----------|
| Server data (prayer times, Eid) | **SWR** (stale-while-revalidate) — automatic caching, revalidation, polling |
| Location settings | SWR + `mutate()` on update |
| Fasting toggle | Local `useState` + POST to backend + SWR `mutate()` |
| Countdown timers | `useCountdown` hook using `setInterval` + `Date.now()` comparison |

**SWR Configuration:**
```typescript
// src/lib/swr.ts
import useSWR from 'swr';
import { fetcher } from '@/lib/api';

// Revalidate prayer times every 5 minutes
export const prayerSWRConfig = {
  refreshInterval: 5 * 60 * 1000,
  revalidateOnFocus: true,
  dedupingInterval: 60 * 1000, // dedupe requests within 1 minute
};
```

### 5.3 Data Fetching

```
┌────────┐    SWR    ┌────────────────┐    HTTP    ┌──────────┐
│ Client │ ────────  │ Next.js Route   │ ────────▶ │ Go API   │
│ (React)│ ◀──────── │ Handler (opt.)  │ ◀──────── │ (Gin)    │
│        │  JSON revalidate │            │   JSON    │          │
└────────┘           └────────────────┘           └──────────┘
```

- **Primary path:** Frontend fetches directly from `http://localhost:8080/api/v1/*` (CORS-enabled in dev).
- **Production:** Next.js API routes can proxy the backend for a single-origin setup (optional).

**Key Hooks:**

```typescript
// src/hooks/usePrayerTimes.ts
import useSWR from 'swr';
import { getPrayerTimes } from '@/lib/api';
import { PrayerTimesResponse } from '@/types/prayer';

export function usePrayerTimes() {
  return useSWR<PrayerTimesResponse>('/api/v1/prayer-times/today', fetcher, {
    refreshInterval: 5 * 60 * 1000,
  });
}

// src/hooks/useCountdown.ts
export function useCountdown(targetSeconds: number) {
  // Returns { hours, minutes, seconds, isExpired }
  // Updates every second via setInterval
}
```

---

## 6. Deployment (Docker)

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - prayer-data:/app/data
    environment:
      - NAVISHA_SERVER_PORT=:8080
      - NAVISHA_SERVER_ENV=production
      - NAVISHA_SERVER_ALLOW_ORIGINS=http://localhost:3000
      - NAVISHA_ALADHAN_BASE_URL=https://api.aladhan.com/v1
      - NAVISHA_DATABASE_PATH=/app/data/navisha.db
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
    depends_on:
      backend:
        condition: service_healthy
    restart: unless-stopped

volumes:
  prayer-data:
```

### Backend Dockerfile

```dockerfile
# backend/Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /navisha-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl sqlite-libs

WORKDIR /app
COPY --from=builder /navisha-server .
RUN mkdir -p /app/data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["./navisha-server"]
```

### Frontend Dockerfile

```dockerfile
# frontend/Dockerfile
FROM node:20-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM node:20-alpine

WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package*.json ./
RUN npm ci --only=production

EXPOSE 3000

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/ || exit 1

CMD ["npm", "start"]
```

---

## 7. README / Setup Instructions

```markdown
# Navisha Prayer 🕌

A modern Islamic prayer times, fasting schedule, and Eid al-Fitr countdown web application.

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker](https://www.docker.com/) (optional, for containerized deployment)

## Quick Start

### Option 1: Docker (Recommended)

```bash
git clone <repo-url>
cd navisha-prayer
docker-compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080/api/v1
- Health check: http://localhost:8080/api/v1/health

### Option 2: Manual Setup

#### Backend

```bash
cd backend

# Copy and edit environment variables
cp .env.example .env

# Install dependencies and run
go mod tidy
go run ./cmd/server

# Server starts on :8080
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Configure API URL
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local

# Run dev server
npm run dev

# App available at http://localhost:3000
```

## Building for Production

```bash
# Backend binary
cd backend
go build -o navisha-server ./cmd/server

# Frontend static build
cd frontend
npm run build
npm start
```

## Development Commands

```bash
# Run all via Docker Compose
make up         # docker-compose up --build
make down       # docker-compose down

# Backend
make be-test    # go test ./...
make be-lint    # golangci-lint run

# Frontend
make fe-install # npm install
make fe-dev     # npm run dev
make fe-lint    # npm run lint
```

## API Documentation

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) for complete API specification.

## Aladhan API

This application uses the free [Aladhan Prayer Times API](https://aladhan.com/prayer-times-api).
No API key is required for basic usage. Rate limits apply.

## License

MIT License — see [LICENSE](./LICENSE) for details.
```
