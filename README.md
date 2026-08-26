# CN Connect API

HTTP API for **CN Connect**, a corridor-based remittance platform.

Customers create transfers, submit payment proof, and track status by reference. Operators manage payment verification and payout lifecycle. The API also covers country/corridor configuration (routes, banks, mobile-money networks) and an admin operations dashboard.

**Status:** early / actively developed  
**Base path:** `/api/v1`

---

## Features

- Transfer create with server-side FX, fees, and rounding
- Idempotent create via `Idempotency-Key`
- Presigned payment-proof upload (Cloudflare R2) with server-side validation
- Explicit admin status transitions (verify, reject, process, complete, cancel)
- Country, payment-channel, and route (corridor) management
- Admin dashboard aggregations (KPIs, volume, top routes, activity)

---

## Stack

| Layer | Choice |
| --- | --- |
| Language | Go 1.26+ |
| HTTP | [chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL + [pgx](https://github.com/jackc/pgx) |
| SQL | [sqlc](https://sqlc.dev/) |
| Migrations | [goose](https://github.com/pressly/goose) |
| Validation | go-playground/validator |
| Money | [shopspring/decimal](https://github.com/shopspring/decimal) |
| Object storage | Cloudflare R2 (S3-compatible) |

---

## Domain overview

```
countries
  └── payment_channels   (BANK | MOBILE_MONEY)
routes                   (source country → destination country + FX / fees)
transfers                (rate, fee, and amounts snapshotted at create)
transfer_events          (status transition history)
```

### Transfer lifecycle

```
PENDING_PAYMENT
  → PAYMENT_RECEIVED   (payment proof confirmed)
  → VERIFYING          (payment verified)
  → PROCESSING         (payout in progress)
  → COMPLETED

Also: CANCELLED
Reject payment: PAYMENT_RECEIVED → PENDING_PAYMENT
```

- `amount_sent` — send amount in **source** currency  
- `amount_received` — payout in **destination** currency  
- `fee` / `exchange_rate` — stored on the transfer at creation (history is immutable)  
- Pending transfers include `expires_at` for the payment window  
- Catalog entities support soft delete via `deleted_at`

---

## Project layout

```
cmd/                     # API entrypoint
internal/
  app/                   # router + middleware wiring
  common/                # env, money helpers, DB error mapping
  common/httpx/          # JSON, validation, HTTP errors
  features/
    transfers/
    countries/
    dashboard/
  infra/
    db/                  # sqlc output (do not edit by hand)
    storage/             # R2 client
db/
  migrations/            # goose migrations + seed
  queries/               # SQL sources for sqlc
sqlc.yml
makefile
```

Each feature typically exposes `routes`, `controller`, `service`, and `dto` packages/files. Controllers stay thin; services own business rules; SQL is authored under `db/queries` and generated into `internal/infra/db`.

---

## Getting started

### Prerequisites

- Go 1.26+
- PostgreSQL 14+
- [goose](https://github.com/pressly/goose)
- [sqlc](https://docs.sqlc.dev/) (only when changing SQL)
- Cloudflare R2 bucket + API token (for payment-proof uploads)

### Configuration

Copy the variables below into a `.env` at the repo root:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/cnmt?sslmode=disable
HOST=
PORT=8080
LOG_LEVEL=info

OBJECT_STORAGE_ACCOUNT_ID=your_account_id
OBJECT_STORAGE_ACCESS_KEY_ID=your_access_key
OBJECT_STORAGE_SECRET_ACCESS_KEY=your_secret_key
OBJECT_STORAGE_BUCKET_NAME=your_bucket
OBJECT_STORAGE_PRESIGNED_URL_EXPIRATION=3600
```

### Database

```bash
make migrate-up
make migrate-create name=describe_change   # optional
make migrate-down                          # roll back one step
```

A seed migration (`005_seed_countries.sql`) loads sample corridors and payment channels.

### Run

```bash
go run ./cmd/api
```

The server listens on `HOST:PORT` (default `:8080`).

### sqlc

After changing `db/queries/*.sql` or migrations that affect schema:

```bash
sqlc generate
```

Commit both the SQL and the regenerated Go under `internal/infra/db`.

---

## Architecture

### Request path

```
HTTP → chi middleware → Controller → Service → sqlc Queries → PostgreSQL
                                      ↘ ObjStorage → R2 (proofs)
```

Middleware includes logging, panic recovery, path cleaning, compression, a request timeout, and CORS.

### Errors

Domain errors use shared sentinels (`BadRequest`, `NotFound`, `Conflict`, …). Postgres constraint failures are translated to HTTP-friendly responses (for example unique violations → `409`). Validation errors return per-field details. Clients never receive raw driver messages.

### Money

- All monetary values use `decimal.Decimal` — never `float64`
- Persistence helpers: 2 decimal places for money, 4 for FX rates
- Reporting endpoints group volume and fees **by currency**; amounts in different currencies are never summed into one number

### Idempotency

`POST /transfers` requires an `Idempotency-Key` header. Replays with the same key return the original create result instead of inserting a duplicate transfer.

### Soft delete

Deleting countries, routes, or payment channels soft-deletes the row. Related transfers and channels are not cascade-removed. Customer-facing lookups ignore soft-deleted / inactive parents.

---

## Typical transfer flow

1. Load corridors: `GET /transfers/options`
2. Create transfer: `POST /transfers` (+ `Idempotency-Key`)
3. Customer pays using local payment instructions
4. Request upload URL: `POST /transfers/payment-proof/upload-url`
5. Client uploads the file to the presigned R2 URL
6. Confirm proof: `PATCH /transfers/payment-proof/confirm` → `PAYMENT_RECEIVED`
7. Operator actions: verify → process → complete  
   Reject returns the transfer to `PENDING_PAYMENT`. Cancel requires a reason.
8. Track: `GET /transfers/{reference}`

Proofs are validated for presence, size (max 10 MiB), and type (JPEG, PNG, or PDF).

---

## API reference

All paths are under `/api/v1`.

### Transfers

| Method | Path | Description |
| --- | --- | --- |
| GET | `/transfers/options` | Source/destination options and channels |
| GET | `/countries/{countryID}/destinations` | Destinations for a source country |
| POST | `/transfers` | Create transfer |
| GET | `/transfers` | List / filter transfers |
| GET | `/transfers/{reference}` | Get by reference |
| POST | `/transfers/payment-proof/upload-url` | Presigned upload URL |
| PATCH | `/transfers/payment-proof/confirm` | Confirm uploaded proof |

### Admin transfers

| Method | Path |
| --- | --- |
| POST | `/admin/transfers/{id}/verify-payment` |
| POST | `/admin/transfers/{id}/reject-payment` |
| POST | `/admin/transfers/{id}/process` |
| POST | `/admin/transfers/{id}/complete` |
| POST | `/admin/transfers/{id}/cancel` |

### Catalog

| Method | Path |
| --- | --- |
| GET, POST | `/admin/countries` |
| GET, PATCH, DELETE | `/admin/countries/{id}` |
| POST | `/admin/countries/{id}/payment-channels` |
| PATCH, DELETE | `/admin/payment-channels/{id}` |
| GET, POST | `/admin/routes` |
| PATCH, DELETE | `/admin/routes/{id}` |
| POST | `/admin/routes/{id}/toggle-active` |

Route list query params: `source_country_id`, `dest_country_id`, `is_active`.

### Dashboard

| Method | Path |
| --- | --- |
| GET | `/admin/dashboard?from=YYYY-MM-DD&to=YYYY-MM-DD` |

- Default range: start of the current UTC month through end of today  
- Maximum range: 366 days  
- Response includes overview metrics, action-required transfers, daily volume by currency, status distribution, recent transfers, top routes, and recent activity  
- Metrics are aggregated in PostgreSQL

---

## Database

| Migration | Purpose |
| --- | --- |
| `001_countries.sql` | Countries, routes, payment channels |
| `002_transfers.sql` | Transfers and status enum |
| `003_idempotency_keys.sql` | Create-transfer idempotency |
| `004_transfer_events.sql` | Status transition events |
| `005_seed_countries.sql` | Sample corridors |
| `006_add_dashboard_indexes.sql` | Reporting indexes |

sqlc uses migrations as the schema source and `db/queries` as the query source.

Optional list filters commonly use:

```sql
column = COALESCE(NULLIF($n::type, ''), column)
```

This avoids ambiguous optional-NULL parameter typing with prepared statements.

---

## Development guidelines

1. Follow the existing feature package shape — don’t invent a parallel architecture for one endpoint.
2. Keep money in `decimal` end-to-end.
3. Prefer soft delete for catalog entities; avoid unintended cascades.
4. Put SQL in `db/queries` and regenerate with sqlc — no ad-hoc SQL strings in services.
5. Prefer explicit admin actions (`verify-payment`, `process`, …) over a generic status update.
6. For reporting, aggregate in SQL and keep currencies separated.

```bash
go test ./...
go build ./...
```

---

## Contributing

Issues and pull requests are welcome. Keep changes focused, match existing style, and include migrations + sqlc output when the schema or queries change.

---

## License

All rights reserved.
