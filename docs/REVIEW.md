# Navisha Prayer — Code Review Report

> **Date:** 2026-06-23  
> **Reviewer:** Automated Code Review  
> **Scope:** Backend (Go/Gin), Frontend (Next.js/TypeScript), Architecture Compliance, Security

---

## Executive Summary

The Navisha Prayer project is a well-structured Islamic prayer times application with a Go backend and Next.js frontend. The codebase demonstrates good architectural practices with clear separation of concerns (handlers → services → repositories), proper use of GORM for SQL injection prevention, and a well-organized component tree. However, several issues were identified ranging from critical (silent error swallowing, unused imports causing compile errors) to minor (missing ARIA labels, inconsistent response wrapping).

**Overall Assessment:** The project is production-ready with caveats. The critical and major issues should be addressed before scaling.

---

## 1. Backend Review

### 1.1 Error Handling

| # | Severity | File | Line | Description |
|---|----------|------|------|-------------|
| 1 | :red_circle: Critical | `backend/internal/service/prayer.go` | 142-143 | `GetMonthlyPrayerTimes` silently skips failed days with `continue`. If a network issue occurs, the caller gets a partial response with no indication of failure. This can mislead the frontend into showing incomplete data. |
| 2 | :yellow_circle: Major | `backend/internal/service/prayer.go` | 179-182 | `repoFromSettings` is a stub that always returns the fallback value and ignores the actual settings repository. The comment says "in production we'd use the settings repo directly" — this means user location preferences are never actually read from the database. The `getLocation()` method always returns defaults. |
| 3 | :yellow_circle: Major | `backend/internal/service/eid.go` | 48 | Variable `_ = now.AddDate(0, monthOffset, 0)` is assigned but never used. The loop iterates with `monthOffset` but doesn't actually query different months — it queries the same current month 6 times. This is both a logic bug (inefficient) and the Eid calculation will never look ahead to future months. |
| 4 | :yellow_circle: Major | `backend/internal/service/eid.go` | 60 | `eidHijri` is constructed but never used (`_ = eidHijri`). The variable is dead code. |
| 5 | :green_circle: Minor | `backend/internal/handler/location.go` | 134 | Unused `parseFloat` helper function. This will cause a Go compile error since it's declared but not used (Go is strict about unused functions in packages). Actually — Go only flags unused *imports* and *local variables*, not unused package-level functions. Still, it's dead code that should be removed. |

### 1.2 SQL Injection Prevention

| # | Severity | File | Description |
|---|----------|------|-------------|
| 6 | :green_circle: Info | `backend/internal/repository/*.go` | All database queries use GORM parameterized queries (`Where("lat = ? AND lon = ?", lat, lon)`). No raw SQL string concatenation is used anywhere. **SQL injection risk is negligible.** |
| 7 | :green_circle: Info | `backend/internal/database/database.go` | The `seedDefaults` function uses GORM's `FirstOrCreate` with struct parameters — safe from injection. |

### 1.3 CORS Configuration

| # | Severity | File | Description |
|---|----------|------|-------------|
| 8 | :yellow_circle: Major | `backend/internal/middleware/cors.go` | CORS allows `PUT`, `PATCH`, `DELETE` methods but the API only defines `GET` and `POST` endpoints. Unnecessary methods should be removed to reduce attack surface. |
| 9 | :green_circle: Minor | `backend/internal/middleware/cors.go` | `AllowCredentials: true` with a specific origin list is correct. However, if `AllowOrigins` contains multiple entries and `AllowCredentials` is true, the browser will only reflect the request origin. This is fine but worth noting for production. |
| 10 | :green_circle: Info | `backend/cmd/server/main.go` | CORS is properly configured via `cfg.Server.AllowOrigins` from environment variables. The default `http://localhost:3000` is appropriate for development. |

### 1.4 API Endpoint Correctness vs Architecture Doc

| # | Severity | Description |
|---|----------|-------------|
| 11 | :green_circle: Info | All endpoints match the architecture specification (Section 4.2): `/api/v1/prayer-times/today`, `/date/:date`, `/monthly`, `/next`, `/api/v1/fasting/status`, `/type`, `/schedule`, `/api/v1/eid/countdown`, `/api/v1/config/location` (GET, POST /gps, POST /city, GET /search), `/api/v1/health`. |
| 12 | :yellow_circle: Major | Architecture doc specifies `GET /api/v1/prayer-times/monthly` should accept `month` and `year` query params. The handler validates these correctly, but the service's `GetMonthlyPrayerTimes` calls `GetPrayerTimesByDate` which checks the cache with the full date string. The cache key includes lat/lon/date/method, so monthly requests for different years could collide if the dates are the same (e.g., March 2025 and March 2026 both have "01-03-XXXX"). Actually the date string includes year (`DD-MM-YYYY`), so this is safe. |
| 13 | :yellow_circle: Major | Architecture doc Section 4.5 specifies: "On failure, return stale cache with `X-Cache-Stale: true` header. If no cache exists, return `503 Service Unavailable`." The current implementation returns stale cache as a normal 200 response without the `X-Cache-Stale` header, and returns 500 instead of 503 when no cache exists. |

### 1.5 Graceful Shutdown

| # | Severity | File | Description |
|---|----------|------|-------------|
| 14 | :green_circle: Info | `backend/cmd/server/main.go` | Graceful shutdown is properly implemented: captures SIGINT/SIGTERM, creates a 5-second context timeout, calls `srv.Shutdown(ctx)`. |
| 15 | :green_circle: Minor | `backend/cmd/server/main.go` | The database connection is never closed on shutdown. `db.Close()` should be called in the shutdown path to ensure SQLite writes are flushed. |

### 1.6 Memory Leaks

| # | Severity | File | Description |
|---|----------|------|-------------|
| 16 | :green_circle: Info | `backend/internal/client/aladhan.go` | HTTP response body is properly closed with `defer resp.Body.Close()` in `doRequest`. |
| 17 | :yellow_circle: Major | `backend/cmd/server/main.go` | The GORM `db` object created in `main()` is never closed. On graceful shutdown, `sql.DB.Close()` should be called to release the SQLite file lock. |

### 1.7 Concurrent Safety

| # | Severity | File | Description |
|---|----------|------|-------------|
| 18 | :green_circle: Info | `backend/internal/repository/*.go` | GORM handles connection pooling natively. SQLite in WAL mode (if enabled) supports concurrent reads. The repository structs hold a `*gorm.DB` which is safe for concurrent use. |
| 19 | :yellow_circle: Major | `backend/internal/service/eid.go` | The `GetEidAlFitrCountdown` method makes up to 6 sequential HTTP calls in a loop. Under concurrent user load, this could cause thread/goroutine exhaustion. Each call also uses the same shared `AladhanAPIClient` which has a 15s timeout — worst case 90 seconds for a single request. |

---

## 2. Frontend Review

### 2.1 TypeScript Types

| # | Severity | File | Description |
|---|----------|------|-------------|
| 20 | :green_circle: Info | `frontend/src/types/*.ts` | All API response types are properly defined. No `any` type abuse found in the codebase. |
| 21 | :yellow_circle: Major | `frontend/src/components/location/LocationSearch.tsx` | Line 43: `response as CitySearchResponse` — type assertion without null check. If the API returns an unexpected shape, this will silently produce `undefined` behavior. Should validate the response structure. |
| 22 | :green_circle: Minor | `frontend/src/components/eid/EidCountdown.tsx` | Line 96-99: The local `useCountdownTo` function is defined inside the component body, which means it's recreated on every render. This is a hook (calls `useCountdown`) defined as a regular function, violating React's rules of hooks if it itself uses hooks. Looking closer — it does call `useCountdown(diff)` which is a hook call. This is **a critical React hooks violation** — calling a hook inside a regular function that's redefined each render. |

### 2.2 Component Structure vs Architecture

| # | Severity | Description |
|---|----------|-------------|
| 23 | :green_circle: Info | Component tree matches the architecture doc (Section 5.1). All specified components exist: `PrayerCard`, `PrayerList`, `NextPrayer`, `FastingCard`, `FastingToggle`, `RamadanBanner`, `EidCountdown`, `LocationSearch`, `LocationDisplay`, `GpsButton`, `Header`, `Footer`, `LoadingSpinner`, `ErrorMessage`. |
| 24 | :yellow_circle: Major | Architecture doc specifies `CalculationMethodSelect` component (Section 5.1, line 773) but it's not implemented. The settings page has no way to change the calculation method. |
| 25 | :green_circle: Minor | Architecture doc specifies `LocationDisplay` as a Server Component, but it's used inside the `Header` (a Server Component) without the `'use client'` directive. However, it uses the `useLocation` hook, which requires client-side rendering. This is correctly marked as a client component in the actual code. |

### 2.3 API Integration Correctness

| # | Severity | File | Description |
|---|----------|------|-------------|
| 26 | :yellow_circle: Major | `frontend/src/lib/api.ts` | `updateFastingType` and other POST methods use `.then((res) => res.json())` without checking `res.ok`. If the server returns a non-200 status, the JSON parse will fail or return an error object that's not caught. |
| 27 | :green_circle: Minor | `frontend/src/lib/api.ts` | The `fetcher<T>` function throws a generic `Error` without attaching the HTTP status code or response body. The SWR `error` object won't have useful debugging information. |
| 28 | :green_circle: Minor | `frontend/src/hooks/useLocation.ts` | `isError` is returned but never used by consumers. The `ErrorMessage` component in settings page checks `isError` but the type is just `unknown` from SWR. |

### 2.4 Responsive Design

| # | Severity | File | Description |
|---|----------|------|-------------|
| 29 | :green_circle: Info | `frontend/src/components/prayer/PrayerList.tsx` | Uses responsive grid: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3`. Properly handles mobile, tablet, and desktop layouts. |
| 30 | :green_circle: Info | `frontend/src/components/ui/Header.tsx` | Uses `hidden sm:flex` for `LocationDisplay` — properly hides location on small screens. |
| 31 | :green_circle: Minor | `frontend/src/app/page.tsx` | The main dashboard doesn't have a mobile-specific padding adjustment. On very small screens (< 360px), the `px-4` might be tight but acceptable. |

### 2.5 Accessibility

| # | Severity | File | Description |
|---|----------|------|-------------|
| 32 | :yellow_circle: Major | `frontend/src/components/prayer/PrayerCard.tsx` | No ARIA attributes. The prayer card doesn't have `role="article"` or `aria-label` for screen readers. The "Next" badge has no semantic meaning. |
| 33 | :yellow_circle: Major | `frontend/src/components/fasting/FastingToggle.tsx` | Toggle switches use `<div>` elements instead of proper `<input type="checkbox">` or `<button>` with `aria-pressed`. Screen readers cannot interpret the toggle state. |
| 34 | :green_circle: Minor | `frontend/src/components/location/LocationSearch.tsx` | Search input lacks `role="combobox"`, `aria-expanded`, `aria-controls`, and `aria-activedescendant` for the dropdown. The dropdown items lack `role="option"`. |
| 35 | :green_circle: Minor | `frontend/src/components/ui/Header.tsx` | Settings link has `aria-label="Settings"` — good. But the main nav lacks `<nav>` semantic element. |
| 36 | :green_circle: Minor | `frontend/src/app/layout.tsx` | `<html lang="en">` is set — good. But the Arabic text in EidCountdown (`عيد مبارك`) should have `dir="rtl"` attribute. |

### 2.6 Error Boundary Handling

| # | Severity | File | Description |
|---|----------|------|-------------|
| 37 | :yellow_circle: Major | `frontend/src/app/page.tsx` | No React Error Boundary is defined. If any component throws during rendering, the entire page will crash with a white screen. Next.js App Router requires an `error.tsx` file in the route segment for proper error handling. |
| 38 | :green_circle: Minor | `frontend/src/components/ui/ErrorMessage.tsx` | The `onRetry` prop is optional. When not provided, users have no way to retry the failed operation from the error message. |

---

## 3. Security Review

| # | Severity | Description |
|---|----------|-------------|
| 39 | :green_circle: Info | No hardcoded secrets found. The `APIKey` defaults to empty string and is loaded from environment variables. |
| 40 | :yellow_circle: Major | **No rate limiting.** The backend has no rate limiting middleware. The Aladhan API has rate limits, and without server-side throttling, a malicious client could exhaust the API quota or cause the backend to make excessive external requests. |
| 41 | :yellow_circle: Major | **No input validation on GPS coordinates.** `SetLocationGPS` accepts any float64 values without range validation (latitude must be -90 to 90, longitude -180 to 180). Invalid coordinates would be stored and sent to Aladhan API, potentially causing errors or unexpected behavior. |
| 42 | :yellow_circle: Major | **No input validation on city search query.** `SearchCities` passes the raw query string to the Aladhan API. While the Aladhan API likely handles this, there's no length limit or sanitization on the backend side. A very long query string could cause issues. |
| 43 | :green_circle: Minor | CORS default allows `http://localhost:3000` which is correct for development. In production, this should be overridden via environment variable to the actual domain. |
| 44 | :green_circle: Minor | The `HealthHandler` exposes service name and version (`"service": "navisha-prayer-backend"`, `"version": "1.0.0"`). This is minor information disclosure but common practice. |
| 45 | :green_circle: Info | No authentication mechanism is implemented. This is acceptable for a local/self-hosted prayer times app but would be a concern if deployed publicly with writeable endpoints. |

---

## 4. Architecture Compliance

### 4.1 Backend Clean Architecture

| # | Severity | Description |
|---|----------|-------------|
| 46 | :green_circle: Info | Backend follows clean architecture: **Handlers** (`internal/handler/`) → **Services** (`internal/service/`) → **Repositories** (`internal/repository/`). Dependency injection is used throughout (services receive repository interfaces). |
| 47 | :yellow_circle: Major | The `prayerService` directly constructs `PrayerCache` structs instead of delegating to the repository. The service should not need to know the repository's internal data model. A `cache.Save()` method on the repository would be cleaner. |
| 48 | :green_circle: Minor | The `PrayerService` interface is defined in the service package but the handler depends on the interface (not the concrete type) — good for testability. Same for other services. |

### 4.2 Frontend Component Tree

| # | Severity | Description |
|---|----------|-------------|
| 49 | :green_circle: Info | Frontend component tree closely matches the architecture doc. The `app/layout.tsx` wraps with `Header` and `Footer`. The `app/page.tsx` composes all dashboard components. |
| 50 | :green_circle: Minor | The `useCountdownTo` hook in `EidCountdown.tsx` is defined as a nested function inside the component but calls `useCountdown` (a hook). This works because `useCountdown` is called with the same arguments each render (the hook rules are satisfied since it's always called), but the wrapper function itself is unnecessary indirection. |

### 4.3 API Endpoints Match Spec

| # | Severity | Description |
|---|----------|-------------|
| 51 | :green_circle: Info | All API endpoints from the architecture doc are implemented. The response structures match the spec. |
| 52 | :yellow_circle: Major | The `HealthResponse` type in the frontend (`types/api.ts`) includes `data.status` and `data.version`, but the backend returns `data.service` and `data.version`. The frontend type doesn't include `service` and incorrectly includes `status` in the data object. |

---

## 5. Summary of Issues by Severity

### :red_circle: Critical (1)
| # | Description |
|---|-------------|
| 1 | `GetMonthlyPrayerTimes` silently swallows errors — partial data returned without indication |

### :yellow_circle: Major (14)
| # | Description |
|---|-------------|
| 2 | `repoFromSettings` is a stub — user location preferences never read from DB |
| 3 | Eid service queries same month 6 times instead of iterating months |
| 6 | CORS allows unnecessary HTTP methods (PUT, PATCH, DELETE) |
| 13 | Stale cache returned without `X-Cache-Stale` header; wrong HTTP status on cache miss |
| 17 | Database connection never closed on graceful shutdown |
| 19 | Eid calculation makes up to 6 sequential HTTP calls (90s worst case) |
| 21 | Unsafe type assertion in LocationSearch |
| 22 | `useCountdownTo` in EidCountdown is a hook-like function defined inside component |
| 24 | `CalculationMethodSelect` component missing from implementation |
| 26 | POST API methods don't check `res.ok` before parsing JSON |
| 32-33 | Missing ARIA attributes on PrayerCard and FastingToggle |
| 37 | No `error.tsx` error boundary in route segments |
| 40 | No rate limiting on API endpoints |
| 41 | No GPS coordinate range validation |
| 42 | No input length validation on city search |
| 47 | Service layer directly constructs repository data models |
| 52 | HealthResponse type mismatch between frontend and backend |

### :green_circle: Minor (12)
| # | Description |
|---|-------------|
| 5 | Unused `parseFloat` helper in location handler |
| 8 | `AllowCredentials` with multiple origins note |
| 15 | DB close missing (same as #17) |
| 25 | LocationDisplay architecture note |
| 27 | Generic error in fetcher |
| 28 | Unused `isError` return value |
| 31 | Mobile padding could be improved |
| 34-36 | Accessibility: missing combobox roles, nav element, RTL dir |
| 38 | `onRetry` optional in ErrorMessage |
| 43 | CORS default is dev-only |
| 44 | Version disclosure in health endpoint |
| 50 | Unnecessary hook wrapper in EidCountdown |

### :green_circle: Info (8)
| # | Description |
|---|-------------|
| 6-7 | SQL injection prevention is solid (GORM parameterized) |
| 10 | CORS properly configured via env vars |
| 11 | All endpoints match spec |
| 12 | Monthly cache key safety |
| 14 | Graceful shutdown properly implemented |
| 16 | HTTP response body properly closed |
| 18 | Concurrent safety via GORM connection pool |
| 46 | Clean architecture with dependency injection |
| 49 | Component tree matches architecture |
| 51 | API response structures match spec |

---

## 6. Recommendations

### Immediate (Before Production)
1. **Fix `repoFromSettings`** — Wire the actual `SettingsRepository` into `PrayerService` so user location preferences are respected.
2. **Add `error.tsx`** — Create error boundaries for the home page and settings page routes.
3. **Add GPS validation** — Validate latitude (-90 to 90) and longitude (-180 to 180) in `SetLocationGPS`.
4. **Fix monthly error handling** — Return an error or partial-data indicator when some days fail in `GetMonthlyPrayerTimes`.
5. **Fix Eid month iteration** — Actually query different months in the loop instead of repeating the same query.

### Short-term (Next Sprint)
6. **Add rate limiting** — Implement a simple token-bucket or use `golang.org/x/time/rate` for per-IP limiting.
7. **Close DB on shutdown** — Extract the `*sql.DB` from GORM and call `Close()` in the shutdown handler.
8. **Add stale cache header** — Implement `X-Cache-Stale: true` header when serving stale cache.
9. **Fix HealthResponse type** — Align frontend types with backend response structure.
10. **Add accessibility attributes** — ARIA roles for combobox, toggle switches, and semantic HTML improvements.

### Long-term (Enhancement)
11. **Add CalculationMethodSelect** — Complete the settings page per architecture spec.
12. **Add request timeout middleware** — Protect against slow clients.
13. **Add structured logging** — Replace `log.Printf` with a structured logger (e.g., `slog` or `zerolog`).
14. **Add integration tests** — Test the full request path with a mocked Aladhan API.
15. **Consider adding ETag/If-None-Match** — For more efficient caching beyond the current TTL approach.

---

## 7. Positive Observations

- **Clean separation of concerns**: The handler → service → repository pattern is consistently applied.
- **Proper use of GORM**: All queries use parameterized statements — no SQL injection risk.
- **Good use of interfaces**: Services depend on interfaces, making them testable.
- **SWR for data fetching**: Excellent choice for stale-while-revalidate caching with automatic revalidation.
- **Well-organized types**: All API responses have corresponding TypeScript types.
- **Responsive design**: Tailwind breakpoints are used effectively.
- **Dark mode**: Properly implemented with a cohesive color scheme.
- **Graceful shutdown**: Signal handling and context-based timeout are correct.
- **HTTP client timeout**: 15-second timeout on Aladhan API calls prevents hanging requests.
- **Stale cache fallback**: The service gracefully degrades when the external API is unavailable.

---

*End of Review Report*
