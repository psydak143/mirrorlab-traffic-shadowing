# MirrorLab – Traffic Shadowing with SLO Guardrails

MirrorLab is a small end-to-end system that shows how to **test a new version of a service using real traffic without impacting users**.

A Go proxy (API gateway) forwards all requests to a **baseline** Java service and mirrors a configurable fraction to a **candidate** Java service. It compares responses and latency, records Prometheus metrics, and uses a simple **SLO-style guardrail** to automatically turn off mirroring if the candidate misbehaves. Prometheus and Grafana provide dashboards for latency, error rate, response diffs, and guardrail state.

> High-level idea: “send a copy of production traffic to a new version, watch it carefully, and cut it off automatically if it’s slower or more error-prone.”

---

## Features

- **Traffic shadowing proxy (Go)**
  - Forwards requests to a **baseline** service and returns that response to the client.
  - Mirrors a fraction of requests to a **candidate** service in the background (candidate responses are never returned to users).
- **Response comparison**
  - Parses JSON responses from baseline and candidate.
  - Ignores volatile fields like `timestamp`, `trace_id`, `request_id`.
  - Masks simple PII-like fields (`email`, `phone`, `name`, `cc_last4`).
  - Counts mismatches per route via `mirror_diff_mismatches_total`.
- **SLO-style guardrail**
  - Compares baseline vs candidate latency and error behavior per request.
  - Maintains a breach streak based on:
    - Latency ratio (candidate / baseline) vs `ABORT_P99_RATIO`.
    - Error behavior (candidate errors when baseline does not).
  - When breaches happen repeatedly, mirroring is **auto-disabled** and a reason is recorded.
  - Guardrail is exposed via `/control/status` and can be toggled via `/control/enable` and `/control/disable`.
- **Java demo services (v1 and v2)**
  - Single Spring Boot app (Java 21) run as:
    - `service-v1` = **baseline** (no chaos).
    - `service-v2` = **candidate** (supports injected latency and error rate).
  - REST API:
    - `GET /api/search?q=...`
    - `GET /api/product/{id}`
    - `POST /api/checkout`
  - Deterministic responses and order IDs (same input → same output) to make diffing meaningful.
- **Observability**
  - Go proxy exposes Prometheus metrics (`/metrics`).
  - Java services expose Micrometer Prometheus metrics (`/actuator/prometheus`).
  - Prometheus scrapes proxy + services.
  - Grafana dashboard for:
    - RPS (baseline vs candidate)
    - p99 latency (baseline vs candidate)
    - Error rate (candidate vs baseline)
    - Response diff mismatches
    - Mirroring state & abort count
- **Docker-first setup**
  - All components (proxy, services, Prometheus, Grafana) run via `docker compose`.
  - Developed with Go `1.25.x` and JDK `21`, but only Docker is required to run.

---

## Architecture

```text
  Client
    │
    ▼
+--------------------------+
| Go Proxy (MirrorLab)    |
| - forwards to baseline  |
| - mirrors to candidate  |
| - JSON diff + guardrail |
+-----------+--------------+
            │
   baseline │   candidate (mirrored)
            │
            ▼
  service-v1 (Java)       service-v2 (Java)
  /api/...                /api/... (chaos flags)

  Proxy & services → Prometheus → Grafana
