#!/usr/bin/env bash
# Redeploy leilao-erp to AWS Lightsail (~1–2 min).
#
# Usage:
#   ./deploy/deploy.sh
#   ./deploy/deploy.sh --skip-tests
#   SYNC_UPLOADS=1 ./deploy/deploy.sh   # also push local product media
#
# Env overrides:
#   SSH_KEY  SSH_HOST  REMOTE_ROOT  HEALTH_URL  KEEP_RELEASES  SKIP_TESTS  SYNC_UPLOADS
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SSH_KEY="${SSH_KEY:-$HOME/.ssh/lightsail-default-key-us-east-1.pem}"
SSH_HOST="${SSH_HOST:-ubuntu@52.73.89.19}"
REMOTE_ROOT="${REMOTE_ROOT:-/opt/leilao-erp}"
HEALTH_URL="${HEALTH_URL:-https://eletronicos.gestaobem.com/health}"
KEEP_RELEASES="${KEEP_RELEASES:-5}"
SKIP_TESTS="${SKIP_TESTS:-0}"
SYNC_UPLOADS="${SYNC_UPLOADS:-0}"

usage() {
  cat <<'EOF'
Usage: ./deploy/deploy.sh [options]

  --skip-tests       Skip go test ./...
  --sync-uploads     Include local web/static/uploads (default: keep server media)
  --keep N           Keep last N releases on server (default: 5)
  -h, --help

Env: SSH_KEY SSH_HOST REMOTE_ROOT HEALTH_URL KEEP_RELEASES SKIP_TESTS SYNC_UPLOADS
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-tests) SKIP_TESTS=1; shift ;;
    --sync-uploads) SYNC_UPLOADS=1; shift ;;
    --keep)
      KEEP_RELEASES="${2:?--keep needs N}"
      shift 2
      ;;
    -h|--help) usage; exit 0 ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

log() { printf '==> %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

[[ -f "$SSH_KEY" ]] || die "SSH key not found: $SSH_KEY"
command -v rsync >/dev/null || die "rsync required"
command -v go >/dev/null || die "go required"
command -v npm >/dev/null || die "npm required"
command -v npx >/dev/null || die "npx required"

SSH=(ssh -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new -o ConnectTimeout=15)
SCP=(scp -i "$SSH_KEY" -o StrictHostKeyChecking=accept-new -o ConnectTimeout=15)

START_TS=$(date +%s)

if [[ "$SKIP_TESTS" != "1" ]]; then
  log "go test ./..."
  go test ./... -count=1
else
  log "skipping tests"
fi

log "tailwind css"
npx tailwindcss -i ./input.css -o ./web/static/css/styles.css --minify

log "vite build"
npm run build

log "linux amd64 binary"
mkdir -p bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/server-linux ./cmd/server
[[ -x bin/server-linux ]] || die "bin/server-linux missing after build"

STAMP=$(date -u +%Y%m%d%H%M%S)
STAGE=$(mktemp -d)
ARCHIVE="/tmp/leilao-erp-release-${STAMP}.tar.gz"
cleanup() {
  rm -rf "$STAGE"
  rm -f "$ARCHIVE"
}
trap cleanup EXIT

log "stage release $STAMP"
mkdir -p "$STAGE/bin" "$STAGE/web" "$STAGE/deploy/systemd"
cp bin/server-linux "$STAGE/bin/server"
chmod 755 "$STAGE/bin/server"
# Never wipe prod product photos by default — exclude local uploads.
rsync -a --exclude uploads/ web/static/ "$STAGE/web/static/"
if [[ "$SYNC_UPLOADS" == "1" ]]; then
  log "including local uploads"
  mkdir -p "$STAGE/web/static/uploads"
  rsync -a web/static/uploads/ "$STAGE/web/static/uploads/"
fi
cp deploy/systemd/leilao-erp.service "$STAGE/deploy/systemd/" 2>/dev/null || true

tar -C "$STAGE" -czf "$ARCHIVE" .
BYTES=$(wc -c <"$ARCHIVE" | tr -d ' ')
log "package $(numfmt --to=iec "$BYTES" 2>/dev/null || echo "${BYTES}B") → $ARCHIVE"

log "upload → $SSH_HOST"
"${SCP[@]}" "$ARCHIVE" "${SSH_HOST}:/tmp/leilao-erp-release.tar.gz"

log "install + restart on server"
"${SSH[@]}" "$SSH_HOST" \
  env REMOTE_ROOT="$REMOTE_ROOT" STAMP="$STAMP" KEEP_RELEASES="$KEEP_RELEASES" bash -s <<'REMOTE'
set -euo pipefail

RELEASE="${REMOTE_ROOT}/releases/${STAMP}"
CURRENT="${REMOTE_ROOT}/current"
# Media lives outside releases so deploys never wipe product photos/videos.
SHARED_UPLOADS=/var/lib/leilao-erp/uploads

sudo mkdir -p "$RELEASE" "$SHARED_UPLOADS"
sudo tar -xzf /tmp/leilao-erp-release.tar.gz -C "$RELEASE"
sudo rm -f /tmp/leilao-erp-release.tar.gz
sudo chmod 755 "$RELEASE/bin/server"

# Seed/migrate shared uploads once (sudo required — dirs owned by leilao).
if sudo test -d "$CURRENT/web/static/uploads" && ! sudo test -L "$CURRENT/web/static/uploads"; then
  sudo rsync -a "$CURRENT/web/static/uploads/" "$SHARED_UPLOADS/"
fi
# Optional package uploads (SYNC_UPLOADS) merge into shared store.
if sudo test -d "$RELEASE/web/static/uploads"; then
  sudo rsync -a "$RELEASE/web/static/uploads/" "$SHARED_UPLOADS/"
  sudo rm -rf "$RELEASE/web/static/uploads"
fi
sudo mkdir -p "$RELEASE/web/static"
sudo ln -sfn "$SHARED_UPLOADS" "$RELEASE/web/static/uploads"
sudo chown -R leilao:leilao "$SHARED_UPLOADS"
sudo chown -R leilao:leilao "$RELEASE"
# Atomic pointer flip (current is a symlink under /opt/leilao-erp).
sudo ln -sfn "$RELEASE" "$CURRENT"
sudo chown -h leilao:leilao "$CURRENT" 2>/dev/null || true

sudo systemctl restart leilao-erp

ok=0
for _ in $(seq 1 40); do
  if curl -fsS http://127.0.0.1:8080/health >/dev/null 2>&1; then
    ok=1
    break
  fi
  sleep 0.5
done
if [[ "$ok" -ne 1 ]]; then
  echo "local /health failed after restart" >&2
  sudo systemctl status leilao-erp --no-pager -l || true
  sudo journalctl -u leilao-erp -n 40 --no-pager || true
  exit 1
fi
echo "local health: $(curl -fsS http://127.0.0.1:8080/health)"

# Prune old releases (GNU head -n -N); never delete the live one.
cd "${REMOTE_ROOT}/releases"
mapfile -t ALL < <(ls -1 | sort)
LIVE=$(readlink -f "$CURRENT" || true)
COUNT=${#ALL[@]}
if (( COUNT > KEEP_RELEASES )); then
  DROP=$((COUNT - KEEP_RELEASES))
  for ((i = 0; i < DROP; i++)); do
    old="${ALL[$i]}"
    path="${REMOTE_ROOT}/releases/${old}"
    if [[ "$(readlink -f "$path")" == "$LIVE" ]]; then
      continue
    fi
    sudo rm -rf "$path"
    echo "pruned release $old"
  done
fi

echo "live → $RELEASE"
REMOTE

log "public health $HEALTH_URL"
curl -fsS "$HEALTH_URL"
echo

ELAPSED=$(( $(date +%s) - START_TS ))
log "done in ${ELAPSED}s (release $STAMP)"
printf '   shop: %s\n' "${HEALTH_URL%/health}/"
printf '   admin: %s/login\n' "${HEALTH_URL%/health}"
