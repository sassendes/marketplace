# SASSENDES MARKETPLACE — MASTER REFERENCE

> Update 2026-10: the system now ships as one image and one binary (`cmd/marketplace`). The VM scripts are gone, services mount in-process under `/api`, and `platform/` holds infrastructure code only. See `README.md` for running it and `docs/CHANGES.md` for everything that was fixed.

> The single source of truth. If context is lost, if you forget, if a new session starts — READ THIS. Everything about what this system is, how it's built, why every decision was made, and what's done vs pending lives here. This is not a pep-talk bible and not API docs. It's the engineering ground truth.

---

## 0. WHAT THIS IS

A production-grade, multi-seller marketplace (eBay/Amazon model). Sellers list products, buyers buy them, money flows through the platform to sellers minus commission. Built as a Go monorepo (one `go.mod` named `marketplace`), with Rust for the CPU-bound engines (search, fraud). Deployed on Kubernetes, fronted by a real network layer (opnsense/wireguard) on a VPS. Nothing faked, nothing demo — real integrations in test/sandbox mode where money/identity legally require it (Stripe test mode).

**The bar:** if a senior engineer read the code, it should look like a real product, not a tutorial. Correct data modeling, proper layering, real concurrency handling, real money handling.

---

## 1. NON-NEGOTIABLE ARCHITECTURE RULES

These are LOCKED. Do not drift. Drifting mid-build is how it becomes spaghetti.

1. **Monorepo, one `go.mod`** named `marketplace`. Every service is a package path under it.
2. **Each service is self-contained** with this exact shape:
   ```
   service-name/
     cmd/main.go              entry point, wires everything, starts server
     model/                   structs: domain entities + request/response shapes
     repository/              DB queries ONLY (no logic)
     service/                 business logic ONLY (no HTTP, no raw SQL)
     handler/                 HTTP layer ONLY (decode req, call service, write resp)
     middleware/              per-service middleware (auth check, etc.)
   ```
3. **Full folder names**: `repository` and `service`, never `repo`/`svc`. Consistency across all services.
4. **NO shared/ folder.** Each service that needs JWT verification has its own small `middleware/` (~30 lines duplicated). Small duplication > cross-service coupling. This was decided and locked.
5. **The 3-layer rule, never mixed:** handler (HTTP) → service (logic) → repository (DB). A bug in a query lives in repository. A bug in a rule lives in service. Always know where.
6. **Money is ALWAYS int64 in the smallest unit (cents).** $19.99 = `1999`. NEVER float. Floats lose cents and make the ledger unreconcilable. Currency code always rides alongside the amount.
7. **Interfaces at boundaries** where it helps testing/swapping (service depends on a repository interface, not the concrete type) — introduce as needed, don't over-abstract early.
8. **Context first arg** on every function that does I/O: `func X(ctx context.Context, ...)`.
9. **Errors bubble up, handled once** — repository/service return errors, the handler maps them to HTTP status codes. Predefined error vars (`ErrEmailTaken`) so handlers can `errors.Is` and pick the right code.
10. **Stateless services** — no user state in memory; state lives in Postgres/Redis. Any replica handles any request.
11. **Language split:** Go for all I/O-bound services + the future operator. Rust for CPU-bound/latency-critical engines ONLY: search/ranking, fraud detection. NOT for the oversell problem (that's DB-locking, not language speed).
12. **Crypto is never hand-rolled.** Use vetted libs: `golang.org/x/crypto/argon2` (hashing), `github.com/golang-jwt/jwt/v5` (JWT). The github import paths are library addresses, not a dependency on github at runtime.

---

## 2. THE MONEY RULE (expanded — this is the thing that makes it real)

- Store all money as `int64` cents. Display divides by 100.
- Every amount has a currency (ISO 4217 string: "USD", "EUR", "TND").
- The **double-entry ledger** is the source of truth for balances — never a single `balance` column. Every movement is two entries (a debit and a credit) that sum to zero. You reconstruct any balance by summing its ledger entries. This is how you can PROVE you never lost or invented money.
- Example flow (buyer pays 10000 = $100.00 for an item, 10% commission):
  - buyer_payment: debit buyer_wallet 10000 / credit platform_clearing 10000
  - commission: debit platform_clearing 1000 / credit platform_revenue 1000
  - seller_owed: debit platform_clearing 9000 / credit seller_payable 9000
  - payout: debit seller_payable 9000 / credit seller_wallet 9000
  - At every step, debits == credits. Platform_clearing nets to zero once settled.
- Idempotency on every money mutation: an idempotency key stored BEFORE acting; a duplicate request returns the first result instead of charging twice.

---

## 3. SERVICES — THE FULL MAP

Status legend: [DONE] built+verified · [WIP] in progress · [TODO] not started

### auth [DONE — compiles, vets, builds, runs]
Identity for buyers, sellers, admins.
- Endpoints: POST /register, /login, /refresh, /logout (email verify pending)
- JWT access (15min, stateless) + refresh (7d, stored in DB so it can be revoked)
- argon2id password hashing (salt$hash format, constant-time compare)
- RBAC roles: buyer, seller, admin
- Files: model/{user,token,request}.go, repository/{user_repo,token_repo}.go, service/{auth_service,jwt}.go, handler/auth_handler.go, middleware/auth.go, cmd/main.go
- PENDING: email verification handler, SQL schema, env-based config (secret + DSN are hardcoded placeholders right now)

### catalog [DONE — compiles, vets, builds, RUNS end-to-end against real Postgres]
Products, variants, categories, images (image records only; bytes handled by media service).
- THE MODEL (corrected per Amazon research):
  - **categories**: self-referencing tree (id, parent_id, name, slug). Electronics → Audio → Headphones.
  - **products** (PARENT): id, seller_id, title, description, slug, category_id, brand, status, timestamps. The concept. NOT directly sellable. No price/stock here.
  - **product_variants** (CHILD/SKU): id, product_id, sku, price (int64 cents), currency, stock, attributes (JSONB: {"size":"M","color":"red"}), active. THE sellable unit. Price and stock live HERE.
  - **product_images**: id, product_id or variant_id, storage_key, position, is_primary.
- Every product has ≥1 variant (a "master variant" even if no real options).
- Flexible attributes via **JSONB** on variant (pragmatic choice over EAV — EAV is "more correct" but a 4-join query nightmare; JSONB indexes fine in Postgres).
- Slugs generated from title (lowercase, dashes, strip specials, append short id on collision) — SEO URLs like /product/sony-wh1000xm5.
- Endpoints: POST /products (seller), GET /products (list+filter+paginate), GET /products/{slug}, PUT /products/{id}, DELETE /products/{id} (soft delete via active flag), + variant + category endpoints.
- Emits events (later, needs NATS): product.created/updated/deleted → consumed by search + media.
- PASS 1 (now): products+variants+categories+slug+JSONB+CRUD+filter+pagination+error contract.
- PASS 2 (later): event emission, image records, CSV bulk import.
- VERIFIED 2026-08-23: register→login→create category→create product w/ 2 variants→get by slug→search all tested live via curl against real Postgres. Parent-child + JSONB attributes + slug gen + cross-service JWT all confirmed working.

### seller [DONE — compiles, vets, builds, RUNS end-to-end against real Postgres]
Onboarding, profiles, Stripe Connect account creation, payout tracking, commission config. Sellers = users with role=seller. Verification state machine (pending -> active).
- Commission stored as BASIS POINTS (1000 = 10%). 
- VERIFIED 2026-08-23: onboard->pending, connect-stripe->active+payout_enabled, duplicate onboard->409, all live via curl.

### cart [DONE — compiles, vets, builds, RUNS end-to-end against real Postgres]
Multi-seller cart. Add/remove/update items, quantity-merge on re-add (ON CONFLICT upsert), per-item seller_id + price snapshot for the eventual order split. Tied to a buyer. NOTE: built on Postgres (cart_items table) not Redis for now — durable and simpler; can move hot-path to Redis later. Subtotal computed in cents.
- VERIFIED 2026-08-23: empty cart, add multi-seller items, merge same variant, update qty, remove, clear — all live via curl.

### inventory [DONE — compiles, vets, builds, PROVEN no-oversell under 20-way concurrency]
Per-variant stock (on_hand, reserved; available = on_hand - reserved). The NO-OVERSELL problem solved with Beher's approach B: atomic conditional UPDATE `SET reserved = reserved + n WHERE on_hand - reserved >= n`. Postgres serializes the updates on the row; only one affects a row, the rest affect zero rows -> insufficient stock. No locks, no read-then-write gap.
- Full reservation lifecycle: reserve (atomic hold) -> commit (payment ok: on_hand & reserved both drop) / release (payment fail: hold returned). 15-min TTL. Background sweeper auto-releases expired holds every minute.
- VERIFIED 2026-08-23: set stock=1, fired 20 CONCURRENT reservations -> exactly 1 held (201), 19 rejected (409), final reserved=1 available=0. NO OVERSELL. This is the proof.

### order [DONE — compiles, vets, builds, RUNS end-to-end against real Postgres]
Checkout. Multi-seller SPLIT: one cart with items from 3 sellers → 3 sub-orders → tracked as ONE buyer-facing order with 3 fulfillment states. Order history, status tracking.

### payment [DONE — compiles, vets, RUNS. Idempotent charge (same key never double-charges, proven with 6 concurrent calls -> 1 payment/4 ledger rows). Double-entry ledger SUMS TO ZERO (buyer -4500, seller10 +3600, seller20 +450, platform +450). Balance = SUM(ledger), never stored. Test-Stripe provider, swap for real client later.]
Real Stripe integration (TEST MODE). Payment intents, confirmation, webhooks (Stripe tells you async when charges settle/fail). Idempotency keys. Stripe Connect for marketplace payouts (buyer→platform→seller, minus commission). THE double-entry ledger. Refunds that claw back from seller balance.

### pipeline [DONE — compiles, vets, RUNS. Saga orchestrator: reserve->charge->confirm, each with compensation. On failure, completed steps compensated in REVERSE. Proven via Go tests against real Postgres: happy path completes; charge-failure compensates reserve (release), saga=failed, step=compensated.]
The SAGA orchestrator. NATS events. Flow: order → reserve stock → charge payment → confirm → schedule payouts → notify. Any step fails → COMPENSATE (release stock, refund, cancel). State machines, compensating actions, retries, idempotency per step, dead-letter queue for repeatedly-failing events.

### search [DONE — Rust, compiles, RUNS. Inverted index + BM25 ranking. Verified: relevant doc ranks first, irrelevant excluded, 2 unit tests pass. HTTP /index + /search.]
Inverted index (tokenize title/desc → term → product IDs). Ranking: BM25/TF-IDF + business signals (seller rating, in-stock, recency, sponsored). Typo tolerance (edit distance/trigrams). Autocomplete (prefix trie). Faceted search (category+price+rating with live counts). Consumes catalog events to index. CPU-bound + latency-critical = why it's Rust.

### fraud [DONE — Rust, compiles, RUNS. Velocity tracking (tx/10min) + geo mismatch + new-account + amount risk rules -> score 0-100 -> approve/review/decline. 4 unit tests pass. HTTP /score + /block.]
Real-time transaction scoring in the checkout path. Velocity checks (orders per user/IP/card per hour), pattern matching (billing≠shipping, new account + high value), rules engine, scoring. Latency-critical = Rust.

### media [DONE — compiles, vets. Pre-signed upload pattern: service stores metadata + object key, issues upload URL, never handles bytes. pending->ready on confirm. content-type allowlist.]
Image upload + processing. Seller gets a PRE-SIGNED upload URL (S3/MinIO) — browser uploads DIRECTLY to storage, API never touches bytes (scales). Then a worker (Go or Rust): validate real image, strip EXIF (privacy), re-encode to WebP/AVIF, generate multiple resolutions (thumb 150 / card 400 / full 1200), compress. Store processed → serve via CDN. Consumes product.created events.

### recommendations [TODO — Go or Rust]
"Customers also bought", "similar", "recently viewed". Co-occurrence matrices, collaborative filtering.

### reviews [DONE — compiles, vets, builds, RUNS end-to-end against real Postgres]
Product reviews with 1..5 rating, one-per-user-per-product (UNIQUE constraint + service check), rating summary (average + star breakdown), helpfulness counter, sorted helpful-first. verified_purchase flag present, wired to order-history later. 
- VERIFIED 2026-08-23: create, duplicate->409, invalid rating->400, list, summary math (avg 4.0 from 5+3), mark-helpful — all live via curl.

### notification [DONE — compiles, vets, builds, RUNS end-to-end against real Postgres]
Stores every notification then attempts delivery (store-then-send so a failed delivery is recorded, not lost). deliver() is the pluggable point — logs for now, swap in SMTP/Twilio later. Internal POST /notifications/send for other services; user-facing GET /notifications + mark-read (auth). Channels: email/sms/in_app.
- VERIFIED 2026-08-23: send (email + in_app), list newest-first, mark-read, no-auth 401 — all live via curl.
- LATER: consume NATS events instead of direct POST; this is where auth email-verify codes get delivered.

### admin [DONE — compiles, vets. Role-gated (admin only). Append-only moderation audit log + platform stats. remove/suspend/reinstate on product/seller/review.]
Platform operator: approve sellers, resolve disputes, ban actors, see all orders/money.

### gateway [DONE — compiles, vets. Reverse proxy: one public front door, prefix-routes to all 12 services. Per-IP fixed-window rate limiting. /health.]
One entry point (BFF). Routing, auth check, rate limiting (token-bucket, Redis-backed, distributed), CORS, request_id injection — all in one place. Frontend talks only to the gateway.

---

## 4. CROSS-CUTTING (build these patterns EARLY, retrofitting across 16 services is hell)

- **Consistent error contract:** every API error returns `{"error":{"code":"PRODUCT_NOT_FOUND","message":"...","request_id":"..."}}`. Frontend switches on `code`. `request_id` ties the user error to your logs/traces.
- **request_id middleware:** every request gets a unique ID, logged, echoed in errors and response headers.
- **Pagination:** every list endpoint returns `{items, total, page, page_size}`. URL reflects page (?page=3).
- **Idempotency layer:** Redis-stored keys on every create the client might retry (order, payment).
- **Rate limiting:** per user/IP/endpoint, Redis-backed so it works across replicas.
- **Audit log:** every money movement, seller action, admin action — logged immutably.
- **Health/readiness + graceful shutdown:** every service. Finish in-flight requests on shutdown.
- **Slugs + SEO:** slug per product, sitemap.xml, per-page meta/OpenGraph, JSON-LD structured data, 301 redirects when slugs change.
- **Empty/error/loading states** (frontend): custom 404/500/403/503, offline state, skeletons, empty-cart/no-results.

---

## 5. INFRA (later, on the VPS / Linux VMs)

- Postgres via cloudnative-pg (HA, RLS where useful), Redis, NATS (event bus + dead-letter), MinIO (images/backups).
- Kubernetes: kustomize base+overlays, netpols (default-deny + explicit), cert-manager (mTLS between services, TLS ingress), HPA, PDB, resource limits.
- Observability: Prometheus + Grafana + Loki + Tempo + Alertmanager. Per-service instrumented, tenant/seller-tagged metrics.
- API gateway (traefik/nginx or Go), edge rate limiting, WAF.
- The NETWORK LAYER (sami's mandate, samcloud's daily world): opnsense/pfsense router, VLANs (guest/admin/server/DMZ segments), wireguard tunnel (VPS public IP ↔ home lab compute), firewalling (default-deny between segments), DMZ for public-facing frontend, VRRP for gateway HA.
- CI/CD: build → test → scan images → deploy → progressive rollout → rollback.
- Provisioning: `./provision-*.sh` style scripts for repeatable setup.
- go.mod init + go mod tidy happens on the Linux VM, not on Windows dev box.

---

## 6. BUILD ORDER (never jump ahead — each compiles+runs before the next)

1. Repo + skeleton ✅
2. auth ✅
3. catalog (products+variants+categories) ← WIP
4. seller
5. cart (Redis enters)
6. inventory + no-oversell (first hard concurrency)
7. order (multi-seller split)
8. payment + Stripe test mode + ledger (the money engine)
9. pipeline saga (NATS, compensation)
10. search (Rust) — can be built in parallel any time as a language change
11. media pipeline
12. fraud (Rust), recommendations, reviews, notification, admin
13. gateway
14. frontend (JS)
15. infra deploy (k8s, network layer, observability) — build it, break it, fix it

---

## 7. HOW EACH SERVICE GETS BUILT (the rhythm)

Per service, plan 5 points first: what it does (1 sentence), endpoints, data model, what it talks to, the hard part. Then build BOTTOM-UP: model → repository → service → handler → cmd/main → verify it compiles+builds+runs. Files created as you climb the layers, not all upfront. Beher hand-types every line (no paste) so it's his and he understands it. Deep-understanding can lag on pass 1 (get syntax comfy, momentum) and lands on pass 2 when modifying. Verify each service in a sandbox (`go build ./...` + `go vet ./...` + build the binary + run it — DB-connection-refused is EXPECTED proof the wiring ran).

---

## 8. WHAT'S PROVEN vs UNPROVEN (be honest about this)

PROVEN (compiled + ran in sandbox):
- auth: all 10 files compile, go vet clean, binary builds (10MB), starts, reaches db.Ping() (fails only on no-DB, which is expected).

UNPROVEN (needs the VM/DB to verify):
- Any SQL actually matching real tables (no schema applied yet).
- Full request cycle (register → row in DB → login → JWT back).
- Everything catalog and beyond.

DON'T claim something works end-to-end until it's run against a real DB. "Compiles" ≠ "runs correctly."

---

## 9. THE PEOPLE / CONTEXT (so a fresh session gets it)

- Beher: 16, self-taught ~8mo, Tunisia, GitHub sassendes. Targeting Dec 2026 placement at SAMCloud (MSP in Plaisir, France, run by brother-in-law/mentor Sami Fantar). Trilingual, competitive R6 player.
- This marketplace is his MAIN — the project of his life, maintained + expanded long-term, not a show-and-abandon.
- Sami's mandate: master the network layer (opnsense/wireguard/firewalling) — samcloud uses it daily. That's why the infra section has the network layer; the marketplace deploys onto it.
- Working style: goes HARD and fast, hand-types everything, wants correctness not hand-holding, calls out fluff, real prod or nothing.

---

END OF MASTER REFERENCE. Keep this in /docs/MASTER.md. Update the [DONE]/[WIP]/[TODO] statuses as services complete. When context is lost, this file is the truth.


### frontend [DONE — Amazon-style storefront (white/clean, no gradients). Gateway URL is a baked-in constant (shoppers never see it). Slug-based product URLs (?p=red-running-shoes) resolving via catalog /products/{slug}. Custom 404/500 error pages served by the gateway. Product grid, search (->Rust), cart, JWT auth, checkout (->order).]
