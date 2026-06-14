# 01 Technical Architecture & Data Flow

> Berisi diagram arsitektur end-to-end dan penjelasan alur data antar layer.

---

## 1.1 High Level System Diagram

```text
                    ┌──────────────────────┐
                    │   Trigger Source     │
                    │  (cURL / Postman /   │
                    │   Datadog Webhook)   │
                    └──────────┬───────────┘
                               │
                               ▼
                    ┌──────────────────────┐
                    │  FastAPI Gateway       │
                    │  (main.py)             │
                    │  POST /incident/trigger│
                    └──────────┬───────────┘
                               │
                               ▼
              ┌────────────────────────────────────┐
              │    LangGraph State Machine           │
              │  (orchestrator/state_machine.py)     │
              │                                      │
              │  ┌─────────┐  ┌────────────┐        │
              │  │  Node   │  │   Node     │        │
              │  │ incident│  │  infra_    │ Parallel│
              │  │ _start  │  │  agent     │ Nodes   │
              │  └────┬────┘  └─────┬──────┘        │
              │       │             │               │
              │       ▼             ▼               │
              │  ┌─────────┐  ┌────────────┐        │
              │  │  Node   │  │   Node     │        │
              │  │  app_   │  │  change_   │        │
              │  │  agent  │  │  agent     │        │
              │  └────┬────┘  └─────┬──────┘        │
              │       │             │               │
              │       ▼             ▼               │
              │  ┌─────────┐  ┌────────────┐        │
              │  │  Node   │  │   Node     │        │
              │  │business │  │  decision  │        │
              │  │ _agent  │  │ _engine    │         │
              │  └─────────┘  └─────┬──────┘        │
              │                     │               │
              │                     ▼               │
              │            ┌────────────┐           │
              │            │   Node     │           │
              │            │  approval  │           │
              │            │   _gate    │           │
              │            │            │           │
              │            └─────┬──────┘           │
              │                  │                   │
              │                  ▼                   │
              │          ┌────────────┐              │
              │          │   Node     │              │
              │          │slack_notify│              │
              │          └────────────┘              │
              └──────────────────────────────────────┘
                               │
               ┌───────────────┴───────────────┐
               ▼                               ▼
    ┌──────────────────────┐        ┌──────────────────────┐
    │ Postgres (Docker)    │        │ Slack Webhook        │
    │ - incidents            │        │ - Thread messages    │
    │ - findings             │        │ - Agent post         │
    │ - timeline             │        │ - Approval buttons   │
    │ - recommendations      │        │   (opsional)         │
    └──────────────────────┘        └──────────────────────┘
               │
               ▼
    ┌──────────────────────┐
    │ Datadog Agent (Docker)│
    │ DogStatsD UDP :8125    │
    │ Log Agent stdout        │
    │ → APM (opt)            │
    └──────────────────────┘
```

---

## 1.2 Incident Data Flow (Detailed)

```text
┌──────────────────────────────────────────────────────────────────┐
│                         INCIDENT LIFECYCLE                        │
└──────────────────────────────────────────────────────────────────┘

┌─────────────┐
│  Watchdog   │
│  Alert      │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────┐
│ FastAPI receives alert payload                │
│ ┌──────────────────────────────────────┐     │
│ │ {                                    │     │
│ │   "type": "outage|billing",         │     │
│ │   "service": "payment-api",         │     │
│ │   "severity": "SEV1|SEV2",          │     │
│ │   "datadog_alert": { ... }           │     │
│ │ }                                    │     │
│ └──────────────────────────────────────┘     │
└─────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ LangGraph: incident_start                                        │
│ - Generate incident_id (INC-{YYYY}-{NNN})                       │
│ - Insert row: incidents table                                  │
│ - Emit: aic.incident.created [type: outage, severity: SEV1]    │
│ - Emit: timeline entry (actor: system, event: incident_created)│
└──────────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ PARALLEL NODE EXECUTION (LangGraph Parallel Fan-Out)            │
│                                                                 │
│  Node 1: infrastructure_agent                                  │
│  ├─ Query LLM dengan system_prompt khusus infra                │
│  ├─ Input: service_name, incident_type, mock_observability_ctx │
│  ├─ Output: finding JSON {agent:"infra", confidence: 92}       │
│  ├─ Write: findings table                                       │
│  ├─ Write: timeline table                                      │
│  └─ Emit: aic.agent.execution.duration [agent: infra]          │
│                                                                 │
│  Node 2: application_agent                                     │
│  ├─ Query LLM dengan system_prompt khusus application          │
│  ├─ Input: service_name, incident_type, mock_context            │
│  ├─ Output: {agent:"app", root_cause:"db_timeout"}             │
│  ├─ Write: findings table                                       │
│  ├─ Write: timeline table                                      │
│  └─ Emit: aic.agent.confidence [agent: app]                   │
│                                                                 │
│  Node 3: change_correlation_agent                              │
│  ├─ Query LLM dengan system_prompt khusus change                │
│  ├─ Input: service_name, incident_type                         │
│  ├─ Output: {agent:"change", change_prob: 92.0}              │
│  ├─ Write: findings table                                       │
│  ├─ Write: timeline table                                      │
│  └─ Emit: aic.agent.execution.duration [agent: change]        │
│                                                                 │
│  Node 4: business_impact_agent                                 │
│  ├─ Query LLM dengan system_prompt khusus business            │
│  ├─ Input: service_name, severity                              │
│  ├─ Output: {agent:"business", revenue_loss: "$25000/hour"}    │
│  ├─ Write: findings table                                       │
│  ├─ Write: timeline table                                      │
│  └─ Emit: aic.agent.execution.duration [agent: business]      │
└──────────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ LangGraph: decision_engine                                      │
│ - Collect all 4 findings dari state graph                      │
│ - Query LLM dengan system_prompt "Incident Commander Decision"  │
│ - Generate:                                                      │
│   ┌─────────────────────────────────────────────────────────────┐│
│   │ recommendation: "rollback deployment"                     ││
│   │ confidence: 94                                             ││
│   │ risk: "LOW"                                               ││
│   │ approval_required: true                                    ││
│   │ affected_users: 14500                                      ││
│   │ revenue_exposure: "$25000/hour"                             ││
│   └─────────────────────────────────────────────────────────────┘│
│ - Write to: recommendations table                                │
│ - Emit: aic.incident.approval.required [severity: SEV1]        │
└──────────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ LangGraph: approval_gate                                        │
│ - IF approval_required AND action class = restricted:          │
│   → Set state: pending_approval = true                           │
│   → Emit: Slack message: "[APPROVAL REQUIRED] ..."             │
│   → Wait (opsional loopback ke API approve endpoint)           │
│ - ELSE:                                                          │
│   → state: approved = true                                      │
│   → Emit: "Action auto-approved. Executing..."                │
└──────────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ LangGraph: slack_notify                                         │
│ - Format final Slack message dari state findings              │
│ - POST to Incoming Webhook                                     │
│ - Output: Slack thread dengan semua findings ter-thread       │
└──────────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ Resolution (Manual / Opsional Automate)                       │
│ - Human mengupdate /incident/{id}/resolve                      │
│ - Write: rca (auto generated oleh LLM)                           │
│ - Emit: aic.incident.closed                                      │
│ - Emit: aic.incident.duration [duration_seconds: 1800]           │
└──────────────────────────────────────────────────────────────────┘
```

---

## 1.3 Layered Architecture

```text
┌──────────────────────────────────────────────────────────────────┐
│                     LAYER 1: GATEWAY / API                       │
│  FastAPI (async)                                                 │
│  - POST /incident/trigger                                        │
│  - GET /incident/{id}                                            │
│  - GET /incident/{id}/findings                                   │
│  - POST /incident/{id}/approve                                   │
│  - GET /incident/{id}/timeline                                     │
│  - DELETE /incident/{id}/resolve (opsional)                     │
└──────────────────────────────────────────────────────────────────┘
                              │
┌──────────────────────────────────────────────────────────────────┐
│                  LAYER 2: ORCHESTRATION                          │
│  LangGraph State Machine                                        │
│  - StateDefinition (TypedDict)                                  │
│  - NodeRegistry                                                  │
│  - Conditional Routing (approval → wait / skip)                  │
│  - Parallel Execution Manager                                   │
└──────────────────────────────────────────────────────────────────┘
                              │
┌──────────────────────────────────────────────────────────────────┐
│                 LAYER 3: AI AGENTS                               │
│  Each Agent = System Prompt + LLM Call + JSON Parser           │
│  - IncidentLeadAgent (orchestrator)                             │
│  - InfrastructureAgent                                         │
│  - ApplicationAgent                                             │
│  - ChangeCorrelationAgent                                       │
│  - BusinessImpactAgent                                          │
│  - DecisionEngine (aggregator)                                 │
└──────────────────────────────────────────────────────────────────┘
                              │
┌──────────────────────────────────────────────────────────────────┐
│              LAYER 4: INTEGRATION & PERSISTENCE                  │
│  Database: PostgreSQL (SQLAlchemy ORM)                         │
│  Slack: Webhook POST                                            │
│  Datadog: DogStatsD + structured logging                       │
│  LLM: OpenRouter HTTP API (Claude / GPT-4o)                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 1.4 Parallel Agent Fan-Out Detail

```text
LangGraph Orchestrator State:

State: {
  "incident_id": "INC-2026-001",
  "service": "payment-api",
  "type": "outage",
  "severity": "SEV1",
  "status": "investigating",
  "findings": [],
  "recommendation": None,
  "approval_required": False,
  "approved": False,
  "created_at": "2026-06-14T10:00:00Z"
}

                     ┌─────────────────────────────────────┐
                     │         START NODE                  │
                     │   incident_start()                  │
                     │   - assign ID                       │
                     │   - insert to DB                    │
                     └──────────────┬──────────────────────┘
                                    │
                                    ▼
                     ┌─────────────────────────────────────┐
                     │        PARALLEL BRANCH               │
                     │   build_parallel_nodes()             │
                     └──────────────┬──────────────────────┘
                     ┌──────────────┼──────────────┬──────────────┐
                     ▼              ▼              ▼              ▼
            ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
            │ infra_   │   │ app_     │   │ change_  │   │ business │
            │ agent    │   │ agent    │   │ agent    │   │ _agent   │
            └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
                 │              │              │              │
                 ▼              ▼              ▼              ▼
            ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
            │ POST   │   │ POST   │   │ POST   │   │ POST   │
            │ LLM    │   │ LLM    │   │ LLM    │   │ LLM    │
            │ (12s)  │   │ (9s)   │   │ (8s)   │   │ (7s)   │
            └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
                 │              │              │              │
                 ▼              ▼              ▼              ▼
            ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
            │ parse    │   │ parse    │   │ parse    │   │ parse    │
            │ JSON     │   │ JSON     │   │ JSON     │   │ JSON     │
            └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
                 └──────────────┴──────────────┴──────────────┘
                                     │
                                     ▼
                     ┌─────────────────────────────────────┐
                     │         JOIN NODE                   │
                     │   collect_findings()                │
                     │   Write to state.findings[]         │
                     └──────────────┬──────────────────────┘
                                    │
                                    ▼
                     ┌─────────────────────────────────────┐
                     │      DECISION ENGINE                │
                     │   aggregate + LLM decision           │
                     │   - root cause                       │
                     │   - recommendation                   │
                     │   - confidence                       │
                     └──────────────┬──────────────────────┘
                                    │
                                    ▼
                     ┌─────────────────────────────────────┐
                     │         APPROVAL GATE               │
                     │   conditional routing               │
                     │   - if required → wait             │
                     │   - else → auto-approve              │
                     └──────────────┬──────────────────────┘
                                    │
                                    ▼
                     ┌─────────────────────────────────────┐
                     │         SLACK NOTIFY               │
                     │   build_message()                   │
                     │   POST to webhook                   │
                     └─────────────────────────────────────┘
```

---

## 1.5 Skenario Use Case Mapping

### A. Production Outage

```text
Trigger: latency spike detected in payment-api

Agent Output:
├── infra_agent:
│   └── "deployment v2.1.5 detected 8 min ago, CPU spike +300%"
├── app_agent:
│   └── "DB connection pool exhausted, 4500 timeout errors logged"
├── change_agent:
│   └── "deployment change correlation: 96%"
└── business_agent:
    └── "14,500 users affected, $25,000/hour revenue exposure"

Decision:
  recommendation: "rollback deployment v2.1.5 immediately"
  confidence: 94
  risk: LOW
  approval_required: true
```

### B. AWS Billing Anomaly

```text
Trigger: daily cost spike from $300 → $1,200

Agent Output:
├── infra_agent:
│   └── "Aurora db.r6.large → db.r6.4xlarge, cost contribution 65%"
├── app_agent:
│   └── "retry storm: 8M additional requests detected in traces"
├── change_agent:
│   └── "resource change detected: Aurora scaling event 3 hours ago"
└── business_agent:
    └── "additional $900/day, projected $27,000/month"

Decision:
  recommendation: "reduce Aurora instance + fix retry config"
  confidence: 94
  risk: MEDIUM
  approval_required: true
```

---

> Next: baca `02-data-flow.md` untuk diagram sequence detail tiap use case.
