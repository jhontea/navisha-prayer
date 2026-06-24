<div align="center">

# Navisha Prayer 🕌

**A modern Islamic prayer times, fasting schedule, and Eid al-Fitr countdown web app.**

🌐 **Live:** [prayer.navisha.cloud](https://prayer.navisha.cloud)

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=nextdotjs&logoColor=white)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

</div>

---

*Bismillahirrahmanirrahim* — In the name of Allah, the Most Gracious, the Most Merciful.

Navisha Prayer gives you accurate prayer times anywhere in the world, a live
fasting countdown, and an Eid al-Fitr counter — wrapped in a clean, fast,
mobile-friendly interface.

## ✨ Features

- 🕌 **Prayer Times** — Fajr, Sunrise, Dhuhr, Asr, Maghrib, and Isha based on your GPS or any city you search
- 📍 **Smart Location** — City search with autocomplete (Open-Meteo geocoding), GPS detection, and IP-based fallback. Change location and prayer times update instantly — no reload needed
- 🌙 **Fasting Schedule** — Live countdown to Suhoor and Iftar, with support for Ramadan, Syawal, and voluntary fasts
- 🎉 **Eid al-Fitr Countdown** — Real-time days-remaining counter to the next Eid
- ⚡ **Live Countdowns** — Auto-updating timers for the next prayer and fasting windows
- 📱 **Responsive** — Looks great on mobile, tablet, and desktop
- 🛡️ **Resilient** — Request timeouts and graceful error handling keep the UI responsive even if the backend is slow or unreachable

## 🧱 Tech Stack

| Layer     | Technology                                  |
|-----------|---------------------------------------------|
| Backend   | Go 1.22 + Gin                               |
| Frontend  | Next.js 16 + React 19 + TypeScript          |
| Styling   | Tailwind CSS v3                             |
| Database  | SQLite (pure-Go driver via GORM)            |
| Data Fetch| SWR (stale-while-revalidate)                |
| External  | Aladhan API (prayer times) · Open-Meteo (geocoding) |
| Deploy    | Docker + Docker Compose + Nginx             |

## 🏗️ Architecture

```
Next.js frontend  ──/api/v1──▶  Go (Gin) backend  ──▶  Aladhan API (prayer times)
   (SWR cache)                    (SQLite cache)    └─▶  Open-Meteo (city search)
```

- The backend caches prayer times in SQLite to minimize external calls.
- City search uses Open-Meteo's free geocoding API (no key, supports autocomplete).
- In production, Nginx serves the UI and proxies `/api/v1` to the backend on the same origin.

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) for the full design,
folder structure, API schemas, and database layout.

## 🚀 Quick Start

### Option 1 — Docker (recommended)

```bash
git clone <repo-url>
cd navisha-prayer
docker compose up --build
```

| Service  | URL                                   |
|----------|---------------------------------------|
| Frontend | http://localhost:3010                 |
| Backend  | http://localhost:8010/api/v1          |
| Health   | http://localhost:8010/api/v1/health   |

### Option 2 — Manual

**Backend**

```bash
cd backend
cp .env.example .env      # adjust if needed
go mod tidy
go run ./cmd/server       # starts on the port set in .env (default :8080)
```

**Frontend**

```bash
cd frontend
npm install
echo "NEXT_PUBLIC_API_URL=http://localhost:8010" > .env.local
npm run dev               # http://localhost:3000
```

> Make sure `NEXT_PUBLIC_API_URL` matches the backend port from your `.env`.

## ⚙️ Configuration

Backend config is set via environment variables (see [`backend/.env.example`](./backend/.env.example)):

| Variable                        | Default                          | Description                       |
|---------------------------------|----------------------------------|-----------------------------------|
| `NAVISHA_SERVER_PORT`           | `:8080`                          | HTTP server port                  |
| `NAVISHA_SERVER_ENV`            | `development`                    | `development` or `production`     |
| `NAVISHA_SERVER_ALLOW_ORIGINS`  | `http://localhost:3000`          | CORS allowed origins              |
| `NAVISHA_ALADHAN_BASE_URL`      | `https://api.aladhan.com/v1`     | Aladhan API base URL              |
| `NAVISHA_GEOCODING_BASE_URL`    | `https://geocoding-api.open-meteo.com/v1` | Geocoding (city search) base URL |
| `NAVISHA_DATABASE_PATH`         | `./data/navisha.db`              | SQLite database file path         |
| `NAVISHA_DEFAULT_LAT`           | `-6.2088` (Jakarta)              | Default latitude fallback         |
| `NAVISHA_DEFAULT_LON`           | `106.8456` (Jakarta)             | Default longitude fallback        |
| `NAVISHA_DEFAULT_METHOD`        | `3` (Muslim World League)        | Calculation method                |
| `NAVISHA_DEFAULT_TIMEZONE`      | `Asia/Jakarta`                   | Default IANA timezone             |

> **Secrets:** `.env` files are gitignored. Never commit real API keys.

## 📡 API Overview

All endpoints are prefixed with `/api/v1`.

| Method | Endpoint                          | Description                  |
|--------|-----------------------------------|------------------------------|
| GET    | `/prayer-times/today`             | Today's prayer times         |
| GET    | `/prayer-times/date/:date`        | Prayer times for a date      |
| GET    | `/prayer-times/monthly`           | Monthly prayer times         |
| GET    | `/prayer-times/next`              | Next upcoming prayer         |
| GET    | `/fasting/status`                 | Current fasting status       |
| POST   | `/fasting/type`                   | Update fasting type          |
| GET    | `/fasting/schedule`               | Today's Suhoor/Iftar times   |
| GET    | `/eid/countdown`                  | Eid al-Fitr countdown        |
| GET    | `/config/location`                | Get current location         |
| POST   | `/config/location/gps`            | Set location via GPS         |
| POST   | `/config/location/city`           | Set location via city name   |
| GET    | `/config/location/search`         | Search cities (autocomplete) |
| GET    | `/health`                         | Health check                 |

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) for full request/response schemas.

## 🌍 Deployment

Production deployment uses Docker behind Nginx with HTTPS, served at
**prayer.navisha.cloud**. Full step-by-step instructions (including TLS,
backups, and updates) are in:

📄 **[docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md)**

## 🔌 External APIs

- [Aladhan Prayer Times API](https://aladhan.com/prayer-times-api) — prayer time calculations (no key required)
- [Open-Meteo Geocoding API](https://open-meteo.com/en/docs/geocoding-api) — city search (free, no key required)

Please respect each provider's rate limits.

## 🤝 Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📜 License

MIT License — see [LICENSE](./LICENSE) for details.

---

<div align="center">Made with ☪ for the Ummah</div>
