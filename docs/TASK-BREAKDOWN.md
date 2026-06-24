# Navisha Prayer — Full Review & Fix Task Breakdown

> **Date:** 2026-06-24
> **Goal:** Comprehensive re-review of frontend + backend — fix code, optimize code, improve UX, fix time & location handling
> **Workflow:** Plan → Execute each task → Update this doc as each completes
> **Final Status:** ✅ All 15 tasks completed. Backend (`go build`) and frontend (`next build`) both pass cleanly.

---

## 📋 Task Status Legend
- `[ ]` Not started
- `[~]` In progress
- `[x]` Completed

---

## 🔴 CRITICAL Issues

### Task 1: Eid Countdown uses hardcoded location (not user location) ✅
- **File:** `backend/internal/service/eid.go`
- **Problem:** `GetHijriCalendar(ctx, 3.1390, 101.6869, month, year)` — hardcoded Kuala Lumpur coordinates. The Eid countdown did NOT respect the user's configured location.
- **Fix:** Rewrote `eidService` to inject `*repository.SettingsRepository` + `*config.Config`. Added `location()` helper that reads `location.lat`/`location.lon` from the settings DB, falling back to `cfg.Defaults`. Updated `NewEidService` signature and `main.go` wiring.
- **Status:** `[x]`

### Task 2: Timezone mismatch — "next prayer" can be wrong ✅
- **Files:** `backend/internal/service/prayer.go`, `backend/internal/service/fasting.go`, `backend/internal/service/eid.go`
- **Problem:** Prayer/fasting times from Aladhan are in the location's local timezone, but `markNextPrayer` and `timeToSeconds` compared against `time.Now()` (server timezone). If server is UTC but user is in Indonesia (UTC+7), countdowns would be off by hours.
- **Fix:** Added `loc()` helper to all three services (`prayerService`, `fastingService`, `eidService`). All `time.Now()` calls now use `time.Now().In(s.loc())`. Config updated with `Timezone` field (default `Asia/Jakarta`). Also added `NAVISHA_DEFAULT_TIMEZONE` env var.
- **Status:** `[x]`

---

## 🟠 MAJOR Issues

### Task 3: Monthly prayer times — N+1 API calls (performance) ✅
- **File:** `backend/internal/service/prayer.go`, `backend/internal/client/aladhan.go`
- **Problem:** `GetMonthlyPrayerTimes` fetched each day individually in a loop (up to 31 API calls). Very slow and rate-limit unfriendly.
- **Fix:** Added `GetMonthlyPrayerTimes()` method to `AladhanClient` using the `/calendar/{year}/{month}` endpoint. Rewrote `prayerService.GetMonthlyPrayerTimes` to make a single API call, then cache each day. N+1 reduced to 1 call.
- **Status:** `[x]`

### Task 4: `NextPrayerResponse` type mismatch (frontend ↔ backend) ✅
- **Files:** `frontend/src/types/prayer.ts`, `backend/internal/service/prayer.go`, `backend/internal/handler/prayer.go`
- **Problem:** Frontend type expected `data: PrayerEntry & { time_remaining_seconds: number }`, but the backend `GetNextPrayer` handler only returned the `PrayerEntry` (name, time, is_next) — it never returned `time_remaining_seconds`.
- **Fix:** Added `GetNextPrayer()` method to `PrayerService` interface with `NextPrayerEntry` struct (includes `time_remaining_seconds`). Updated handler to delegate to service. Frontend types already aligned.
- **Status:** `[x]`

### Task 5: `getMonthlyPrayerTimes` return type mismatch ✅
- **File:** `frontend/src/types/prayer.ts`, `frontend/src/lib/api.ts`
- **Problem:** Typed as `fetcher<PrayerTimesResponse>` (single object) but backend returns an array `[]PrayerTimesResponse`.
- **Fix:** Created `MonthlyPrayerTimesResponse` type (`data: PrayerTimesData[]`) and updated `api.getMonthlyPrayerTimes` return type.
- **Status:** `[x]`

### Task 6: `UpdateEnabled` lacks transaction (data integrity) ✅
- **File:** `backend/internal/repository/settings.go`
- **Problem:** `UpdateEnabled` disables all then enables the target in two separate queries without a transaction. If the second query fails, all fasting types are left disabled.
- **Fix:** Wrapped both operations in `r.db.Transaction(func(tx *gorm.DB) error { ... })`.
- **Status:** `[x]`

### Task 7: Settings page does not load current fasting state (UX) ✅
- **File:** `frontend/src/app/settings/page.tsx`
- **Problem:** `fastingType` and `fastingEnabled` were initialized to `'none'`/`false` and never synced with the actual backend fasting status. Toggle buttons didn't reflect what's actually active.
- **Fix:** Added `useFasting()` hook call with `useEffect` to sync local state from `fastingData.data.fasting_type` and `fastingData.data.is_fasting` on mount. Also added `mutateFasting()` after toggling.
- **Status:** `[x]`

---

## 🟡 MINOR / Optimization Issues

### Task 8: Duplicated fasting progress calculation ✅
- **Files:** `frontend/src/lib/utils.ts`, `frontend/src/components/fasting/FastingCard.tsx`
- **Problem:** Same math duplicated in both files.
- **Fix:** Removed local `calculateProgress` in `FastingCard.tsx`, now uses `calculateFastingProgress` from utils.
- **Status:** `[x]`

### Task 9: City search has no debounce (UX + API spam) ✅
- **File:** `frontend/src/components/location/LocationSearch.tsx`
- **Problem:** Every keystroke triggered an API call. Spammy and wasteful.
- **Fix:** Added 300ms debounce using `useRef<setTimeout>` pattern. Extracted `runSearch()` for the actual API call. Clean up timer on unmount.
- **Status:** `[x]`

### Task 10: Eid fallback hardcodes Hijri year 1446 (will rot) ✅
- **File:** `backend/internal/service/eid.go`
- **Problem:** `hijriYear := 1446` is a hardcoded magic number that becomes stale.
- **Fix:** Replaced with dynamic `getHijriYear(eidDate)` function that estimates the Hijri year from a Gregorian date using the approximation `(year - 622) * 1.030684`.
- **Status:** `[x]`

### Task 11: `formatTime12h` doesn't guard invalid input ✅
- **File:** `frontend/src/lib/utils.ts`
- **Problem:** If `time24` is empty or malformed, `parseInt` returns `NaN`, producing "NaN:NaN AM".
- **Fix:** Added early-return guards: check for `null`/empty/missing colon, and check for `NaN` after parsing. Returns `--:--` on invalid input.
- **Status:** `[x]`

### Task 12: Default location should be Jakarta, not Kuala Lumpur ✅
- **Files:** `backend/internal/config/config.go`, `backend/internal/database/database.go`
- **Problem:** App is targeted at Indonesian users but default location was Kuala Lumpur, Malaysia.
- **Fix:** Changed default to Jakarta (`lat: -6.2088, lon: 106.8456`). Updated DB seed defaults to `location.city: "Jakarta"`, `location.country: "ID"`.
- **Status:** `[x]`

### Task 13: Eid messages are Indonesian — localization consistency ✅
- **File:** `backend/internal/service/eid.go`
- **Verification:** Eid countdown messages use Indonesian ("hari menuju Idul Fitri", "Eid Mubarak!"). FastingCard frontend labels are in English but this is acceptable for the UI. No issues found.
- **Status:** `[x]`

### Task 14: Backend build verification ✅
- **Action:** Ran `go build ./...` and `go vet ./...` — both pass with exit code 0, no warnings.
- **Status:** `[x]`

### Task 15: Frontend build verification ✅
- **Action:** Ran `npx tsc --noEmit` (clean) and `npx next build` (compiled successfully in 2.7s, all 4 static pages generated).
- **Status:** `[x]`

---

## ✅ Summary Checklist (Quick Reference)

| # | Task | Severity | Status |
|---|------|----------|--------|
| 1 | Eid hardcoded location | CRITICAL | `[x]` |
| 2 | Timezone mismatch (time) | CRITICAL | `[x]` |
| 3 | Monthly N+1 API calls | MAJOR | `[x]` |
| 4 | NextPrayerResponse type mismatch | MAJOR | `[x]` |
| 5 | Monthly type mismatch | MAJOR | `[x]` |
| 6 | UpdateEnabled no transaction | MAJOR | `[x]` |
| 7 | Settings fasting state not loaded | MAJOR | `[x]` |
| 8 | Duplicated progress calc | MINOR | `[x]` |
| 9 | City search no debounce | MINOR | `[x]` |
| 10 | Eid hardcoded Hijri year | MINOR | `[x]` |
| 11 | formatTime12h invalid input | MINOR | `[x]` |
| 12 | Default location Jakarta | MINOR | `[x]` |
| 13 | Localization consistency | MINOR | `[x]` |
| 14 | Backend build verification | VERIFY | `[x]` |
| 15 | Frontend build verification | VERIFY | `[x]` |

---

## 📝 Files Modified

### Backend
- `backend/internal/config/config.go` — Added `Timezone` field, Jakarta defaults
- `backend/internal/database/database.go` — Jakarta seed defaults
- `backend/internal/client/aladhan.go` — Added `GetMonthlyPrayerTimes` + `MonthlyCalendarResponse`
- `backend/internal/service/prayer.go` — Timezone-aware `loc()`, `GetNextPrayer()`, optimized `GetMonthlyPrayerTimes`
- `backend/internal/service/fasting.go` — Injected `*config.Config`, timezone-aware `loc()`, fixed `timeToSeconds`
- `backend/internal/service/eid.go` — Full rewrite: dynamic location, dynamic Hijri year, Aladhan calendar lookup
- `backend/internal/repository/settings.go` — Transaction-wrapped `UpdateEnabled`
- `backend/internal/handler/prayer.go` — Updated `GetNextPrayer` to use service method
- `backend/cmd/server/main.go` — Updated constructor wiring

### Frontend
- `frontend/src/types/prayer.ts` — Added `MonthlyPrayerTimesResponse`
- `frontend/src/lib/api.ts` — Fixed `getMonthlyPrayerTimes` return type
- `frontend/src/lib/utils.ts` — Guarded `formatTime12h`, shared `calculateFastingProgress`
- `frontend/src/components/fasting/FastingCard.tsx` — Deduped progress calc
- `frontend/src/components/location/LocationSearch.tsx` — Added 300ms debounce
- `frontend/src/app/settings/page.tsx` — Sync fasting state from backend

---

*All tasks completed. Both backend and frontend build cleanly.*

---

## 🐞 Post-Review Bugfix: "Failed to fetch" on GPS / location save

- **Symptom:** `setLocationGPS` threw `TypeError: Failed to fetch` from the settings page.
- **Root cause analysis:**
  1. CORS was hardcoded to only allow `http://localhost:3000`. If the Next.js dev server starts on an alternate port (e.g. `3001` when `3000` is busy), the browser blocks the request → "Failed to fetch".
  2. `backend/.env` still pinned the old Kuala Lumpur defaults, overriding the new Jakarta config defaults.
- **Fixes:**
  - `backend/internal/middleware/cors.go` — Switched to `AllowOriginFunc` that accepts the configured origins **plus** any `http://localhost:*` / `http://127.0.0.1:*` origin in dev. Verified preflight returns `204` with correct CORS headers.
  - `backend/.env` — Updated `NAVISHA_DEFAULT_LAT/LON` to Jakarta and added `NAVISHA_DEFAULT_TIMEZONE=Asia/Jakarta`.
  - `frontend/src/lib/api.ts` — `fetcher` now catches network failures and throws a clear message (`Cannot reach the server at <url>. Is the backend running?`).
- **⚠️ Action required:** Restart the backend server so the new CORS rules and `.env` defaults take effect (a stale instance was running on `:8010`).
- **Status:** `[x]`
