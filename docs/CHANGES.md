# Production pass

What changed when the repo moved from "15 processes on a VM" to "one image on Kubernetes", and every bug fixed on the way.

## Packaging

- Removed `setup.sh`, `run.sh`, `stop.sh` and the 13 per-service `cmd/main.go` files.
- One binary, `cmd/marketplace`, with `serve`, `migrate`, `seed`, `healthcheck`, `version`.
- Migrations are embedded and tracked in `schema_migrations` with checksums, behind a Postgres advisory lock so parallel init containers are safe. `serve` refuses to boot with pending migrations.
- Multi-stage Dockerfile: node builds the storefront, rust:alpine builds static musl engines, Go builds a static binary with the storefront embedded, final image is `distroless/static:nonroot`.
- The gateway reverse proxy is gone. All Go routers mount in one chi router under `/api`, so there are no internal HTTP hops.

## Security fixes

| Problem | Fix |
|---|---|
| `POST /register` accepted `"role":"admin"` | only `buyer` or `seller` can be chosen |
| JWT parsing accepted any algorithm | HS256 only, issuer and expiry required, in every service middleware |
| `PUT /stock`, `/reserve`, `/commit`, `/release` were public | internal calls are in-process now; stock edits need a seller who owns the variant |
| `/orders/{id}/paid` and `/cancel` were public | removed; only checkout moves order state, through a guarded state machine |
| `POST /charge` let the client choose amount and seller splits | removed from the API; checkout charges the server-computed total |
| Cart and order trusted client `unit_price`, `seller_id`, `commission_bps` | all priced from the database |
| `/fraud/block` and `/search/index` were exposed through the gateway | engines are internal, behind `INTERNAL_TOKEN` and a NetworkPolicy |
| `/notifications/send` was public | admin only |
| `/ledger/balance/{account}` and `/sagas/{id}` were public | admin only, and checkout status is owner only |
| Review "helpful" votes were unauthenticated and unlimited | one vote per user, not on your own review |
| Refresh tokens stored in plaintext and never rotated | stored as SHA-256, rotated on every refresh, reuse revokes the whole session family |
| Email verification codes could be brute forced across all users | scoped to the signed-in user |
| Login leaked whether an email exists through timing | constant work with a dummy hash |
| argon2 used 64 MiB per hash with no concurrency cap (memory DoS) | OWASP parameters (19 MiB, t=2) with a hash semaphore; old hashes still verify |
| Database password hardcoded as a fallback in 12 files | removed; `DATABASE_URL` and a 32+ char `JWT_SECRET` are required |
| Rate limiter keyed on spoofable `X-Forwarded-For`, reset everyone every minute | per-IP token bucket, proxy headers only trusted when configured, stricter limit on login and register |
| `Access-Control-Allow-Origin: *` with credentials headers | CORS off by default, allow list via `CORS_ORIGINS` |
| No security headers | CSP, frame deny, nosniff, referrer policy, HSTS behind TLS |
| No request body limit | 1 MiB default |
| Admins could only log moderation, nothing happened | remove, suspend and reinstate now take effect in the same transaction as the audit row |
| Seller public profile exposed Stripe account and user id | public view only shows store fields |
| Open redirect via `?next=` | only same-site paths, backslashes rejected |
| Vulnerable `x/crypto` and `react-router` | upgraded, `govulncheck` and `npm audit` clean |

## Correctness fixes

- **Checkout did nothing.** The saga ran no-op steps. It now prices the cart, creates the order, scores fraud, reserves stock, charges, commits stock and marks the order paid, compensating in reverse on failure.
- Multi-line reservations are one transaction, ordered by variant to avoid deadlocks, guarded per order so a retry cannot double reserve.
- Commit and release lock held reservations, so they are idempotent and a reservation that expired cannot be committed.
- The expiry sweeper used select-then-update outside a transaction; it now uses `FOR UPDATE SKIP LOCKED` so many replicas can run it.
- Payment refunds write reversing ledger entries; replaying an idempotency key with a different amount is rejected; the ledger insert verifies entries sum to zero.
- Interrupted sagas (pod killed mid-checkout) are recovered by a background job.
- `products.category_id` is nullable but was scanned into `int64`; creating a product without a category violated the foreign key.
- Gateway routed `/products` to catalog which only served `/catalog/products`; review routes collided with catalog.
- Product listings had no price, rating or stock; the storefront could not render a card without N extra calls.
- Catalog `stock` column and inventory `stock` table drifted; inventory is the source of truth and both are updated together.
- Re-indexing a product in the search engine duplicated its postings; added replace and delete. Search falls back to Postgres when the engine is down or empty.
- Fraud engine kept velocity history forever; now pruned.
- Duplicate email or slug races returned 500; they are mapped to 409 or retried.
- Internal errors are logged with request ids instead of being swallowed.

## Operations

- `/healthz` and `/readyz` on every binary. Readiness turns 503 on SIGTERM, then the server drains.
- JSON logs with request ids, Prometheus metrics on `:9090`.
- HTTP read, write, header and idle timeouts, DB pool limits.
- Tests run against real Postgres, each test in its own schema, with `-race`.
