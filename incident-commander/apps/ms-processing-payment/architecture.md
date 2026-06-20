# Datadog Hackathon Demo Platform

## E-Commerce Order Processing System

## 1. Overview

Dokumen ini menjelaskan arsitektur dan implementasi sistem e-commerce sederhana yang dirancang khusus untuk mendemonstrasikan kemampuan observability menggunakan Datadog.

Platform ini terdiri dari tiga microservices utama yang saling terintegrasi menggunakan PostgreSQL, Redis, Kafka, dan External API sehingga menghasilkan distributed tracing yang lengkap dan memungkinkan simulasi berbagai jenis incident.

## 2. Objectives

### Functional Objectives

* Create Order
* Reserve Inventory
* Process Payment
* View Order Status
* Cancel Order

### Observability Objectives

* Distributed Tracing
* Service Map
* Error Tracking
* Deployment Tracking
* Database Monitoring
* Redis Monitoring
* Kafka Monitoring
* Log Correlation
* Custom Business Metrics
* Synthetic Incident Simulation

---

# 3. High Level Architecture

```text
                           ┌───────────────┐
                           │     Client    │
                           └───────┬───────┘
                                   │
                                   ▼

                      ┌────────────────────────┐
                      │     Order Service      │
                      └──────────┬─────────────┘
                                 │
                  ┌──────────────┼──────────────┐
                  │              │              │
                  ▼              ▼              ▼

            PostgreSQL        Redis         Kafka
                                                 │
                                                 ▼

                      ┌────────────────────────┐
                      │   Inventory Service    │
                      └──────────┬─────────────┘
                                 │
                  ┌──────────────┼──────────────┐
                  │              │              │
                  ▼              ▼              ▼

            PostgreSQL        Redis         Kafka
                                                 │
                                                 ▼

                      ┌────────────────────────┐
                      │    Payment Service     │
                      └──────────┬─────────────┘
                                 │
                  ┌──────────────┼──────────────┐
                  │              │
                  ▼              ▼

            PostgreSQL      Payment Gateway
                              Mock API
```

---

# 4. Technology Stack

| Component      | Technology                  |
| -------------- | --------------------------- |
| Language       | Golang                      |
| Database       | PostgreSQL 16               |
| Cache          | Redis 7                     |
| Messaging      | Kafka                       |
| Observability  | Datadog                     |
| Tracing        | OpenTelemetry               |
| Metrics        | Prometheus + Datadog        |
| Logging        | Structured JSON             |
| Deployment     | Docker Compose / Kubernetes |
| Load Generator | k6                          |

---

# 5. Services

## 5.1 Order Service

### Responsibility

Order Service merupakan entry point utama aplikasi.

### Endpoints

#### Create Order

```http
POST /api/orders
```

Request

```json
{
  "customer_id": "123",
  "product_id": "SKU-001",
  "quantity": 2
}
```

Response

```json
{
  "order_id": "ORD-10001",
  "status": "PENDING"
}
```

### Dependencies

* PostgreSQL
* Redis
* Kafka Producer

### Database Table

orders

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id VARCHAR(50),
    product_id VARCHAR(50),
    quantity INT,
    status VARCHAR(20),
    created_at TIMESTAMP
);
```

### Kafka Event

Topic

```text
order-created
```

Payload

```json
{
  "order_id": "ORD-10001",
  "product_id": "SKU-001",
  "quantity": 2
}
```

---

## 5.2 Inventory Service

### Responsibility

Inventory Service bertugas melakukan reservasi stok.

### Kafka Consumer

Consume topic

```text
order-created
```

### Processing Flow

1. Consume event
2. Check stock
3. Reserve stock
4. Update inventory
5. Publish next event

### Database Table

inventory

```sql
CREATE TABLE inventory (
    product_id VARCHAR(50),
    stock INT,
    updated_at TIMESTAMP
);
```

### Kafka Output

Topic

```text
inventory-reserved
```

Payload

```json
{
  "order_id": "ORD-10001",
  "status": "RESERVED"
}
```

---

## 5.3 Payment Service

### Responsibility

Memproses pembayaran order.

### Kafka Consumer

Consume topic

```text
inventory-reserved
```

### Processing Flow

1. Receive event
2. Call Payment Gateway
3. Update payment status
4. Update order status

### Database Table

payments

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    order_id VARCHAR(50),
    amount NUMERIC,
    status VARCHAR(20),
    created_at TIMESTAMP
);
```

---

# 6. Redis Usage

## Order Service

Key

```text
order:{order_id}
```

TTL

```text
5 Minutes
```

---

## Inventory Service

Key

```text
inventory:{product_id}
```

TTL

```text
1 Minute
```

---

# 7. Kafka Topics

| Topic              | Producer          | Consumer          |
| ------------------ | ----------------- | ----------------- |
| order-created      | Order Service     | Inventory Service |
| inventory-reserved | Inventory Service | Payment Service   |
| payment-completed  | Payment Service   | Order Service     |

---

# 8. End-to-End Transaction Flow

## Create Order Flow

```text
Client
  │
  ▼
Order Service
  │
  ├─ Insert PostgreSQL
  │
  ├─ Update Redis
  │
  └─ Publish Kafka
            │
            ▼
Inventory Service
  │
  ├─ Check Redis
  ├─ Check PostgreSQL
  ├─ Reserve Stock
  │
  └─ Publish Kafka
            │
            ▼
Payment Service
  │
  ├─ Call Payment Gateway
  ├─ Update PostgreSQL
  │
  └─ Publish Kafka
            │
            ▼
Order Service
  │
  └─ Update Status COMPLETED
```

---

# 9. Datadog Instrumentation

## Tracing

All services must instrument:

* HTTP Server
* HTTP Client
* PostgreSQL
* Redis
* Kafka Producer
* Kafka Consumer

Required Tags

```text
service
env
version
team
region
```

Example

```text
service=order-service
env=hackathon
version=1.0.0
```

---

# 10. Custom Metrics

## Order Metrics

```text
orders_created_total
orders_completed_total
orders_failed_total
```

## Inventory Metrics

```text
inventory_reserved_total
inventory_reservation_failed_total
```

## Payment Metrics

```text
payment_success_total
payment_failure_total
```

---

# 11. Incident Simulation

## Incident #1

Payment Gateway Timeout

Description

External payment API latency meningkat dari 200ms menjadi 10s.

Expected Result

* Trace latency spike
* Service map dependency issue
* Error rate increase

---

## Incident #2

Redis Failure

Description

Redis dimatikan sementara.

Expected Result

* Cache miss meningkat
* Database query meningkat
* Response time meningkat

---

## Incident #3

Slow PostgreSQL Query

Description

Menjalankan query tanpa index.

Expected Result

* Slow query detected
* Database Monitoring alert

---

## Incident #4

Kafka Consumer Lag

Description

Inventory Service diberikan delay 5 detik.

Expected Result

* Consumer lag increase
* Queue buildup
* Increased transaction latency

---

## Incident #5

Bad Deployment

Description

Deploy Payment Service v2 dengan bug.

Example

```go
panic("payment processing failed")
```

Expected Result

* Deployment marker visible
* Error spike visible
* Trace failure visible

---

# 12. Dashboards

## Executive Dashboard

* Orders per Minute
* Success Rate
* Failure Rate
* Revenue
* Payment Success Rate

## Infrastructure Dashboard

* CPU
* Memory
* Network
* Disk

## Application Dashboard

* P95 Latency
* Error Rate
* Throughput
* Kafka Lag
* PostgreSQL Queries
* Redis Hit Ratio

---

# 13. Success Criteria

The demo is considered successful when:

* Distributed traces span all three services
* Kafka traces are visible
* PostgreSQL traces are visible
* Redis traces are visible
* External API traces are visible
* Deployment tracking is visible
* Error tracking captures simulated incidents
* Service Map displays complete topology
* Business metrics are visible in Datadog dashboards
* Incident simulations generate observable impact across the platform

```
```
