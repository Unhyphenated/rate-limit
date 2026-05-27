# Distributed API Gateway Rate Limiter

![Go](https://img.shields.io/badge/Go-1.25.5-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)
[![Rate Limiter Tests](https://github.com/Unhyphenated/rate-limit/actions/workflows/ci.yml/badge.svg)](https://github.com/Unhyphenated/rate-limit/actions/workflows/ci.yml)

> A high-performance, horizontally scalable API gateway rate-limiting tier built in Go — designed to protect downstream microservices from traffic surges, bursts, and DDoS attempts using an atomic Token Bucket engine backed by Redis Lua scripts.

By decoupling the application logic from the state store, the gateway scales horizontally across multiple stateless Go nodes balanced by Nginx. It features a mathematically continuous Token Bucket algorithm executed atomically inside **Redis via Lua scripts**, eliminating application-level locking and network-heavy race conditions.

---

## Core Features

* **Atomic Token Bucket Engine:** Computes token replenishment mathematically on-the-fly inside Redis using Lua scripts. Avoids background cron processes and keeps memory complexity at `O(1)`.
* **Resilient Fail-Open Circuit Breaker:** Implements a custom state machine that wraps the cache client. During a total Redis outage, the gateway drops into a fast fail-open state to prioritize 100% system availability over absolute traffic policing.
* **Stateless Multi-Node Topology:** Horizontally scales across multiple Go nodes balanced via Nginx Round Robin, bypassing node clock skew by using the Redis server as the central source of truth for time.
* **Production Observability:** Fully instrumented using the RED Method (Rate, Errors, Duration) via Prometheus and customized sub-millisecond histogram buckets, visualized through a pre-configured Grafana dashboard.

---

## Architecture

```
                       [ Incoming Traffic ]
                                │
                                ▼
                      ┌────────────────────┐
                      │    Nginx Proxy     │
                      │   (Round Robin)    │
                      └─────────┬──────────┘
                                │
          ┌─────────────────────┼─────────────────────┐
          ▼                     ▼                     ▼
┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐
│  Go API Node 1    │ │  Go API Node 2    │ │  Go API Node 3    │
│┌─────────────────┐│ │┌─────────────────┐│ │┌─────────────────┐│
││ Circuit Breaker ││ ││ Circuit Breaker ││ ││ Circuit Breaker ││
│└────────┬────────┘│ │└────────┬────────┘│ │└────────┬────────┘│
└─────────┼─────────┘ └─────────┼─────────┘ └─────────┼─────────┘
          │                     │                     │
          └─────────────────────┼─────────────────────┘
                                ▼
                      ┌────────────────────┐
                      │ Centralized Redis  │ <── [Prometheus Exporter]
                      │ (Lua Script State) │
                      └────────────────────┘
```

---

## Project Structure

The project follows standard Go enterprise layout patterns, separating operational configurations from core domain logic:

```
rate-limit/
├── docker-compose.dev.yaml
├── docker-compose.yaml
├── Dockerfile
├── Dockerfile.dev
├── go.mod
├── go.sum
├── Makefile
├── nginx.conf
├── prometheus.yml
├── .air.toml
├── cmd/
│   └── main.go
├── grafana/
│   └── provisioning/
│       └── datasources/
│           └── datasources.yaml
├── internal/
│   ├── breaker/
│   │   ├── breaker.go
│   │   └── breaker_test.go
│   ├── cache/
│   │   ├── circuit_breaker_cache.go
│   │   ├── circuit_breaker_cache_test.go
│   │   └── redis.go
│   ├── config/
│   │   ├── limits.go
│   │   └── limits_test.go
│   ├── handlers/
│   │   ├── orders.go
│   │   ├── prices.go
│   │   ├── trades.go
│   │   └── wallet.go
│   ├── limiter/
│   │   ├── limiter.go
│   │   └── limiter_integration_test.go
│   ├── load-tests/
│   │   ├── README.md
│   │   ├── burst.js
│   │   ├── common.js
│   │   ├── ddos.js
│   │   └── steady-load.js
│   ├── metrics/
│   │   └── metrics.go
│   ├── middleware/
│   │   └── handlers.go
│   └── models/
│       ├── bucket.go
│       └── payload.go
└── .github/
    └── workflows/
        └── ci.yml
```

---

## Load Test Performance Overview

All metrics were validated using an open-loop **k6** load testing engine executing against a local multi-container Docker Compose topology.

### Performance Matrix

| Scenario | Target Load | Achieved Throughput | Median (p50) | Tail (p99) | Error Rate / Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Steady State** | 500 RPS | 500.01 RPS | 933µs | 11.26ms | 0% Errors / Healthy |
| **Sustained Burst** | 3000 RPS | 3,000.00 RPS | 305µs | 9.99ms | 0% Errors / Immediate Recovery |
| **DDoS Isolation** | 1050 RPS | 1,044.00 RPS | 565µs (Legit) | 82.85ms (Legit) | **98.15% Attacker Requests Dropped** |
| **Infrastructure Chaos** | 500 RPS | 497.68 RPS | 2.63ms | 138.25ms | 0% Errors / **Fail-Open Active** |

### Key Operational Observations
* **Immediate Burst Recovery:** During a sudden 60x traffic spike (from 50 RPS to 3,000 RPS), tail latency remained bounded under 10ms. The system returned to its baseline footprint within milliseconds of the spike terminating, demonstrating no lingering queue buildup.
* **Identity-Level Blast Isolation:** Under a simulated 1,000 RPS single-IP DDoS assault, the system isolated and rejected 98.15% of the malicious requests. Legitimate concurrent consumers experienced a 0% drop rate, preserving application usability.
* **Outage Containment Bounds:** When the Redis instance was abruptly killed mid-flight, the custom circuit breaker tripped within 5 failed iterations (~10ms window), safely bypassing the broken cache layer to serve traffic without dropping requests.

---

## Key Design Decisions

For an exhaustive analysis of the trade-offs made during development, refer to [DECISIONS.md](DECISIONS.md).

* **Concurrency Control:** Chosen Redis Lua scripts over `WATCH`/`MULTI` or distributed locks. Lua scripts execute atomically on the single-threaded Redis event loop, shifting the business logic to the data layer and avoiding the retry loops or heavy network overhead of alternative approaches.
* **Rate Limiting Algorithm:** Chosen Token Bucket over Sliding Window Log. Token Bucket maintains an `O(1)` memory footprint by tracking only two fields per client (tokens and last timestamp) using a time-dilated mathematical formula, avoiding the memory spikes of tracking individual timestamps under DDoS load.
* **Resilience Strategy:** Chosen a Fail-Open approach using a custom Circuit Breaker state machine. For a public data gateway, maximizing service availability is prioritized over strict traffic policing.
* **State Storage Layout:** Chosen Redis Hashes over flat independent String keys. Grouping client token balances and timestamps into single localized Hash structures reduces keyspace clutter and maximizes memory efficiency via Redis listpack encoding.

---

## Tech Stack

* **Language:** Go (Golang) using native `net/http` primitives.
* **State Store:** Redis (with transactional Lua scripting payloads).
* **Reverse Proxy:** Nginx.
* **Containerization:** Docker / Docker Compose.
* **Load Simulation:** k6.
* **Observability Architecture:** Prometheus Server architecture + Grafana + `redis_exporter`.

---

## Quickstart & Local Deployment

### Prerequisites
Ensure you have Docker and Docker Compose installed locally.

### 1. Spin Up Infrastructure
Launch the 3 Go API gateways, Nginx reverse proxy, Redis instance, Prometheus metrics collector, and Grafana dashboard in detached mode:

```bash
git clone https://github.com/Unhyphenated/rate-limit.git
cd rate-limit
docker compose up -d --build
```

### 2. Verify Ingress Routing

Send an ad-hoc test request directly to the Nginx load balancer:

```bash
curl -i http://localhost:8080/api/v1/prices
```

### 3. Access Live Metrics

* Open your browser and navigate to Grafana at `http://localhost:3000` (Default credentials: `admin` / `admin`).
* Access the pre-loaded **Rate Limiter Operational Dashboard** to monitor request rates, active Goroutine allocations, latency quantiles (p50, p95, p99), and current circuit breaker states in real time.

---

## License

This project is licensed under the [MIT License](LICENSE).