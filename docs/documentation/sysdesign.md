# System Design

## Components

- **Next.js webapp** — SSR, runs as a persistent Node.js process on :3000
- **Go API** (Gin) — REST/JSON, single binary on :8080
- **PostgreSQL** — primary datastore, named Docker volume
- **Redis** — session/token storage, feed cache, rate limiting
- **Meilisearch** — book full-text search, named Docker volume. The Go API indexes all books at startup and on upsert. `/books/search` queries Meilisearch (local results) and Open Library (discovery) concurrently.
- **MinIO** — S3-compatible object store for avatars and file uploads; runs in Docker locally. In production, swap for AWS S3 (change `MINIO_ENDPOINT`, `MINIO_PUBLIC_URL`, and credentials — no code changes required).
- **nginx** — TLS termination, reverse proxy (runs on host, not in Docker)

## Deployment

AWS infrastructure managed via Terraform (`infra/`). Single EC2 instance (`t3.medium` / `t3.large`).

App services run via `docker-compose.prod.yml`. Images pulled from GHCR. nginx runs on the host for TLS termination — not in Docker. EC2 IAM role grants S3 access, no credentials in containers.

```
EC2 host
├── nginx (host)              TLS, reverse proxy → :3000 / :8080
└── docker-compose.prod.yml
    ├── webapp      :3000
    ├── api         :8080
    ├── postgres    :5432  (named volume)
    ├── redis       :6379
    ├── meilisearch :7700  (named volume)
    └── minio       :9000  (named volume) — swap for S3 in prod
```

DNS via Route 53. TLS via Let's Encrypt/Certbot.

Scaling path: move Postgres, Redis, and/or Meilisearch to dedicated instances. No architectural changes required.

### First-time EC2 setup

```bash
# On the instance after terraform apply
mkdir ~/rosslib
cp .env.example ~/rosslib/.env   # fill in prod values
# nginx config + certbot TLS — done once manually
```

### Provisioning

```bash
cd infra
cp terraform.tfvars.example terraform.tfvars  # fill in ssh_public_key, domain
terraform init && terraform apply
# outputs: instance_ip, ssh_command, s3_bucket
```

## Book Cover Images

Cover images are sourced from Open Library's CDN (`covers.openlibrary.org`) and the URL is stored once in the `books.cover_url` column at the time the book is first added to the catalog.

**Write path (book add/import):**

1. A user adds a book or an import runs, triggering a lookup against the Open Library API.
2. The Go API extracts the numeric cover ID from the OL search or work response.
3. It constructs a CDN URL (`https://covers.openlibrary.org/b/id/{cover_id}-{size}.jpg`) and upserts it into `books.cover_url`. Medium (`-M`) is used for search results; Large (`-L`) for direct work lookups.
4. Subsequent adds of the same book (same `open_library_id`) hit the `ON CONFLICT ... DO UPDATE` path and may overwrite the stored URL, but no new outbound request is made at read time.

**Read path (shelf/profile page loads):**

1. The API returns `cover_url` as part of the books JSON from collection queries — no outbound call to Open Library is made.
2. The webapp renders the URL directly in a plain `<img src={cover_url}>` tag.
3. The browser fetches the image directly from Open Library's CDN. Caching is entirely governed by OL's CDN cache headers — there is no proxy or server-side image cache on our end.

**Current limitations / future work:**

- No image proxy or optimization layer (no Next.js `<Image>` component, no self-hosted proxy). This means we have no control over cache headers, no resizing, and no AVIF/WebP conversion.
- Cover images are still served from Open Library's CDN and are not stored in MinIO. A future improvement would be to download and re-host cover images in MinIO/S3, eliminating the CDN dependency and enabling cache control.

## Local Dev

```bash
docker compose up          # starts postgres, redis, meilisearch, minio
cd api && go run .         # Go API on :8080
cd webapp && npm run dev   # Next.js on :3000
```

Copy `.env.example` → `.env` before starting.

MinIO runs on `:9000` (S3 API) and `:9001` (web console, login: `minioadmin`/`minioadmin`). The Go API creates the `rosslib` bucket and sets the avatars public-read policy on startup. Avatar URLs are served directly from `http://localhost:9000`.

## Google OAuth Setup

Google sign-in is optional. When `GOOGLE_CLIENT_ID` is not set, the "Continue with Google" button is hidden and the app works with email/password only.

### 1. Create Google Cloud OAuth credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/) (free — no billing required)
2. Create a project (or select an existing one)
3. Navigate to **APIs & Services > OAuth consent screen**
   - Choose "External" user type
   - Fill in app name and your email (required fields only)
4. Navigate to **APIs & Services > Credentials**
5. Click **Create Credentials > OAuth 2.0 Client ID**
   - Application type: **Web application**
   - Authorized JavaScript origins: your app's base URL (e.g. `http://localhost:3000`)
   - Authorized redirect URIs: `<base-url>/api/auth/google/callback` (e.g. `http://localhost:3000/api/auth/google/callback`)
6. Copy the **Client ID** and **Client Secret**

### 2. Configure environment variables

Add to your `.env` file in the repo root:

```
GOOGLE_CLIENT_ID=<your-client-id>.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=<your-client-secret>
NEXT_PUBLIC_URL=http://localhost:3000
```

`docker-compose.yml` maps these into the webapp container as `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `NEXT_PUBLIC_URL`.

`NEXT_PUBLIC_GOOGLE_CLIENT_ID` is baked into the webapp at **build time** (Next.js inlines `NEXT_PUBLIC_*` vars), so you must rebuild after setting it:

```bash
docker compose build webapp && docker compose up -d
```

### 3. OAuth flow

```
Browser → GET /api/auth/google
       → 307 redirect to Google consent screen
       → User approves
       → Google redirects to /api/auth/google/callback?code=...
       → Callback exchanges code for Google tokens (server-side)
       → Callback fetches Google user info (id, email, name)
       → Callback calls POST /auth/google on the Go API
       → API finds or creates user, returns JWT
       → Callback sets `token` and `username` cookies on the redirect response
       → 302 redirect to base URL (user is now logged in)
```

### Notes

- The redirect URI must **exactly** match what's configured in GCP — including protocol, hostname, port, and path. No trailing slash.
- If accessing the app via a non-localhost hostname (e.g. Tailscale), set `NEXT_PUBLIC_URL` to that hostname and add it to the GCP authorized origins and redirect URIs.
- Google-only users have a random password set internally (PocketBase requires one on auth records). They can set a real password later via Settings to enable email+password login.

## CI/CD

`.github/workflows/deploy.yml` — runs on push to `main`:

1. `test` — Go tests, TS typecheck, Next.js lint (also runs on PRs)
2. `build-push` — builds `api` and `webapp` images, pushes to GHCR tagged with commit SHA and `latest`
3. `deploy` — copies `docker-compose.prod.yml` to EC2, pulls new images, restarts services

Required GitHub Secrets: `EC2_HOST`, `EC2_SSH_KEY`, `GHCR_TOKEN`.

Rollback: redeploy previous image tag by re-running the deploy job with the target SHA.
