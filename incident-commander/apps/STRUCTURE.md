# AI-Ready E-Commerce — Application Structure & End-to-End Flow

Version: **2.0** · Observability: **Datadog Native Stack** · Language: **Golang 1.24**

---

## 1. Directory Structure

```
apps/
├── docker-compose.platform.yml          ← PostgreSQL + Redis + Kafka + Elasticsearch
├── docker-compose.services.yml          ← All microservices (build + run)
│
├── ms-order/                            # Order Service (HTTP :8080 + Kafka)
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   └── internal/
│       ├── config/config.go             # Env loader
│       ├── model/model.go               # Structs: Order, Events, CachedOrder
│       ├── handler/
│       │   ├── order.go                 # POST /api/orders → DB + Redis + Kafka
│       │   └── payment.go              # Consumes payment-completed → update status
│       └── infra/
│           ├── postgres.go              # PostgreSQL connection + migration
│           ├── redis.go                 # Redis client
│           └── kafka.go                 # Kafka producer (order-created) + consumer (payment-completed)
│
├── ms-inventory/                        # Inventory Service (Kafka consumer)
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   └── internal/
│       ├── config/config.go
│       ├── model/model.go               # Structs: Inventory, Events, CachedInventory
│       ├── handler/
│       │   └── inventory.go            # Consume order-created → validate → reserve → publish
│       └── infra/
│           ├── postgres.go              # PostgreSQL + seed inventory data (SKU-001..005)
│           ├── redis.go
│           └── kafka.go                 # Kafka consumer (order-created) + producer (inventory-reserved)
│
├── ms-processing-payment/               # Payment Service + Gateway Mock
│   ├── .env.example
│   ├── architecture.md                  # Original architecture spec
│   ├── docker-compose.platform.yml      # Local copy (run independently)
│   ├── docker-compose.services.yml      # Local copy (run independently)
│   ├── payment-gateway-mock/            # Mock external payment gateway
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go                      # POST /api/payment → {SUCCESS|FAILED}
│   └── payment-service/                 # Main payment processor
│       ├── Dockerfile
│       ├── go.mod
│       ├── main.go
│       └── internal/
│           ├── config/config.go
│           ├── model/model.go           # Structs: Payment, Events, ESPaymentDocument
│           ├── handler/
│           │   └── handler.go           # Consume → Gateway → DB → Elasticsearch → Kafka
│           └── infra/
│               ├── postgres.go          # PostgreSQL + migration
│               ├── elasticsearch.go     # ES client (auto-traced via dd-trace)
│               ├── kafka.go             # Kafka consumer (inventory-reserved) + producer (payment-completed)
│
└── incident-commander-apps/             # AI Incident Commander (Python/FastAPI — unchanged)
    └── ...

├── ms-load-generator/                   # Traffic simulator (HTTP client)
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go                          # Concurrent POST /api/orders loop
```

---

## 2. Infrastructure Platform

```yaml
# docker-compose.platform.yml

Services:
  postgres        postgres:16-alpine    :5432   volumes: postgres_data
  redis           redis:7-alpine        :6379
  kafka           bitnami/kafka         :9092   KRaft mode (no Zookeeper)
  elasticsearch   elasticsearch:8.19.4  :9200   security disabled, single-node

Network:          platform-network (bridge)

Run:              docker-compose -f docker-compose.platform.yml up -d
```

---

## 3. Services Orchestration

```yaml
# docker-compose.services.yml

Services (startup order):

  1. payment-gateway-mock    :8081   Mock HTTP server
  2. order-service           :8080   HTTP API + Kafka consumer
  3. inventory-service       —       Kafka consumer (depends on order-service)
  4. payment-service         —       Kafka consumer (depends on gateway + inventory)
  5. load-generator          —       Traffic simulator (depends on order-service)

Network:                     platform-network (external)

Run:                         docker-compose -f docker-compose.services.yml up -d --build
```

---

## 4. End-to-End Transaction Flow

```
Load Generator (5 workers, 10 rps)
  │
  │  POST /api/orders   {"customer_id":"cust-421","product_id":"SKU-003","quantity":3}  ← random
  │
  ▼
┌──────────────────────────────────────────────────────────────────┐
│                      ORDER SERVICE  (:8080)                      │
│                                                                  │
│  [span: order.create]                                            │
│    │                                                             │
│    ├──  PostgreSQL INSERT  →  orders table                       │
│    │    [span: postgres.query]                                   │
│    │                                                             │
│    ├──  Redis SET  →  order:{order_id}  (TTL 5m)                 │
│    │    [span: redis.cache]                                      │
│    │                                                             │
│    └──  Kafka PUBLISH  →  topic: order-created                   │
│         {order_id, product_id, quantity}                         │
│         [span: kafka.produce]                                    │
│                                                                  │
│  Response:  {"order_id":"ORD-...","status":"PENDING"}            │
└──────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼  order-created
┌──────────────────────────────────────────────────────────────────┐
│                    INVENTORY SERVICE                             │
│                                                                  │
│  [span: inventory.reserve]                                       │
│    │                                                             │
│    ├──  Redis GET  →  inventory:{product_id}                     │
│    │    [span: redis.cache]    ← hit / miss                      │
│    │                                                             │
│    ├──  PostgreSQL SELECT  →  inventory table                    │
│    │    [span: postgres.query]                                   │
│    │                                                             │
│    ├──  PostgreSQL UPDATE  →  stock = stock - qty                │
│    │    [span: postgres.update]   (only if stock ≥ qty)          │
│    │                                                             │
│    ├──  Redis SET  →  inventory:{product_id}  (TTL 1m)           │
│    │    [span: redis.cache]                                      │
│    │                                                             │
│    └──  Kafka PUBLISH  →  topic: inventory-reserved              │
│         {order_id, product_id, quantity, amount, status}         │
│         [span: kafka.produce]                                    │
└──────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼  inventory-reserved
┌──────────────────────────────────────────────────────────────────┐
│                     PAYMENT SERVICE                              │
│                                                                  │
│  [span: payment.process]                                         │
│    │                                                             │
│    ├──  HTTP POST  →  Payment Gateway Mock  /api/payment         │
│    │    [span: http.client]                                      │
│    │    ←  {transaction_id, status: SUCCESS|FAILED}              │
│    │                                                             │
│    ├──  PostgreSQL INSERT  →  payments table                     │
│    │    [span: postgres.query]                                   │
│    │                                                             │
│    ├──  Elasticsearch INDEX  →  index: payment-transactions  ★   │
│    │    [span: elasticsearch.index]                               │
│    │    Document: {payment_id, order_id, amount, status, ...}    │
│    │    (ES HTTP calls auto-traced via dd-trace RoundTripper)    │
│    │                                                             │
│    └──  Kafka PUBLISH  →  topic: payment-completed               │
│         {order_id, payment_id, amount, status, timestamp}        │
│         [span: kafka.produce]                                    │
└──────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼  payment-completed
┌──────────────────────────────────────────────────────────────────┐
│                      ORDER SERVICE  (callback)                   │
│                                                                  │
│  [span: order.payment_completed]                                 │
│    │                                                             │
│    ├──  PostgreSQL UPDATE  →  orders.status = COMPLETED|FAILED   │
│    │    [span: postgres.query]                                   │
│    │                                                             │
│    └──  Redis UPDATE  →  order:{order_id}                        │
│         [span: redis.cache]                                      │
└──────────────────────────────────────────────────────────────────┘
```

---

## 5. Datadog APM — Expected Trace Flame Graph

```
[Load Generator]                        ← upstream caller
 └── http.client             (POST /api/orders → order-service)

[Order Service]
 ├── order.create
 │   ├── postgres.query          (INSERT orders)
 │   ├── redis.cache             (SET order:{id})
 │   └── kafka.produce           (order-created)
 │
 ├── order.payment_completed     (via Kafka consumer callback)
 │   ├── postgres.query          (UPDATE orders)
 │   └── redis.cache             (UPDATE order:{id})
 │
[Inventory Service]
 └── inventory.reserve
     ├── redis.cache             (GET inventory:{sku})
     ├── postgres.query          (SELECT inventory)
     ├── postgres.update         (UPDATE stock)
     ├── redis.cache             (SET inventory:{sku})
     └── kafka.produce           (inventory-reserved)

[Payment Service]
 └── payment.process
     ├── http.client             (POST /api/payment → gateway)
     ├── postgres.query          (INSERT payments)
     ├── ★ elasticsearch.index   (INDEX payment-transactions)
     └── kafka.produce           (payment-completed)

[Payment Gateway Mock]
 └── http.server                 (POST /api/payment)
```

---

## 6. Kafka Topics

| Topic              | Producer          | Consumer           |
|--------------------|-------------------|--------------------|
| `order-created`    | Order Service     | Inventory Service  |
| `inventory-reserved`| Inventory Service| Payment Service    |
| `payment-completed`| Payment Service   | Order Service      |

---

## 7. Database Tables

| Table      | Service           | Key Columns                              |
|------------|-------------------|------------------------------------------|
| `orders`    | Order Service     | id, customer_id, product_id, qty, status |
| `inventory` | Inventory Service | product_id, stock, updated_at            |
| `payments`  | Payment Service   | id, order_id, amount, status, gateway_txn_id |

---

## 8. Redis Keys

| Key                       | Service           | TTL      |
|---------------------------|-------------------|----------|
| `order:{order_id}`        | Order Service     | 5 minutes|
| `inventory:{product_id}`  | Inventory Service | 1 minute |

---

## 9. Elasticsearch Index

| Index                  | Service         | Purpose                         |
|------------------------|-----------------|---------------------------------|
| `payment-transactions` | Payment Service | Full-text search + analytics on all payment documents |

---

## 10. Business Metrics (DogStatsD)

| Metric               | Type    | Tags                          |
|----------------------|---------|-------------------------------|
| `orders.created`     | COUNT   | service:order-service         |
| `orders.completed`   | COUNT   | service:order-service         |
| `orders.failed`      | COUNT   | service:order-service         |
| `inventory.reserved` | COUNT   | service:inventory-service     |
| `inventory.failed`   | COUNT   | service:inventory-service     |
| `payment.success`    | COUNT   | service:payment-service       |
| `payment.failed`     | COUNT   | service:payment-service       |
| `payment.latency`    | GAUGE   | service:payment-service       |
| `revenue.total`      | COUNT   | —                             |

---

## 11. Quick Start

```bash
# 1. Start infrastructure
docker-compose -f docker-compose.platform.yml up -d

# 2. Start Datadog Agent (replace API key)
docker run -d --name dd-agent \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /proc/:/host/proc/:ro -v /sys/fs/cgroup/:/host/sys/fs/cgroup:ro \
  -e DD_API_KEY=<YOUR_KEY> -e DD_SITE=us5.datadoghq.com \
  -p 8126:8126 -p 8125:8125/udp \
  gcr.io/datadoghq/agent:latest

# 3. Start all services (including automatic load generator)
docker-compose -f docker-compose.services.yml up -d --build

# 4. (Optional) Manual single order
curl -X POST http://localhost:8080/api/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":"123","product_id":"SKU-001","quantity":2}'

# 5. Open Datadog APM → Service Map → click any trace
```

---

## 12. Service Tags (all services)

```
env:hackathon
team:platform
region:ap-southeast
version:1.0.0
```

---

## 13. Key Integration: Elasticsearch in Distributed Tracing

The Payment Service indexes every processed payment to Elasticsearch. This call is **automatically traced** in Datadog APM via `dd-trace-go/contrib/net/http` wrapping the Elasticsearch HTTP transport:

```go
// internal/infra/elasticsearch.go
esCfg := elasticsearch.Config{
    Addresses: []string{esURL},
    Transport: httptrace.WrapRoundTripper(http.DefaultTransport), // ← auto-span
}
es, _ := elasticsearch.NewClient(esCfg)
```

This produces a **native Datadog APM span** for every Elasticsearch operation (Index, Search, etc.), visible alongside PostgreSQL, Redis, and Kafka spans in the distributed trace flame graph — without any manual instrumentation beyond the client setup.

---

## 14. Load Generator (`ms-load-generator`)

### 14.1 Overview

Traffic simulator built-in. Sends concurrent `POST /api/orders` requests with randomized payload to emulate real e-commerce traffic. Every HTTP call is automatically traced by Datadog APM and appears as a span in the distributed trace.

| Property | Value |
|----------|-------|
| Target | `http://order-service:8080/api/orders` |
| Random `customer_id` | `cust-1` … `cust-999` |
| Random `product_id` | `SKU-001` … `SKU-005` |
| Random `quantity` | `1` … `5` |
| DD tracing | `httptrace.WrapClient()` → auto `http.client` span |
| DD Service name | `load-generator` |

---

### 14.2 Environment Variables

| Env | Default | Description |
|-----|---------|-------------|
| `ORDER_SERVICE_URL` | `http://order-service:8080` | Target endpoint |
| `WORKERS` | `5` | Number of concurrent goroutines |
| `RPS` | `10` | Total requests per second (auto-split across workers) |
| `DURATION_SECONDS` | `0` | Seconds to run (`0` = infinite) |
| `DD_AGENT_HOST` | `host.docker.internal` | Datadog Agent host |
| `DD_ENV` | `hackathon` | Datadog environment tag |
| `DD_SERVICE` | `load-generator` | Datadog service name |
| `DD_VERSION` | `1.0.0` | Datadog version tag |

---

### 14.3 Traffic Scenarios

```
┌─────────────────────────────────────────────────────────────────┐
│  Scenario        WORKERS   RPS   DURATION    Result             │
├─────────────────────────────────────────────────────────────────┤
│  Idle            1         1      0          ~1 req/s, demo     │
│  Default         5         10     0          ~10 req/s, steady  │
│  Moderate        10        50     0          ~50 req/s          │
│  High            20        100    0          ~100 req/s, stress │
│  Burst (1 min)   10        200    60         200 req/s × 60s    │
│  Burst (5 min)   20        500    300        500 req/s × 5m     │
└─────────────────────────────────────────────────────────────────┘
```

**Pola payload yang dikirim:**

```json
// Worker 0, tick 1
{"customer_id":"cust-421","product_id":"SKU-003","quantity":3}

// Worker 1, tick 1
{"customer_id":"cust-887","product_id":"SKU-001","quantity":5}

// Worker 0, tick 2
{"customer_id":"cust-102","product_id":"SKU-005","quantity":1}
```

---

### 14.4 Cara Menjalankan

#### Default (via docker-compose)

Load generator otomatis start bersama semua service:

```bash
docker-compose -f docker-compose.platform.yml up -d
docker-compose -f docker-compose.services.yml up -d --build
# Load generator berjalan dengan WORKERS=5, RPS=10 (default)
```

#### Custom traffic via env override

```bash
# High traffic: 20 workers, 100 req/s
WORKERS=20 RPS=100 docker-compose -f docker-compose.services.yml up -d --build

# Burst 60 detik, lalu berhenti sendiri
WORKERS=10 RPS=200 DURATION_SECONDS=60 docker-compose -f docker-compose.services.yml up -d --build
```

#### Standalone (tanpa compose)

```bash
cd apps/ms-load-generator
docker build -t load-generator .

# Run against order-service on host
docker run --rm --network platform-network \
  -e ORDER_SERVICE_URL=http://order-service:8080 \
  -e WORKERS=5 -e RPS=10 \
  -e DD_AGENT_HOST=host.docker.internal \
  load-generator
```

#### Scale via docker-compose (multiple replicas)

```bash
# Tambah 3 container load-generator, total = 3 × 10 = 30 req/s
docker-compose -f docker-compose.services.yml up -d --scale load-generator=3
```

---

### 14.5 Verifikasi

```bash
# 1. Cek logs — stats muncul setiap 10 detik
docker logs -f svc-load-generator

# Output:
# [loadgen] starting: workers=5 rps=10 target=http://order-service:8080/api/orders
# [loadgen] stats: success=142 errors=3
# [loadgen] stats: success=291 errors=5
# [loadgen] stats: success=438 errors=8

# 2. Datadog APM → Service Map
#    Node baru: [load-generator] ──► [order-service]

# 3. Datadog APM → Traces
#    Span di load-generator: http.client → POST /api/orders
#    Span di order-service:    order.create → postgres.query, redis.cache, kafka.produce

# 4. Stop load generator saja (tanpa stop service lain)
docker stop svc-load-generator
```

---

### 14.6 Internal Flow

```
main()
  │
  ├── dd-trace init          (tracer.Start)
  ├── httptrace.WrapClient   (auto-instrument HTTP)
  │
  ├── spawn N WORKERS        (goroutine loop)
  │     │
  │     ├── ticker (RPS rate)
  │     │
  │     ├── sendOrder()
  │     │     ├── random customer_id, product_id, quantity
  │     │     ├── json.Marshal
  │     │     └── client.Post(url, body)    ← auto [http.client] span
  │     │
  │     └── loop until stopCh
  │
  ├── stats reporter         (every 10s: success/error counts)
  │
  └── graceful shutdown      (SIGINT, SIGTERM, or DURATION reached)
        └── tracer.Stop()
```
