# Agent rules — leilao-erp

Dense rules for coding agents. Prefer small edits, greppable names, one-shot tests.

## One-shot commands

```bash
# From repo root (worktree ok)
go test ./... -count=1
make ci                    # test + golangci-lint + prettier check
make format                # prettier
gofmt -w internal/ cmd/
npx vite build             # REQUIRED after web/src/** changes (cais dev does NOT rebuild Svelte)
npx tailwindcss -i ./input.css -o ./web/static/css/styles.css
```

Dev server: `PORT=:8080 ENV=development cais dev` → http://127.0.0.1:8080  
Admin (dev only): `admin@leilao.local` / `change-me-now`

## Code style

- Functions: 4–20 lines when practical. Files: **under 500 lines** (target 200–300).
- One responsibility per file. Prefer `items.go` / `sales.go` over god `crud.go`.
- Names greppable and domain-specific: `CreateSale`, `ListItemsInStock`, `isAccessoryTitle`. Avoid bare `handler`, `data`, `util`.
- Explicit types on exported Go APIs. Cents are `int64`. Money formatting lives in `internal/domain`.
- Early returns; max ~2 nesting levels for control flow.
- Errors: wrap with `fmt.Errorf("…: %w", err)` and include ids/values (`item %d not in stock`).
- Keep WHY comments and issue refs. Do not strip intent comments on refactor.
- Format with `gofmt` + `prettier`. Do not bikeshed style.

## Architecture (predictable paths)

| Path                | Role                                                                  |
| ------------------- | --------------------------------------------------------------------- |
| `cmd/server`        | entry                                                                 |
| `internal/app`      | routes, bootstrap                                                     |
| `internal/handlers` | HTTP + Inertia props                                                  |
| `internal/store`    | SQLite, migrations in `store/migrations/`                             |
| `internal/domain`   | pure money/margin/allocate (no DB)                                    |
| `internal/models`   | structs only                                                          |
| `internal/db`       | seeds (`RunSeeds` idempotent)                                         |
| `web/src/pages`     | Inertia Svelte pages (`Sales/New` → `web/src/pages/Sales/New.svelte`) |
| `web/src/lib`       | pure JS helpers (filter/sort)                                         |
| `web/static/build`  | **gitignored** Vite output — always rebuild after FE change           |

Money: integer **cents** BRL. Net = gross − fee − shipping. Margin = net − unit_cost_at_sale (sum of sale_lines).

## Domain invariants (do not break)

- `CreateSale` may take main `ItemID` + `AccessoryIDs`; writes `sale_lines`; freezes total cost.
- Cancel pending sale restores **all** line items to `in_stock`.
- Sold items cannot be sold again; lot status open/partial/sold.
- Multi-line JSON forms: `parseFormOrJSON` expands arrays; use `accessory_ids` as array or CSV.
- Stock UI classifies accessories by title keywords (`cabo`, `vga`, `hdmi`, `força`) via `isAccessoryTitle`.

## Frontend / Cais gotchas

- **`cais dev` does not rebuild Svelte.** After any `web/src/**` edit run `npx vite build` or UI stays stale.
- SPA entry is `web/static/build/assets/main.js` (stable name). Bump `?v=` in `web/templates/app.html` or hard-refresh; templates are `//go:embed` — rebuild Go after template change.
- Inertia root: `web/templates/app.html`. CSRF: cookie `cais_csrf` + header `X-CSRF-Token` (see `web/src/main.js`).
- Svelte 5: `mount()` not `new App`. Prefer defensive `Array.isArray` on form arrays (`accessory_ids`).

## Tests

- `go test ./...` must pass with no external services.
- New store/handler behavior → test in same package (`store`, `handlers`).
- Acceptance path: seed monitores → sell → settle (`TestAcceptance_SeedSellSettle`).
- Seeds are idempotent; do **not** bulk-rename all monitors to a model (set models per unit in DB/UI).

## Defensive programming (this app)

- [x] Timeouts: HTTP via stdlib defaults; keep requests short.
- [ ] Retries: not required for local SQLite.
- [x] Auth: `RequireAuthFunc` on all finance routes; no public signup.
- [x] Input validation: BRL parse + payment_status rules before store writes.
- [x] SQL: parameterized queries only; no string-concat SQL.

## Do not

- Commit `data/*.db` or `web/static/build/`.
- Grow `internal/store/crud.go` or `handlers/sales.go` past 500 lines — split by domain.
- Invent multi-tenant or cloud deploy unless asked.
