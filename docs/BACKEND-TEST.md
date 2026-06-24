# Backend Test Report

**Date:** 2026-06-24
**Project:** Navisha Prayer Backend
**Location:** `C:\Users\PC\go\src\project\navisha-prayer\backend`

---

## 1. Code Review Summary

### Files Read (18 Go source files)

| File | Purpose |
|------|---------|
| `cmd/server/main.go` | Entry point: wires config, DB, repos, services, handlers, routes, graceful shutdown |
| `internal/config/config.go` | Env-based config with defaults (port :8080, Aladhan API, SQLite DB, default location KL) |
| `internal/database/database.go` | SQLite init via GORM auto-migrate + seed defaults |
| `internal/repository/models.go` | GORM models: PrayerCache, UserSettings, FastingSchedule |
| `internal/repository/settings.go` | Repositories: PrayerCacheRepo, SettingsRepo, FastingRepo |
| `internal/client/aladhan.go` | Aladhan API client: prayer times, Hijri calendar, city search |
| `internal/service/prayer.go` | Prayer time logic: cache-first, stale fallback, mark next prayer |
| `internal/service/fasting.go` | Fasting status: Suhoor/Iftar countdown from prayer times |
| `internal/service/eid.go` | Eid al-Fitr countdown via Hijri calendar lookup |
| `internal/service/location.go` | Location: get/set GPS/set city/search cities |
| `internal/handler/health.go` | Health check handler |
| `internal/handler/prayer.go` | Prayer times handlers (today, by date, monthly, next) |
| `internal/handler/fasting.go` | Fasting handlers (status, update type, schedule) |
| `internal/handler/eid.go` | Eid countdown handler |
| `internal/handler/location.go` | Location handlers (get, set GPS, set city, search) |
| `internal/middleware/cors.go` | CORS middleware |
| `internal/middleware/error.go` | Error recovery + request timer middleware |
| `internal/middleware/logger.go` | Request logger + duplicate error handler |

### Architecture

- **Framework:** Gin v1.10
- **ORM:** GORM v1.25 with SQLite driver
- **Database:** SQLite (file-based, auto-migrated)
- **External API:** Aladhan prayer times API
- **Pattern:** Clean layered architecture: handler → service → repository → database

---

## 2. Compilation

### Initial Build Issue

The original project used `gorm.io/driver/sqlite` which requires CGO + a C compiler (gcc). The Windows environment has **no C compiler installed** (no GCC, no MinGW, no Zig).

**Fix applied:** Replaced `gorm.io/driver/sqlite` with `github.com/glebarez/sqlite` (pure Go, CGO-free SQLite driver compatible with GORM).

```
go build ./...  →  SUCCESS (after driver swap)
```

### Files Modified for Build Fix

| File | Change |
|------|--------|
| `go.mod` | `gorm.io/driver/sqlite` → `github.com/glebarez/sqlite` |
| `internal/database/database.go` | Import: `gorm.io/driver/sqlite` → `github.com/glebarez/sqlite` |

---

## 3. Test Results

### 3.1 `GET /api/v1/health`

- **Status:** 200 OK
- **Response:**
```json
{
  "code": 200,
  "data": {
    "service": "navisha-prayer-backend",
    "version": "1.0.0"
  },
  "message": "Navisha Prayer API is running",
  "status": "ok"
}
```
- **Verdict:** PASS

### 3.2 `GET /api/v1/prayer-times/today?latitude=-6.2088&longitude=106.8456&method=2`

- **Status:** 200 OK
- **Response:**
```json
{
  "code": 200,
  "data": {
    "date": {
      "hijri": "09-01-1448",
      "hijri_month_number": 1,
      "gregorian": "24-06-2026"
    },
    "latitude": 3.139,
    "longitude": 101.6869,
    "method": 3,
    "prayers": [
      {"name": "Fajr", "time": "05:51", "is_next": false},
      {"name": "Sunrise", "time": "07:07", "is_next": true},
      {"name": "Dhuhr", "time": "13:16", "is_next": false},
      {"name": "Asr", "time": "16:42", "is_next": false},
      {"name": "Maghrib", "time": "19:25", "is_next": false},
      {"name": "Isha", "time": "20:36", "is_next": false}
    ]
  },
  "status": "ok"
}
```
- **ISSUE FOUND:** Query params `latitude`, `longitude`, `method` are **ignored**. The handler doesn't read query parameters — it uses the stored/default location (Kuala Lumpur: 3.139, 101.6869, method 3) regardless of what the user passes. The `GetTodayPrayerTimes` handler in `internal/handler/prayer.go` calls `h.service.GetTodayPrayerTimes(ctx)` without any parameters, and the service's `getLocation()` reads from DB settings or config defaults.
- **Severity:** Medium — the API accepts the params but silently ignores them.

### 3.3 `GET /api/v1/fasting/status`

- **Status:** 200 OK
- **Response:**
```json
{
  "code": 200,
  "data": {
    "is_fasting": false,
    "fasting_type": "none",
    "suhoor_time": "",
    "iftar_time": "",
    "suhoor_in": "",
    "iftar_in": "",
    "suhoor_seconds": 0,
    "iftar_seconds": 0
  },
  "status": "ok"
}
```
- **Verdict:** PASS (correct — no fasting type is enabled by default)

### 3.4 `GET /api/v1/eid/countdown`

- **Status:** 200 OK
- **Response:**
```json
{
  "code": 200,
  "data": {
    "is_islamic_today": false,
    "hijri_date": "01-10-1446",
    "eid_date": "2026-03-20T00:00:00+07:00",
    "eid_date_hijri": "1-10-1446",
    "days_remaining": -96,
    "message": "Eid Mubarak!"
  },
  "status": "ok"
}
```
- **ISSUES FOUND:**
  1. **`days_remaining: -96`** — Negative days means the fallback date (2026-03-20) is in the past. The Hijri calendar API lookup failed for all 6 months checked, so it fell back to a hardcoded date that's already passed.
  2. **`message: "Eid Mubarak!"`** with negative days — The message logic is wrong. When `days_remaining < 0`, it shows "Eid Mubarak!" but it should show the next future Eid date or a proper countdown.
  3. **Hijri date construction bug** (line 62 in `eid.go`): `hijriDate` is built as `fmt.Sprintf("%s-%d-%s", day.Hijri.Date, day.Hijri.Month.Number, day.Hijri.Date)` which produces "01-10-01" (date-month-date) instead of a proper format like "01-10-1446" (date-month-year).
- **Severity:** High — Eid countdown is broken; shows negative days and wrong message.

### 3.5 `GET /api/v1/config/location`

- **Status:** 200 OK
- **Response:**
```json
{
  "code": 200,
  "data": {
    "latitude": 3.139,
    "longitude": 101.6869,
    "city": "Kuala Lumpur",
    "country": "MY",
    "source": "default"
  },
  "status": "ok"
}
```
- **Verdict:** PASS (returns seeded default location)

### 3.6 `GET /api/v1/config/location/search?q=Jakarta`

- **Status:** 500 Internal Server Error
- **Response:**
```json
{
  "code": 500,
  "message": "city search failed: aladhan API returned status 400: {\n    \"code\": 400,\n    \"status\": \"BAD_REQUEST\",\n    \"data\": \"Please specify a city and country.\"\n}",
  "status": "error"
}
```
- **ISSUE FOUND:** The Aladhan `cityInfo` endpoint requires both `city` and `country` parameters. The code only sends `city=Jakarta` without a country, causing the API to return 400.
- **Severity:** Medium — city search is broken; needs a default country or the API endpoint needs to be changed.

---

## 4. Issues Summary

| # | Severity | Location | Issue |
|---|----------|----------|-------|
| 1 | **High** | `internal/service/eid.go` | Eid countdown shows `days_remaining: -96` with message "Eid Mubarak!" — fallback date is in the past; Hijri date format is wrong (date-month-date instead of date-month-year) |
| 2 | **Medium** | `internal/handler/prayer.go` | Query params `latitude`, `longitude`, `method` on `/prayer-times/today` are silently ignored — always uses stored/default location |
| 3 | **Medium** | `internal/service/location.go` | City search sends only `city` param to Aladhan API, but the API requires `city` + `country` — returns 400 |
| 4 | **Low** | `internal/middleware/logger.go` + `internal/middleware/error.go` | Duplicate panic recovery middleware: both `ErrorHandler()` and `ErrorRecovery()` do the same thing. `ErrorRecovery` is used in main.go; `ErrorHandler` is defined but unused. |
| 5 | **Info** | `go.mod` | Original `gorm.io/driver/sqlite` requires CGO + gcc (not available on this Windows machine). Switched to `github.com/glebarez/sqlite` (pure Go). |
| 6 | **Info** | `internal/service/eid.go` | Hardcoded fallback Eid date (`2026-03-20`) will become stale. Should compute dynamically or use a more robust estimation. |

---

## 5. Additional Observations

- **No test files exist** — the project has zero Go test files (`*_test.go`).
- **Graceful shutdown** is properly implemented with SIGINT/SIGTERM handling and DB close.
- **Caching** works well — prayer times are cached for 1 hour with stale fallback.
- **CORS** is configurable via env var.
- **Request logging and timing** middleware are in place.
- **Error recovery** middleware handles panics gracefully.
- The `GetFastingSchedule` handler reuses `GetCurrentFastingStatus` — this means the "schedule" endpoint returns Suhoor/Iftar times even though it's semantically about the schedule config. Minor design concern.

---

## 6. Recommendations

1. **Fix Eid countdown** — Use the Hijri calendar year from the API response instead of hardcoding "1446"; compute the next 1 Shawwal dynamically; handle the case where the API is unreachable.
2. **Honor query params on `/prayer-times/today`** — Parse `latitude`, `longitude`, `method` from query string and pass them to the service.
3. **Fix city search** — Either add a default country parameter (e.g., "ID" for Indonesia) or switch to a different Aladhan endpoint that accepts city-only queries.
4. **Remove duplicate middleware** — Delete `ErrorHandler()` from `logger.go` since `ErrorRecovery()` in `error.go` is already in use.
5. **Add unit tests** — Especially for service layer logic (prayer time calculation, fasting countdown, Eid logic).
6. **Document the CGO dependency** — If keeping `gorm.io/driver/sqlite`, document that a C compiler is required. The `glebarez/sqlite` swap is a better default for Windows.
