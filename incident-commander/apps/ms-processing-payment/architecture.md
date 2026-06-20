# Datadog Hackathon Demo Platform

# AI-Ready E-Commerce Order Processing System

Version: 2.0

Observability Platform: Datadog Native Stack

---

# 1. Overview

Dokumen ini menjelaskan arsitektur dan implementasi platform e-commerce berbasis microservices yang dirancang khusus untuk mendemonstrasikan kemampuan observability, monitoring, troubleshooting, dan incident investigation menggunakan Datadog.

Platform ini menggunakan tiga microservices utama yang saling terintegrasi melalui PostgreSQL, Redis, Kafka, dan External Payment Gateway.

Seluruh telemetry dikirim langsung ke Datadog menggunakan Datadog Agent dan Datadog APM tanpa menggunakan Prometheus maupun OpenTelemetry Collector.

Tujuan utama platform ini adalah menghasilkan distributed tracing end-to-end dan memungkinkan simulasi berbagai incident yang dapat dianalisis menggunakan fitur-fitur Datadog.

---

# 2. Objectives

## Functional Objectives

Platform harus mampu:

* Create Order
* Reserve Inventory
* Process Payment
* Update Order Status
* Cancel Order
* View Order Status

---

## Observability Objectives

Platform harus menunjukkan:

* Distributed Tracing
* Service Map
* Deployment Tracking
* Error Tracking
* Log Correlation
* Database Monitoring
* Kafka Monitoring
* Redis Monitoring
* Infrastructure Monitoring
* Watchdog Anomaly Detection
* Custom Business Metrics
* Incident Investigation Workflow

---

# 3. High Level Architecture

```text
                           ┌───────────────┐
                           │    Client     │
                           └───────┬───────┘
                                   │
                                   ▼

                    ┌──────────────────────────┐
                    │      Order Service       │
                    │      dd-trace-go         │
                    └───────────┬──────────────┘
                                │
          ┌─────────────────────┼─────────────────────┐
          │                     │                     │
          ▼                     ▼                     ▼

      PostgreSQL            Redis                 Kafka
          │                   │                     │
          └───────────────────┴─────────────────────┘
                              │
                              ▼

                    ┌──────────────────────────┐
                    │    Inventory Service     │
                    │      dd-trace-go         │
                    └───────────┬──────────────┘
                                │
          ┌─────────────────────┼─────────────────────┐
          │                     │                     │
          ▼                     ▼                     ▼

      PostgreSQL            Redis                 Kafka
                                                       
                              │
                              ▼

                    ┌──────────────────────────┐
                    │     Payment Service      │
                    │      dd-trace-go         │
                    └───────────┬──────────────┘
                                │
                    ┌───────────┴─────────────┐
                    │                         │
                    ▼                         ▼

              PostgreSQL             Payment Gateway
                                       Mock API


                    ┌──────────────────────────┐
                    │      Datadog Agent       │
                    └───────────┬──────────────┘
                                │
      ┌─────────────────────────┼─────────────────────────┐
      │                         │                         │
      ▼                         ▼                         ▼

    Metrics                   Traces                    Logs
      │                         │                         │
      └─────────────────────────┴─────────────────────────┘
                                │
                                ▼

                     Datadog SaaS Platform
```

---

# 4. Technology Stack

| Component           | Technology                        |
| ------------------- | --------------------------------- |
| Language            | Golang 1.24                       |
| Database            | PostgreSQL 16                     |
| Cache               | Redis 7                           |
| Messaging           | Kafka                             |
| Observability       | Datadog                           |
| Tracing             | Datadog APM                       |
| Metrics             | DogStatsD                         |
| Logging             | Datadog Logs                      |
| Deployment Tracking | Datadog CI Visibility             |
| Database Monitoring | Datadog DBM                       |
| Runtime Monitoring  | Datadog Infrastructure Monitoring |
| Load Generator      | k6                                |
| Container Runtime   | Docker                            |
| Orchestration       | Docker Compose / Kubernetes       |

---

# 5. Services

## 5.1 Order Service

### Responsibilities

* Create Order
* Store Order Data
* Publish Order Event
* Update Final Order Status

---

### REST API

#### Create Order

POST /api/orders

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

---

### Dependencies

* PostgreSQL
* Redis
* Kafka Producer

---

### Database Table

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

---

### Kafka Topic

order-created

```json
{
  "order_id":"ORD-10001",
  "product_id":"SKU-001",
  "quantity":2
}
```

---

## 5.2 Inventory Service

### Responsibilities

* Validate Inventory
* Reserve Stock
* Update Inventory Database
* Publish Reservation Event

---

### Consume Topic

order-created

---

### Processing Flow

1. Receive Kafka Event
2. Check Redis Cache
3. Check PostgreSQL Inventory
4. Reserve Inventory
5. Publish Kafka Event

---

### Database Table

```sql
CREATE TABLE inventory (
    product_id VARCHAR(50),
    stock INT,
    updated_at TIMESTAMP
);
```

---

### Kafka Topic

inventory-reserved

```json
{
  "order_id":"ORD-10001",
  "status":"RESERVED"
}
```

---

## 5.3 Payment Service

### Responsibilities

* Receive Reservation Event
* Process Payment
* Call External Gateway
* Publish Payment Event

---

### Consume Topic

inventory-reserved

---

### Database Table

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

### External Dependency

Mock Payment Gateway

```http
POST /api/payment
```

---

# 6. Redis Usage

## Order Service

```text
order:{order_id}
```

TTL

```text
5 Minutes
```

---

## Inventory Service

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

```text
Client

   │

   ▼

Order Service
   │
   ├── PostgreSQL
   ├── Redis
   └── Kafka Publish

                │
                ▼

Inventory Service
   │
   ├── Redis Lookup
   ├── PostgreSQL
   └── Kafka Publish

                │
                ▼

Payment Service
   │
   ├── Payment Gateway
   ├── PostgreSQL
   └── Kafka Publish

                │
                ▼

Order Service

   └── Update Status COMPLETED
```

---

# 9. Datadog Instrumentation

## Datadog Agent

Enabled Components:

* APM
* Logs
* Process Monitoring
* Container Monitoring
* DogStatsD
* Database Monitoring

---

## Service Tags

All Services Must Include:

```text
env:hackathon
team:platform
region:ap-southeast
```

Example:

```text
service:order-service
env:hackathon
version:1.0.0
```

---

## Instrumentation Scope

### HTTP Server

* Request Duration
* Status Code
* Error Tracking

### HTTP Client

* External API Tracing
* Dependency Mapping

### PostgreSQL

* Query Tracing
* Slow Query Detection

### Redis

* Cache Hit
* Cache Miss
* Command Latency

### Kafka

* Producer Tracing
* Consumer Tracing
* Message Propagation

---

# 10. Datadog Native Features

## APM

Visualize:

* Distributed Trace
* Latency
* Throughput
* Error Rate

---

## Service Map

Expected Topology:

```text
Client
   │
Order Service
   │
Inventory Service
   │
Payment Service
   │
Payment Gateway
```

---

## Error Tracking

Automatically Capture:

* Panic
* HTTP 500
* Database Error
* Kafka Error

---

## Deployment Tracking

Track:

* Version Changes
* Error Correlation
* Latency Changes

---

## Database Monitoring

Monitor:

* Top Queries
* Slow Queries
* Query Latency
* Lock Wait
* Deadlocks

---

## Kafka Monitoring

Monitor:

* Consumer Lag
* Topic Throughput
* Broker Health
* Consumer Health

---

## Redis Monitoring

Monitor:

* Hit Ratio
* Miss Ratio
* Memory Usage
* Ops/sec

---

## Watchdog

Automatically Detect:

* Traffic Spike
* Latency Anomaly
* Error Spike
* Infrastructure Issues

---

# 11. Business Metrics

Metrics Sent Using DogStatsD

---

## Orders

```text
orders.created
orders.completed
orders.failed
```

---

## Inventory

```text
inventory.reserved
inventory.failed
```

---

## Payments

```text
payment.success
payment.failed
payment.latency
```

---

## Revenue

```text
revenue.total
```

---

# 12. Incident Simulations

## Incident #1

Payment Gateway Timeout

Description

Increase API latency:

```text
200ms → 10 seconds
```

Expected Observation

* Latency Spike
* Error Rate Increase
* Dependency Latency Increase

---

## Incident #2

Redis Failure

Description

Shutdown Redis

Expected Observation

* Cache Miss Increase
* Database Load Increase
* Request Latency Increase

---

## Incident #3

Slow Database Query

Description

Execute Query Without Index

Expected Observation

* Slow Query Detected
* DBM Alert Triggered
* Latency Increase

---

## Incident #4

Kafka Consumer Lag

Description

Add Consumer Delay

```go
time.Sleep(5 * time.Second)
```

Expected Observation

* Consumer Lag Increase
* Queue Buildup
* Processing Delay

---

## Incident #5

Bad Deployment

Description

Deploy Payment Service v2

```go
panic("payment processing failed")
```

Expected Observation

* Error Spike
* Failed Traces
* Deployment Correlation

---

## Incident #6

Database Connection Pool Exhausted

```go
db.SetMaxOpenConns(5)
```

Expected Observation

* Query Queueing
* Timeout Errors
* Latency Increase

---

## Incident #7

Memory Leak

```go
for {
   leak = append(
      leak,
      make([]byte,1024*1024),
   )
}
```

Expected Observation

* Memory Growth
* OOMKilled
* Pod Restart

---

## Incident #8

Payment Gateway HTTP 500

Expected Observation

* External Dependency Failure
* Trace Errors
* Error Tracking Events

---

# 13. Dashboards

## Executive Dashboard

Display:

* Orders Per Minute
* Revenue
* Success Rate
* Failure Rate
* Payment Success Rate

---

## Application Dashboard

Display:

* P95 Latency
* Throughput
* Error Rate
* Kafka Lag
* Redis Hit Ratio
* DB Query Duration

---

## Infrastructure Dashboard

Display:

* CPU
* Memory
* Disk
* Network
* Container Health

---

## Incident Dashboard

Display:

* Error Rate
* P95 Latency
* Deployment Timeline
* Kafka Lag
* Redis Health
* Database Health

---

# 14. Demo Scenario

## Step 1

Create Order

Show:

* Trace Generated
* Logs Correlated

---

## Step 2

Open Service Map

Show:

* Order Service
* Inventory Service
* Payment Service
* Payment Gateway

---

## Step 3

Open Trace

Show:

* PostgreSQL Span
* Redis Span
* Kafka Span
* External API Span

---

## Step 4

Deploy Payment Service v2

Register Deployment

```bash
datadog-ci deployment mark \
  --service payment-service \
  --env hackathon \
  --version v2.0.0
```

---

## Step 5

Trigger Incident

Payment Gateway Timeout

---

## Step 6

Observe

* Error Rate Spike
* Latency Spike
* Watchdog Alert
* Deployment Correlation

---

## Step 7

Root Cause Analysis

Navigate:

Service Map
→ Trace
→ Logs
→ Deployment Event

---

# 15. Success Criteria

The demo is considered successful when:

* Distributed traces span all services
* PostgreSQL traces are visible
* Redis traces are visible
* Kafka traces are visible
* External API traces are visible
* Service Map is fully connected
* Error Tracking captures failures
* Deployment Tracking correlates releases
* Watchdog detects anomalies
* Custom business metrics appear on dashboards
* Incident simulations create measurable impact
* Root cause can be identified within Datadog in less than 5 minutes

END OF DOCUMENT
