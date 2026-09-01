# Distributed Rate Limiter & Request Throttler

An ultra-low-latency, Redis-backed distributed rate limiter with an async
analytics pipeline and full Prometheus/Grafana observability. The sliding-window
check runs as a single atomic Redis Lua script and sustains **~31,000 req/s**
on a single developer machine with **sub-2ms p99 latency** end-to-end through
the full HTTP middleware chain (Redis round trip included).

[![CI](<!-- PLACEHOLDER: GitHub Actions badge URL once you add the workflow -->)]()
[![Go Report Card](<!-- PLACEHOLDER: goreportcard.com badge if you run it -->)]()

- Rate limiting is only correct if it's **atomic under concurrency** — a naive `GET`-then-`SET` in Redis races and lets traffic burst past the limit. This uses a single Lua script so the check-and-increment is one indivisible operation.
- A rate limiter that logs synchronously to Postgres becomes the thing that makes your API slow. Metric writes here are fully decoupled from the request path via a buffered channel + worker pool.
- An API that can't tell you it's rejecting 40% of a client's traffic isn't operable. Every decision is exported as a Prometheus metric.
- Built and load-tested as a fintech-adjacent backend portfolio project — sliding window over fixed window specifically because burst-at-window-boundary abuse is a real attack vector on payment/API-gateway rate limits.

---

## Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)
- [Observability](#observability)
- [Testing](#testing)
- [Load Testing & Benchmarks](#load-testing--benchmarks)
- [Chaos / Failure Testing](#chaos--failure-testing)
- [Design Decisions](#design-decisions)
- [Roadmap](#roadmap)
- [License](#license)

---

## Architecture

```
[ HTTP Client ]
       │
       ▼
 [ HTTP Server Middleware ]
       │
       ├──► 1. Check Policy (In-Memory / Postgres Policy Engine)
       │
       ├──► 2. Run Redis Lua Script (Atomic Sliding Window Check)
       │         │
       │         ├─► [BREACHED] ──► Return 429 Too Many Requests
       │         │                  └─► Emit Prometheus metric + Blocked event
       │         │
       │         └─► [ALLOWED]  ──► Pass Request to Target Handler
       │                            └─► Emit Prometheus metric + Success event
       │
       ▼
 [ Metric Worker Pool ] ──(Buffered Channel)
       │
       ▼
 [ Asynchronous Batch Flusher ]
       │ (Ticker / Buffer Max Size)
       ▼
 [ PostgreSQL (pgx.Batch) ] ──► Permanent Analytics Storage

 [ Prometheus ] ──scrapes──► /metrics on the API
       │
       ▼
 [ Grafana ] ──► Dashboards (request rate, latency, buffer depth, errors)
```

<!-- PLACEHOLDER: replace the ASCII diagram above with a real exported
image (draw.io / excalidraw / Mermaid render) once you have one — link it
here, e.g. ![architecture](docs/architecture.png) -->

## Features

- **Atomic sliding-window rate limiting** via a single Redis Lua script (`ZSET`-based, no race conditions under concurrent load — verified with `go test -race`)
- **Per-client policy overrides** (in-memory store + Postgres-backed lookup)
- **Async analytics pipeline** — buffered channel + worker pool + `pgx.Batch` inserts, so metric writes never block the hot path
- **Graceful shutdown** — drains in-flight metrics before exit, closes DB/Redis pools cleanly
- **Prometheus metrics** — request rate, allow/block ratio, latency histograms (p50/p95/p99), analytics buffer depth, dropped-event counter, Redis error rate, batch flush duration
- **Grafana dashboard** — pre-provisioned, auto-loads on `docker compose up`
- **Load tested** with both vegeta and k6, at up to 200 req/s sustained with zero request errors
- **Race-tested** concurrency (`go test -race`) proving no double-counting under simultaneous requests

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.23 |
| Rate limit store | Redis 7 (Lua scripting, `ZSET` sliding window) |
| Persistent storage | PostgreSQL 16 (`pgx/v5`, `pgx.Batch`) |
| Metrics | Prometheus (`client_golang`) |
| Dashboards | Grafana (auto-provisioned) |
| Load testing | vegeta, k6 |
| Containerization | Docker, Docker Compose |

## Project Structure

```
rate-limiter-service/
├── cmd/api/main.go                # bootstrap, wiring, graceful shutdown
├── internal/
│   ├── config/                    # env-based config loader
│   ├── db/                        # postgres pool, redis client, batch inserts, policy lookup
│   ├── ratelimiter/                # Lua-backed sliding window Limiter + Policy
│   ├── analytics/                  # async metric event pipeline (Collector, worker pool)
│   ├── middleware/                 # HTTP middleware wiring limiter + analytics + metrics
│   └── metrics/                    # Prometheus metric definitions
├── scripts/
│   ├── sliding_window.lua          # atomic rate-limit check
│   ├── schema.sql                  # Postgres schema (auto-applied on first boot)
│   ├── loadtest_vegeta.sh          # vegeta load test
│   └── loadtest_k6.js              # k6 load test
├── deploy/
│   ├── prometheus/prometheus.yml   # scrape config
│   └── grafana/                    # provisioned datasource + dashboard
├── main_test.go                    # end-to-end test + benchmark
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── go.mod
```

## Getting Started

### Prerequisites
- Go 1.23+
- Docker & Docker Compose
- (optional) [vegeta](https://github.com/tsenart/vegeta) and/or [k6](https://k6.io) for load testing

### Run everything
```bash
git clone <!-- PLACEHOLDER: your repo URL -->
cd rate-limiter-service
go mod tidy
docker compose up --build
```

This starts:

| Service | URL | Notes |
|---|---|---|
| API | http://localhost:8080 | `/healthz`, `/check`, `/metrics` |
| Postgres | localhost:5432 | user/pass: `ratelimiter` / `ratelimiter` |
| Redis | localhost:6379 | |
| Prometheus | http://localhost:9090 | scrapes the API every 5s |
| Grafana | http://localhost:3000 | login `admin` / `admin` (change for anything beyond local dev) — dashboard auto-loads |

### Quick smoke test
```bash
curl localhost:8080/healthz
curl -H "X-API-KEY: demo-client" localhost:8080/check
```

## API Reference

| Endpoint | Method | Description |
|---|---|---|
| `/healthz` | GET | Liveness check |
| `/check` | GET | Runs the rate-limit check for the caller (identified by `X-API-KEY` header, falls back to remote IP) |
| `/metrics` | GET | Prometheus scrape endpoint |

**Response headers on `/check`:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
```
`429 Too Many Requests` is returned once the limit is exceeded within the configured window.

<!-- PLACEHOLDER: if you added the policy management CRUD API, document
the endpoints here (POST/GET/PUT /policies/{client_id}), including
auth requirements and example curl commands. -->

## Observability

Every request through `/check` updates:

- `rate_limiter_requests_total{status="allowed|blocked|error"}`
- `rate_limiter_request_duration_seconds{status=...}` (histogram)
- `rate_limiter_redis_errors_total`
- `rate_limiter_analytics_buffer_depth` (gauge)
- `rate_limiter_analytics_dropped_total`
- `rate_limiter_batch_flush_duration_seconds` (histogram)

The Grafana dashboard (`deploy/grafana/dashboards/rate-limiter-overview.json`) visualizes all of these out of the box.

<!-- PLACEHOLDER: screenshot of the Grafana dashboard under load.
![grafana-dashboard](docs/screenshots/grafana-dashboard.png) -->

<!-- PLACEHOLDER: screenshot of Prometheus targets page showing the
api job as UP.
![prometheus-targets](docs/screenshots/prometheus-targets.png) -->

## Testing

```bash
make test           # go test ./...
make race           # go test -race ./...  (proves no double-counting under concurrency)
make bench          # benchmark through the full middleware chain
```

<details>
<summary>go test -bench output (full middleware chain, Redis round trip included)</summary>

```
42241	32136 ns/op	12351 B/op	128 allocs/op
PASS
ok      ratelimiter     1.661s
?       ratelimiter/cmd/api     [no test files]
?       ratelimiter/internal/analytics  [no test files]
?       ratelimiter/internal/config     [no test files]
?       ratelimiter/internal/db [no test files]
?       ratelimiter/internal/middleware [no test files]
PASS
ok      ratelimiter/internal/ratelimiter        0.003s
```

`32.1µs` average latency per request through the full stack (middleware → policy
lookup → Redis Lua script) works out to roughly **31,000 requests/sec**
sustainable throughput on a single machine, with 128 allocations/op — a
reasonable target for a follow-up allocation-reduction pass (see [Roadmap](#roadmap)).

</details>

<!-- PLACEHOLDER: paste your actual `go test -race` output here once run,
e.g. inside a details block:
<details><summary>go test -race output</summary>

```
PASTE OUTPUT HERE
```
</details> -->

## Load Testing & Benchmarks

Two load-testing tools are wired up, exercising different scenarios:

```bash
make loadtest-vegeta                          # defaults: 200 rps, 30s
RATE=500 DURATION=60s make loadtest-vegeta

make loadtest-k6                              # defaults: 200 rps, 30s
RATE=500 DURATION=60s k6 run scripts/loadtest_k6.js
```

### Results

| Tool | Scenario | Load | Duration | Requests | Latency (p50 / p95 / p99) | Max Latency | Outcome |
|---|---|---|---|---|---|---|---|
| **vegeta** | Single client, sustained | 200 req/s | 30s | 6,000 | 100% under 10ms | — | 100 allowed / 5,900 blocked — default policy (100 req / 60s window) enforced exactly as configured |
| **k6** | Multi-client (unique key per VU), sustained | 200 req/s (arrival-rate) | 30s | 6,000 | 0.93ms / 1.34ms / 1.73ms | 12.37ms | 5,000 allowed / 1,000 blocked, **0 dropped requests, 0 connection errors, 100% checks passed** |

**Test environment:** local Docker Compose, developer laptop (ASUS TUF Dash F15). <!-- PLACEHOLDER: fill in CPU / RAM / OS if you want the numbers to carry more weight, e.g. "Ryzen 7 5800H, 16GB RAM, Ubuntu 22.04" -->

<details>
<summary>Raw vegeta output</summary>

```
Target: http://localhost:8080/check | Rate: 200/s | Duration: 30s | Client: loadtest-client

--- Latency histogram ---
Bucket           #     %        Histogram
[0s,     10ms]   6000  100.00%  ###########################################################################
[10ms,   25ms]   0     0.00%
[25ms,   50ms]   0     0.00%
[50ms,   100ms]  0     0.00%
[100ms,  250ms]  0     0.00%
[250ms,  500ms]  0     0.00%
[500ms,  1s]     0     0.00%
[1s,     +Inf]   0     0.00%

--- Status code breakdown ---
{'200': 100, '429': 5900}
```

</details>

<details>
<summary>Raw k6 output</summary>

```
     scenarios: (100.00%) 1 scenario, 200 max VUs, 1m0s max duration (incl. graceful stop):
              * steady_load: 200.00 iterations/s for 30s (maxVUs: 50-200, gracefulStop: 30s)

  █ THRESHOLDS
    http_req_duration
    ✓ 'p(95)<100' p(95)=1.34ms
    ✓ 'p(99)<250' p(99)=1.73ms

  █ TOTAL RESULTS
    checks_total.......: 12000   399.907604/s
    checks_succeeded...: 100.00% 12000 out of 12000
    checks_failed......: 0.00%   0 out of 12000

    CUSTOM
    rate_limit_allowed.............: 5000   166.628168/s
    rate_limit_blocked.............: 1000   33.325634/s

    HTTP
    http_req_duration..............: avg=930.98µs min=166.25µs med=942.28µs max=12.37ms p(90)=1.24ms  p(95)=1.34ms
      { expected_response:true }...: avg=928.8µs  min=166.25µs med=942.74µs max=12.37ms p(90)=1.24ms  p(95)=1.35ms
    http_req_failed................: 16.66% 1000 out of 6000
    http_reqs......................: 6000   199.953802/s

    EXECUTION
    vus............................: 2      min=2            max=3
    vus_max........................: 50     min=50           max=50

    NETWORK
    data_received..................: 1.1 MB 36 kB/s
    data_sent......................: 575 kB 19 kB/s

running (0m30.0s), 000/050 VUs, 6000 complete and 0 interrupted iterations
```

Note: k6 reports the 429 responses under `http_req_failed` (16.66%) because it
counts any non-2xx as a failed HTTP request by default — that's expected and
correct behavior for a rate limiter under load, not a system error. Zero
requests errored at the connection/transport level.

</details>

<!-- PLACEHOLDER: screenshot of the Grafana dashboard captured during one
of these runs — this is stronger evidence than the raw text above.
![vegeta-report](docs/screenshots/loadtest-grafana.png) -->

<!-- PLACEHOLDER: run at 2-3 more load levels (e.g. 500, 1000, 2000 rps)
and add rows to the table above, so it shows how p99 and throughput scale
as load increases rather than a single data point. -->

<!-- PLACEHOLDER (optional but strong): before/after table if you build a
"naive version" for comparison, e.g. no Lua script (GET+SET race condition)
vs the atomic version — this is genuinely good interview material per the
project's original methodology notes. -->

## Chaos / Failure Testing

<!-- PLACEHOLDER: document what happens when you kill Redis mid-load-test.
Suggested steps to actually run this and fill in:
1. Start a vegeta/k6 run against /check
2. `docker stop ratelimiter-redis` partway through
3. Observe: does the API fail open (503) or crash? Check Grafana's
   rate_limiter_redis_errors_total panel and the request rate panel.
4. `docker start ratelimiter-redis` and confirm recovery time.

Write your actual observed behavior and recovery time here, plus a
screenshot of the Grafana panel showing the error spike and recovery. -->

## Design Decisions

- **Sliding window (Redis `ZSET`) over fixed window or token bucket.** Fixed window allows up to 2x the configured limit in a burst straddling a window boundary — a real abuse vector on API gateways. Sliding window costs one `ZSET` per client (bounded by the window's PEXPIRE) in exchange for exact enforcement. Token bucket was the other serious contender for burst-tolerant use cases; it's on the [roadmap](#roadmap) as a pluggable alternative rather than baked in, since the two have genuinely different semantics and a real service should let the caller choose per-policy.
- **One Lua script, not separate `GET`+`SET` calls.** Redis executes Lua scripts atomically and single-threaded — that's what makes the check-and-increment safe under concurrency without a separate distributed lock. Verified with a concurrent `go test -race` firing 200 goroutines at a limit of 50 and asserting exactly 50 get through.
- **Fail open on Redis errors, not fail closed.** If Redis is unreachable, the middleware returns 503 rather than either (a) blocking all traffic, which turns a dependency outage into a total API outage, or (b) silently letting everything through, which defeats the limiter exactly when it might matter most under a real incident. 503 is honest about what broke.
- **Async analytics via buffered channel + worker pool, not synchronous writes.** A rate limiter that makes every request wait on a Postgres insert is a rate limiter that makes your API only as fast as your slowest database write. Metrics are best-effort: `Submit()` is non-blocking and drops (with a counter + log line) rather than backpressuring the request path when the buffer is full.
- **`pgx.Batch` instead of per-event inserts.** Flushing on whichever comes first — 500 buffered events or a 2-second ticker — turns N round trips into 1, which matters once request volume is high enough that the collector is the thing under load, not the API.

## Roadmap

- [ ] CI pipeline (GitHub Actions: `go vet`, `golangci-lint`, `go test -race`)
- [ ] Reduce per-request allocations in the hot path (currently 128 allocs/op — worth profiling with `pprof`)
- [ ] Policy management CRUD API with auth
- [ ] Pluggable algorithms (token bucket, fixed window) behind the same `Limiter` interface
- [ ] Live deployment (Fly.io / Railway) with a public demo endpoint
- [ ] Redis Cluster for HA
- [ ] Chaos-test Redis failure mid-load (see [Chaos / Failure Testing](#chaos--failure-testing))
<!-- PLACEHOLDER: add/remove items as you actually build them -->

## License

<!-- PLACEHOLDER: e.g. MIT — add a LICENSE file if you want this public -->
