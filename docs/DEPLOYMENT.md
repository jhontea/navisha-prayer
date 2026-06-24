# Deployment Guide — Navisha Prayer

This guide deploys Navisha Prayer with Docker on the VPS
(`103.139.193.21`) behind the existing Nginx, served at
**https://prayer.navisha.cloud**.

## Architecture

```
                       ┌─────────────────────────────────────────┐
 Internet  ──HTTPS──▶  │  Nginx (host)  prayer.navisha.cloud      │
                       │   /          → 127.0.0.1:3010 (frontend) │
                       │   /api/...    → 127.0.0.1:8010 (backend)  │
                       └───────────────┬─────────────┬───────────┘
                                       │             │
                       ┌───────────────────┐   ┌─────────────────────┐
                       │ navisha-prayer-    │   │ navisha-prayer-      │
                       │ frontend           │   │ backend              │
                       │ Next.js :3000      │   │ Go API :8010         │
                       └───────────────────┘   └──────────┬──────────┘
                                                          │
                                                  ┌───────▼────────┐
                                                  │ navisha-data    │
                                                  │ (SQLite volume) │
                                                  └────────────────┘
```

Both containers bind to `127.0.0.1` only — they are never exposed directly to
the internet. Nginx on the host is the single public entry point.

## Prerequisites

On the VPS:

- Docker Engine + Docker Compose plugin (`docker compose version`)
- Nginx (already installed)
- A DNS **A record**: `prayer.navisha.cloud → 103.139.193.21`
- Certbot for TLS (`sudo apt install certbot python3-certbot-nginx`)

## 1. Get the code onto the VPS

SSH in as your user (example user: `ahmadhafizh`):

```bash
ssh ahmadhafizh@103.139.193.21
```

**Option A — clone into your home directory (no sudo needed, recommended):**

```bash
git clone <your-repo-url> ~/navisha-prayer
cd ~/navisha-prayer
```

**Option B — clone into `/opt` (system-wide, requires sudo):**

```bash
# /opt is root-owned, so create the dir with sudo and hand it to your user
sudo mkdir -p /opt/navisha-prayer
sudo chown "$USER":"$USER" /opt/navisha-prayer
git clone <your-repo-url> /opt/navisha-prayer
cd /opt/navisha-prayer
```

> Your user must be in the `docker` group to run `docker compose` without sudo:
> ```bash
> sudo usermod -aG docker "$USER"      # then log out and back in
> ```
> The rest of this guide assumes you run commands from the cloned directory.
> If you chose Option B, replace `~/navisha-prayer` with `/opt/navisha-prayer`.


## 2. Review production environment

Production config is set directly in `docker-compose.yml` under the `backend`
service `environment` block. Adjust if needed:

- `NAVISHA_SERVER_ALLOW_ORIGINS` → `https://prayer.navisha.cloud`
- `NAVISHA_DEFAULT_LAT/LON/METHOD/TIMEZONE` → default location fallback
- `NAVISHA_DATABASE_PATH` → `/app/data/navisha.db` (mapped to the `navisha-data` volume)

> **Secrets:** Do not commit real API keys. The local `backend/.env` is
> gitignored and excluded from the image via `.dockerignore`. If you add a
> required secret later, inject it through the compose `environment` block or an
> `env_file` that stays on the server only.

## 3. Build and start the containers

```bash
docker compose build
docker compose up -d
docker compose ps
```

Verify both are healthy:

```bash
# Backend health (through the localhost-bound port)
curl -s http://127.0.0.1:8010/api/v1/health

# Frontend
curl -sI http://127.0.0.1:3010/ | head -n 1
```

## 4. Configure Nginx

Two configs are shipped:

- `deploy/nginx/prayer.navisha.cloud.http.conf` — **HTTP-only bootstrap**, no
  `ssl_*` lines. Use it for the very first certificate issuance.
- `deploy/nginx/prayer.navisha.cloud.conf` — **full TLS config**. It references
  Let's Encrypt cert paths, so it only passes `nginx -t` *after* a cert exists.

> **Why two files?** The full config references
> `/etc/letsencrypt/live/prayer.navisha.cloud/fullchain.pem`. Before that file
> exists, `nginx -t` fails with `cannot load certificate ... No such file`,
> and the certbot nginx plugin refuses to run. The HTTP-only config breaks this
> chicken-and-egg by giving certbot a valid config to work with first.

### Step 1 — Enable the HTTP-only bootstrap config

```bash
sudo cp deploy/nginx/prayer.navisha.cloud.http.conf \
        /etc/nginx/sites-available/prayer.navisha.cloud.conf
sudo ln -sf /etc/nginx/sites-available/prayer.navisha.cloud.conf \
            /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### Step 2 — Issue the TLS certificate

```bash
sudo certbot --nginx -d prayer.navisha.cloud
```

### Step 3 — Swap in the full TLS config

Now that the cert exists, replace the bootstrap config with the full one and
reload:

```bash
sudo cp deploy/nginx/prayer.navisha.cloud.conf \
        /etc/nginx/sites-available/prayer.navisha.cloud.conf
sudo nginx -t && sudo systemctl reload nginx
```

Certbot installs a renewal timer automatically. Test renewal with:

```bash
sudo certbot renew --dry-run
```

## 5. Verify

Open <https://prayer.navisha.cloud>:

- Home page loads prayer times
- Settings → search a city (e.g. "Majalengka") → prayer times update without reload
- `https://prayer.navisha.cloud/api/v1/health` returns `{"status":"ok"}`

## Updating to a new version

```bash
cd ~/navisha-prayer   # or /opt/navisha-prayer if you used Option B
git pull
docker compose build
docker compose up -d
docker image prune -f   # optional: clean old layers
```

The SQLite database persists in the `navisha-data` volume across rebuilds.

## Continuous deployment (GitHub Actions)

A workflow at `.github/workflows/deploy.yml` deploys automatically on every
push to `main` (and can be triggered manually from the Actions tab). It SSHes
into the VPS and runs the same steps as a manual deploy: `git reset --hard
origin/main` → `docker compose build` → `docker compose up -d` → health check.

### One-time setup

1. **Create a dedicated deploy SSH key** (on your machine or the VPS):

   ```bash
   ssh-keygen -t ed25519 -C "github-actions-deploy" -f deploy_key -N ""
   ```

2. **Authorize the public key on the VPS** (as the deploy user):

   ```bash
   cat deploy_key.pub >> ~/.ssh/authorized_keys
   ```

3. **Add repository secrets** in GitHub → Settings → Secrets and variables →
   Actions → *New repository secret*:

   | Secret | Value |
   | --- | --- |
   | `VPS_HOST` | `103.139.193.21` |
   | `VPS_USER` | `ahmadhafizh` |
   | `VPS_SSH_KEY` | contents of the **private** `deploy_key` file |
   | `VPS_PORT` | *(optional)* SSH port, defaults to `22` |
   | `VPS_APP_DIR` | *(optional)* repo path, defaults to `~/navisha-prayer` |

4. **Ensure the deploy user can run Docker without sudo** (`docker` group) and
   that the repo is already cloned at `VPS_APP_DIR` — the workflow updates an
   existing checkout, it does not clone fresh.

> The workflow uses `git reset --hard origin/main`, which discards any local
> changes on the server. Keep server-only files (e.g. a private `env_file`)
> outside the repo or gitignored so they are not wiped.

## Operations

```bash
# Logs
docker compose logs -f backend
docker compose logs -f frontend

# Restart a service
docker compose restart backend

# Stop everything
docker compose down

# Stop AND delete the database volume (destructive!)
docker compose down -v
```

## Backup & restore the database

```bash
# Backup
docker run --rm -v navisha-prayer_navisha-data:/data -v "$PWD":/backup \
  alpine sh -c "cp /data/navisha.db /backup/navisha-backup-$(date +%F).db"

# Restore
docker run --rm -v navisha-prayer_navisha-data:/data -v "$PWD":/backup \
  alpine sh -c "cp /backup/navisha-backup-YYYY-MM-DD.db /data/navisha.db"
docker compose restart backend
```

> The volume name is prefixed with the compose project (directory) name. Run
> `docker volume ls` to confirm the exact name on your server.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| 502 Bad Gateway | `docker compose ps` — are containers up and healthy? |
| API 404 at `/api/...` | Nginx `location /api/` block present and reloaded? |
| Prayer times wrong/Jakarta | Set location in Settings; default fallback is configurable in compose |
| Frontend can't reach API | `NEXT_PUBLIC_API_URL` build arg should be empty (same-origin) |
| `Failed to find Server Action "…"` in frontend logs after a redeploy | Harmless. A browser tab still running the **old** build sent a request to the **new** container; build hashes change on every rebuild. Hard-refresh the browser (Ctrl+Shift+R) — fresh clients are unaffected. |
| Cert errors | `sudo certbot certificates`; re-run `certbot --nginx` |
| `cannot load certificate .../fullchain.pem ... No such file` when running certbot | The full TLS config is enabled before a cert exists, so `nginx -t` fails and the certbot plugin won't run. Enable the **HTTP-only bootstrap** config first (`prayer.navisha.cloud.http.conf`), reload nginx, issue the cert, then swap in the full config — see Step 1–3 above. |
| `Conflict. The container name "/navisha-prayer-backend" is already in use` | A stale container still holds the name. Run `docker compose down` then `docker compose up -d`. If it was created outside this compose project, run `docker rm -f navisha-prayer-backend navisha-prayer-frontend`, or `docker compose up -d --remove-orphans --force-recreate`. Do **not** use `down -v` (it deletes the database). |
| Port `8010`/`3010` already allocated | An old container still holds the host port. Stop it with `docker compose down` (or `docker rm -f <old-container>`) before starting. |
