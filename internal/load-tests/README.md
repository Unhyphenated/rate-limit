# Load Testing & Performance Analysis

## Overview

This project was evaluated under multiple load scenarios using `k6` to assess **latency, throughput, and rate-limiting behavior** under realistic and extreme conditions.

The system maintains **low median latency (<2ms)** under most conditions, while correctly enforcing rate limits under overload. Tail latency increases under contention, indicating **Redis/Lua execution as the primary bottleneck**.

---

## Test Scenarios & Results

### Summary

| Scenario        | Target RPS | Achieved RPS | p50    | p95    | Max   | Rate Limited |
| --------------- | ---------- | ------------ | ------ | ------ | ----- | ------------ |
| Steady Load     | 50         | 50           | 1.83ms | 7.56ms | 1.59s | 0%           |
| Ramp-Up         | 0 → 200    | 105          | 1.3ms  | 5.1ms  | 110ms | 17.6%        |
| Burst Load      | 500        | 498          | 0.76ms | 13.5ms | 265ms | 78%          |
| DDoS Simulation | 1000       | 1000         | 0.4ms  | 20.6ms | 1.43s | 89%          |

---

## Scenario Breakdown

### 1. Steady Load (50 req/s)

**Objective:** Establish baseline performance under normal conditions.

* Stable throughput at 50 req/s
* Low latency:

  * p50: 1.83ms
  * p95: 7.56ms
* 0% rate limiting

**Analysis:**
System operates with minimal contention. Redis is not a bottleneck at this load level.

---

### 2. Ramp-Up (0 → 200 req/s)

**Objective:** Evaluate scaling behavior as load increases.

* Throughput scales linearly (~105 req/s avg)
* Latency remains low:

  * p50: 1.3ms
  * p95: 5.1ms
* Rate limiting begins (~17.6%)

**Analysis:**
The system scales smoothly with increasing load. Rate limiting activates progressively without significant latency impact.

---

### 3. Burst Load (500 req/s)

**Objective:** Test sustained high-throughput conditions.

* Throughput maintained at ~498 req/s
* Latency:

  * p50: 0.76ms
  * p95: 13.5ms
* High rate limiting (~78%)

**Analysis:**
Median latency remains low, but tail latency increases significantly. This indicates **contention at the Redis/Lua execution layer** under sustained load.

---

### 4. DDoS Simulation (1000 req/s, single hot key)

**Objective:** Simulate worst-case contention (single key saturation).

* Throughput sustained at ~1000 req/s
* Latency:

  * p50: 0.4ms
  * p95: 20.6ms
* Very high rate limiting (~89%)

**Analysis:**
Heavy contention on a single key leads to increased tail latency. Despite this, the system maintains correctness and effectively protects downstream services by aggressively rejecting excess traffic.

---

## Key Insights

### 1. Tail Latency Under Contention

p95 latency increases from ~7ms (steady) to ~20ms (DDoS), while p50 remains sub-millisecond.

**Interpretation:**
Latency degradation is caused by **serialization and contention in Redis Lua scripts**, not application-level inefficiencies.

---

### 2. Strong Median Performance

Across all scenarios, p50 remains <2ms.

**Interpretation:**
The system is efficient for the majority of requests and not CPU-bound at the application layer.

---

### 3. Correct Rate Limiting Behavior

* High percentage of 429 responses under load is expected
* 100% of checks passed (no correctness failures)

**Interpretation:**
The limiter reliably enforces constraints even under extreme load.

---

### 4. Redis as Primary Bottleneck

Performance degradation correlates with increased contention.

**Interpretation:**
Redis (single instance) becomes the limiting factor at higher throughput levels.

---

## Limitations

* **Single Redis instance** (no clustering or replication)
* **Single-key contention** in worst-case scenarios
* Tests executed in **local Docker environment** (not production network conditions)

---

## Conclusion

The system is **production-viable for moderate workloads**, delivering low latency and correct rate limiting behavior.

For high-concurrency environments, improvements such as **Redis clustering, sharding, or multi-region strategies** would be required to reduce tail latency and eliminate single points of failure.
