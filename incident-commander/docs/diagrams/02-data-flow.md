# 02 Data Flow & Sequence Diagrams

> Sequence diagram mendetail per step: interaksi antar komponen, pesan HTTP, query DB, dan LLM call.

---

## 2.1 Sequence Diagram: Trigger Incident (Outage)

```text
           Client       FastAPI    LangGraph    InfraAI     AppAI      ChangeAI    BizAI     Postgres   Slack      Datadog
             │            │           │            │           │            │           │           │           │          │
             ├─POST /incident/trigger───────────────────────────────│──────────│──────────│───────────│───────────│──────────│
             │            │           │            │           │            │           │           │           │          │
             │            ├──────────►│            │           │            │           │           │           │          │
             │            │ state = init│            │           │            │           │           │           │          │
             │            │           ├─query INSERT incidents────────────────────────────────────────────────────────│
             │            │           │            │           │            │           │           ├─INSERT────►│          │
             │            │           │◄───────────│───────────│────────────│───────────│───────────│           │          │
             │            │           │            │           │            │           │           │           │          │
             │            │           │ send_metrics("aic.incident.created")──────────────────────────────────────────────────────────────────►
             │            │           │            │           │            │           │           │           │          │
             │            │◄───────── incident_id = INC-001 ─────────────────────────────────────────────────────────────│
             │            │           │            │           │            │           │           │           │          │
             │◄─────────── 200 OK {id: "INC-001"}───────────────────────────────────────────────────────────────────────│
             │            │           │            │           │            │           │           │           │          │
             │            │           │            │           │            │           │           │           │          │
             │            │           │ ▶ Parallel Fan-Out ◀│────────────│───────────│───────────│           │          │
             │            │           │            │           │            │           │           │           │          │
             │            │           ├────────────►│           │            │           │           │           │          │
             │            │           │            ├─POST OpenRouter (infra prompt)─────────────────────────────────────────────────────►
             │            │           │            │ LLM response (12s)          │            │           │          │
             │            │           │            │◄───── JSON finding ────────────────────────────────────────────────────────────────
             │            │           │            │ Parse + save to findings     │            │           │          │
             │            │           │            ├─query INSERT findings───────────────────────────────────────────────────────────►
             │            │           │            │           ├─POST OpenRouter (app prompt)─────────────────────────────────────────────────────►
             │            │           │            │           │ LLM response (9s)           │            │          │
             │            │           │            │           │◄───── JSON finding ────────────────────────────────────────────────────────────
             │            │           │            │           │─query INSERT findings───────────────────────────────────────────────────────────►
             │            │           │            │           │            ├─POST OpenRouter (change prompt)─────────────────────────────────────────────────────►
             │            │           │            │           │            │ LLM response (8s)           │          │
             │            │           │            │           │            │◄───── JSON finding ────────────────────────────────────────────────────────────
             │            │           │            │           │            │─query INSERT findings───────────────────────────────────────────────────────────►
             │            │           │            │           │            │           ├─POST OpenRouter (biz prompt)──────────────────────────────────────────────────────►
             │            │           │            │           │            │           │ LLM response (7s)                   │          │
             │            │           │            │           │            │           │◄───── JSON finding ────────────────────────────────────────────────────────
             │            │           │            │           │            │           │─query INSERT findings───────────────────────────────────────────────────────────►
             │            │           │            │           │            │           │           │           │          │
             │            │           │ ◀ Join: collect 4 findings ◀────────────────────────────────────────────────────│
             │            │           │            │           │            │           │           │           │          │
             │            │           │ Decision Engine: aggregate findings → LLM decision ──────────────────────────────────────────────────────►
             │            │           │            │           │            │           │           │           │          │
             │            │           │ ◄ LLM returns recommendation, confidence, risk ◀─────────────────────────────────────────────────────
             │            │           │            │           │            │           │           │           │          │
             │            │           │ Save recommendation to table ───────────────────────────────────────────────────────────────────────────►
             │            │           │            │           │            │           │           │           │          │
             │            │           │ send_metrics("aic.agent.confidence") ──────────────────────────────────────────────────────────────────────────►
             │            │           │            │           │            │           │           │           │          │
             │            │           │ Build Slack message ────────────────────────────────────────────────────────────────────────►
             │            │           │            │           │            │           │           │           ├─POST Webhook───►
             │            │           │            │           │            │           │           │           │◄───200 OK──│
             │            │           │            │           │            │           │           │           │          │
             │            │           │ Emit structured log──────────────────────────────────────────────────────────────────────────────────────────────────►
             │            │           │            │           │            │           │           │           │          │
             │            │◄───────── State: {"status": "resolved", "recommendation": "rollback"} ─────────────────────────────│
             │            │           │            │           │            │           │           │           │          │
    
    Total Latency Target: < 45 detik (paling lambat LLM call + DB writes + Slack POST)
```

---

## 2.2 Sequence Diagram: Approval Flow

```text
                              Slack                FastAPI           LangGraph         DB
                                │                    │                  │              │
                                │                    │                  │              │
    [Agent posting approval request]
                                │                    │                  │              │
                                ◄──────────────────── Slack Thread dengan tombol/menu ──
                                │                    │                  │              │
    [Human Engineer Member tekan approve via webhook]
                                │                    │                  │              │
                                ├─POST /incident/{id}/approve───────────────────────────│
                                │                    │                  │              │
                                │                    ├──────────────────►│              │
                                │                    │  Update state.approved = true   │
                                │                    │                  │              │
                                │                    │  Update approval table          │
                                │                    │                  ├─INSERT─────►│
                                │                    │                  │              │
                                │                    │                  │◄─────────────│
                                │                    │                  │              │
                                │                    │  Resume orchestration (opsional)│
                                │                    │                  │              │
                                │                    │                  │  Execute action │
                                │                    │                  │  (opsional DD workflow)
                                │                    │                  │              │
                                │                    │ Send response: {"approved": true} │
                                │                    │                  │              │
                                │◄─────────────────── 200 OK ────────────────────────────
                                │                    │                  │              │
```

> Note: untuk demo 4 jam, approval sebaiknya **simulated** via cURL (`POST /approve`) tanpa UI interaktif Slack.

---

## 2.3 Sequence Diagram: Query Incident History

```text
    Client          FastAPI            Postgres
      │                │                   │
      ├──GET /incident/INC-001─────────────│
      │                │                   │
      │                ├─SELECT * FROM incidents WHERE id = 'INC-001'─►
      │                │                   │
      │                │◄─────────────────│
      │                │                   │
      │                ├─SELECT * FROM findings WHERE incident_id = 'INC-001'─►
      │                │                   │
      │                │◄─────────────────│
      │                │                   │
      │                ├─SELECT * FROM timeline WHERE incident_id = 'INC-001'─►
      │                │                   │
      │                │◄─────────────────│
      │                │                   │
      │◄───────────────{ incident + findings + timeline } ──│
```

---

## 2.4 State Machine Diagram (LangGraph)

```text
    ┌──────────┐
    │  START   │
    └────┬─────┘
         │ trigger received
         ▼
    ┌────────────────┐
    │ incident_start │
    │  - assign ID   │
    │  - persist     │
    └────┬───────────┘
         │
         ▼
    ┌────────────────────────────┐
    │    PARALLEL_EXECUTION       │
    │  [infrastructure_agent]     │
    │  [application_agent]        │
    │  [change_correlation_agent] │
    │  [business_impact_agent]    │
    └────┬────────────────────────┘
         │ after all complete
         ▼
    ┌────────────────┐
    │ decision_engine│
    │  - collect     │
    │  - aggregate   │
    │  - recommend   │
    └────┬───────────┘
         │
         ▼
    ┌────────────────┐
    │  approval_gate │──────(if required)─────► [PENDING_APPROVAL]
    │                │                                    │
    │                │◄──human approves via /approve─────┘
    └────┬───────────┘
         │ approved
         ▼
    ┌────────────────┐
    │  slack_notify  │
    │  - post thread │
    └────┬───────────┘
         │
         ▼
    ┌────────────────┐
    │   RESOLVED     │
    │  - generate RCA│
    │  - update KB   │
    └────────────────┘
```

---

## 2.5 Data Transformation Across Flow

```text
Input Webhook Payload (Datadog Alert):
─────────────────────────────────────
{
  "alert_type": "outage",
  "service": "payment-api",
  "severity": "SEV1",
  "metric": "avg:p99_latency",
  "threshold": 500,
  "value": 2100,
  "datadog_url": "https://app.datadoghq.com/monitors/12345"
}

↓

LangGraph State:
─────────────────────────────────────
{
  "incident_id": "INC-2026-001",
  "service": "payment-api",
  "type": "outage",
  "severity": "SEV1",
  "status": "investigating",
  "findings": [],
  "recommendation": null,
  "approval_required": false,
  "approved": false,
  "created_at": "2026-06-14T10:00:00Z",
  "datadog_url": "https://app.datadoghq.com/monitors/12345"
}

↓

Agent Prompt Context (injected mock):
─────────────────────────────────────
You are the Infrastructure & Platform AI.
Service: payment-api
Incident type: outage (SEV1)
Observability context:
- Kubernetes deployment v2.1.5 rolled out 8 minutes ago
- CPU usage jumped from 45% to 95%
- Memory stable at 60%
- 3 pods in CrashLoopBackOff
- No node failures detected
Analyze and return JSON: {"finding": "...", "confidence": 0-100}

↓

Agent LLM Response (JSON):
─────────────────────────────────────
{
  "agent": "infrastructure",
  "finding": "Deployment v2.1.5 correlated with CPU spike and pod crash loops. Likely resource limit regression.",
  "confidence": 92,
  "suggested_action": "rollback deployment to v2.1.4",
  "evidence": ["k8s_deployment_event", "cpu_metrics"]
}

↓

Decision Engine Aggregation:
─────────────────────────────────────
{
  "root_cause": "Deployment v2.1.5 introduced CPU regression due to unbounded goroutines.",
  "contributing_factors": ["No resource limit on new worker pool", "HPA max replicas reached"],
  "recommended_action": "rollback deployment to v2.1.4",
  "confidence": 94,
  "risk": "LOW",
  "approval_required": true,
  "affected_users": 14500,
  "revenue_exposure": "$25000/hour",
  "recovery_time_estimate": "15 minutes"
}

↓

Slack Thread Message:
─────────────────────────────────────
🔴 INC-2026-001 | payment-api | SEV1

🤖 Infrastructure AI:
> Deployment v2.1.5 correlated with CPU spike and pod crash loops.
> Confidence: 92%

🤖 Application AI:
> DB connection pool exhausted. 4500 timeout errors in last 10 minutes.
> Confidence: 94%

🤖 Business Impact AI:
> Affected Users: 14,500
> Revenue Exposure: $25,000/hour
> SLA Breach Risk: HIGH

👨‍✈️ Incident Lead AI:
> RECOMMENDATION: Rollback deployment to v2.1.4
> Confidence: 94% | Risk: LOW
> ⚠️ APPROVAL REQUIRED
```

---

## 2.6 Latency Budget

| Step | Target Latency | Keterangan |
|------|---------------|------------|
| FastAPI receive → LangGraph start | < 100ms | Sync overhead |
| PostgreSQL INSERT incident | < 50ms | Local Docker PG |
| LLM call per agent (OpenRouter via HTTP) | 5-15s | Paralel, max 15s timeout |
| Aggregate + Decision LLM | 3-8s | Single LLM call |
| PostgreSQL INSERT findings × 4 | < 100ms total | Batch prepared stmt |
| Slack POST via Webhook | < 500ms | Async fire-and-forget |
| Total end-to-end | **< 30 detik** | Asumsi paralel optimal |

> Untuk demo, kita target total latency **< 30 detik** dari trigger sampai Slack message final terposting.

---

> Next: baca `03-file-structure.md` untuk blueprint lengkap semua file yang harus dibuat.
