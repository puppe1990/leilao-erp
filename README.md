# leilao-erp

Full-stack Go app built with [Cais](https://github.com/puppe1990/cais): Inertia.js, Svelte, Tailwind, and SQLite.

## Stack

- Go 1.26 (net/http stdlib) + gonertia
- Svelte 5 + @inertiajs/svelte (Vite → web/static/build/)
- Tailwind CSS 3.x
- SQLite (modernc.org/sqlite, no CGO)

## Quick start

```bash
cais install  # npm install + go mod tidy
cais dev        # http://localhost:8080
cais test       # full test suite
cais build      # bin/server
```

## Cais CLI

This app was scaffolded with the Cais CLI. Useful commands:

```bash
cais install               # npm install + go mod tidy
cais css                   # build Tailwind
cais dev                   # hot reload + tailwind watch
cais server                # go run ./cmd/server
cais console               # interactive Go REPL + SQL
cais g handler <name>      # handler + test + page template
cais g resource <name>     # model + migration + admin CRUD
cais g page <name>         # page template only
cais g migration <name>    # SQL migration file
cais test                  # go test ./...
cais doctor                # verify setup
```

## CI and pre-commit

GitHub Actions runs Go tests, `golangci-lint`, Prettier, and `npm test` on every push/PR to `main`.

```bash
make pre-commit-install   # once: installs git hooks
make ci                   # test + lint + format-check locally
```

Pre-commit hooks run: trailing whitespace, Prettier, `go fmt`, `go test`, `golangci-lint`, and `npm test`.

## Structure

```
pkg/cais/          → framework (via dependency)
internal/app/      → bootstrap and routes
internal/handlers/ → HTTP handlers
internal/store/    → SQLite + migrations
web/templates/app.html → Inertia root shell
web/src/pages/       → Svelte pages
web/static/          → CSS + Vite build + PWA
cmd/server/        → entry point
```

## Environment variables

| Variable  | Default         | Description      |
| --------- | --------------- | ---------------- |
| PORT      | :8080           | Server port      |
| DB_PATH   | ./data/app.db   | SQLite file path |
| ENV       | development     | Environment      |

Health check: GET /health → {"status":"ok"}

## Auth / admin bootstrap

Public signup is disabled. There is no `/signup` route.

**Development only:** when `ENV=development`, the app seeds a single admin if missing:

- Email: `admin@leilao.local`
- Password: `change-me-now`

**Production:** no auto-seed. Create the first admin via `cais console` (e.g. `store.CreateUser` with a bcrypt hash), or temporarily boot once with `ENV=development` so the seed runs, then switch back to production and change the password. A dedicated seed/console command may be added later.

## Testing on phone (LAN)

1. Run `cais dev` and note the **LAN** URL printed at boot (e.g. `http://192.168.1.10:8080`).
2. Open that URL in mobile Safari/Chrome on the same Wi‑Fi.
3. After template or SSE changes, run `cais pwa --bump` and reinstall the PWA (or clear site data) so the service worker cache refreshes.
4. Run `cais doctor --mobile` to catch flash markup, font CSP, and SW cache issues.

## Deploy (Lightsail)

Target: Ubuntu Lightsail nano (or similar) without Docker — binary + SQLite + systemd + Caddy/nginx HTTPS.

### 1. Build on your machine

```bash
export PATH="$HOME/go/bin:$PATH"
npm run build          # or: cais css && vite build
cais build --os linux --arch amd64 -o bin/server-linux
```

Static assets are embedded via `web/embed.go` for templates; production still needs `web/static` on disk (or set `STATIC_DIR`). Package:

```bash
tar czf release.tar.gz bin/server-linux web/static
scp -i ~/.ssh/your-key.pem release.tar.gz ubuntu@SEU_IP:/tmp/
```

### 2. Layout on the server

```text
/opt/leilao-erp/
  current/
    bin/server          # renamed from server-linux
    web/static/         # CSS, JS, PWA
/var/lib/leilao-erp/
  app.db                # SQLite (persist outside deploys)
/etc/leilao-erp/env     # production env (chmod 600)
```

Create a dedicated user and dirs:

```bash
sudo useradd --system --home /opt/leilao-erp --shell /usr/sbin/nologin leilao
sudo mkdir -p /opt/leilao-erp/current /var/lib/leilao-erp /etc/leilao-erp
sudo chown -R leilao:leilao /opt/leilao-erp /var/lib/leilao-erp
```

Unpack release into `/opt/leilao-erp/current`, rename binary to `bin/server`, and ensure `leilao` can read it.

### 3. Environment

Copy [`.env.production.example`](.env.production.example) to `/etc/leilao-erp/env`:

| Variable         | Example                         | Notes                                      |
| ---------------- | ------------------------------- | ------------------------------------------ |
| `ENV`            | `production`                    | Required                                   |
| `PORT`           | `:8080`                         | Local bind; reverse proxy in front         |
| `DB_PATH`        | `/var/lib/leilao-erp/app.db`    | Outside deploy tree                        |
| `SESSION_SECRET` | random 32+ bytes                | `openssl rand -base64 48`                  |
| `APP_URL`        | `https://seu-dominio.com`       | Canonical public URL                       |
| `TRUSTED_PROXIES`| `127.0.0.1`                     | When behind Caddy/nginx                    |
| `LOCALE`         | `pt`                            | Optional                                   |

```bash
sudo install -m 600 /dev/stdin /etc/leilao-erp/env <<'EOF'
PORT=:8080
ENV=production
APP_URL=https://seu-dominio.com
DB_PATH=/var/lib/leilao-erp/app.db
SESSION_SECRET=replace-me
TRUSTED_PROXIES=127.0.0.1
LOCALE=pt
EOF
```

### 4. systemd

Unit file: [`deploy/systemd/leilao-erp.service`](deploy/systemd/leilao-erp.service)

```bash
sudo cp deploy/systemd/leilao-erp.service /etc/systemd/system/leilao-erp.service
sudo systemctl daemon-reload
sudo systemctl enable --now leilao-erp
curl -s http://127.0.0.1:8080/health
```

WorkingDirectory must be `/opt/leilao-erp/current` (where `web/static` lives), or set `STATIC_DIR` in the env file.

### 5. HTTPS (Caddy or nginx)

**Caddy** (automatic TLS):

```
seu-dominio.com {
  encode gzip
  reverse_proxy 127.0.0.1:8080
}
```

**nginx** sketch: terminate TLS, `proxy_pass http://127.0.0.1:8080`, forward `Host` and `X-Forwarded-For` / `X-Forwarded-Proto`.

Point DNS `A` to the Lightsail static IP. Open ports 80/443 on the instance firewall.

### 6. Daily SQLite backup

```bash
sudo tee /etc/cron.d/leilao-erp-backup >/dev/null <<'EOF'
# Daily backup of SQLite DB (keep ~14 days)
15 3 * * * leilao cp /var/lib/leilao-erp/app.db /var/lib/leilao-erp/app.db.$(date +\%F).bak
20 3 * * * leilao find /var/lib/leilao-erp -name 'app.db.*.bak' -mtime +14 -delete
EOF
```

Prefer stopping writes briefly or using `sqlite3 .backup` if the DB grows large; for a small single-user ERP, file copy is usually enough.

### 7. First admin user

Production does **not** auto-seed an admin.

Options:

1. **Temporary development boot (simplest):** set `ENV=development` once, start the app so `admin@leilao.local` / `change-me-now` is created, stop, set `ENV=production`, restart, **change the password immediately**.
2. **Console:** `cais console` and call `store.CreateUser` with a bcrypt hash from `session.HashPassword`.
3. **Demo finance data:** `cais db seed` (idempotent PIX account + 22 monitores lot) — safe to re-run; does not create the admin by itself.

### 8. Post-deploy checklist

- [ ] `GET /health` → `{"status":"ok"}`
- [ ] HTTPS works; login redirects to `/dashboard`
- [ ] Admin password changed
- [ ] Backup cron present and writable path owned by `leilao`
- [ ] `SESSION_SECRET` is not the example value
