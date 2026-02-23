# System Design

## Components

- **Next.js webapp** — SSR, runs as a persistent Node.js process on :3000
- **Go API** (Gin) — REST/JSON, single binary on :8080
- **PostgreSQL** — primary datastore, named Docker volume
- **Redis** — session/token storage, feed cache, rate limiting
- **Meilisearch** — book/author full-text search, named Docker volume
- **S3** — cover images, avatars, CSV exports (managed; no self-hosting)
- **nginx** — TLS termination, reverse proxy (runs on host, not in Docker)

## Deployment

AWS infrastructure managed via Terraform (`infra/`). Single EC2 instance (`t3.medium` / `t3.large`).

App services run via `docker-compose.prod.yml`. Images pulled from GHCR. nginx runs on the host for TLS termination — not in Docker. EC2 IAM role grants S3 access, no credentials in containers.

```
EC2 host
├── nginx (host)              TLS, reverse proxy → :3000 / :8080
└── docker-compose.prod.yml
    ├── webapp    :3000
    ├── api       :8080
    ├── postgres  :5432  (named volume)
    ├── redis     :6379
    └── meilisearch :7700 (named volume)
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

## Local Dev

```bash
docker compose up          # starts postgres, redis, meilisearch
cd api && go run .         # Go API on :8080
cd webapp && npm run dev   # Next.js on :3000
```

Copy `.env.example` → `.env` before starting.

## CI/CD

`.github/workflows/deploy.yml` — runs on push to `main`:

1. `test` — Go tests, TS typecheck, Next.js lint (also runs on PRs)
2. `build-push` — builds `api` and `webapp` images, pushes to GHCR tagged with commit SHA and `latest`
3. `deploy` — copies `docker-compose.prod.yml` to EC2, pulls new images, restarts services

Required GitHub Secrets: `EC2_HOST`, `EC2_SSH_KEY`, `GHCR_TOKEN`.

Rollback: redeploy previous image tag by re-running the deploy job with the target SHA.
