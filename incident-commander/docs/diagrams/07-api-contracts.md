# 07 API Contracts & Endpoints

> Spesifikasi lengkap API endpoints, request/response schemas, dan error handling.

---

## 7.1 Base URL

```
http://localhost:8000
```

Health Check:
```
GET /health
→ {"status": "ok", "version": "0.1.0"}
```

---

## 7.2 Endpoints

### 1. Trigger Incident

```
POST /incident/trigger
```

**Description:** Memulai incident investigation baru. LangGraph state machine akan berjalan secara async.

**Request Body (Pydantic: TriggerPayload):**

```json
{
  "type": "outage",
  "service": "payment-api",
  "severity": "SEV1",
  "datadog_alert": {
    "metric": "avg:p99_latency",
    "threshold": 500,
    "value": 2100,
    "url": "https://app.datadoghq.com/monitors/12345"
  },
  "mock_context_type": "outage"  // optional: "outage", "billing", "generic"
}
```

**Fields:**
- `type` (str, required): `outage`, `billing`, `database`, `slo_breach`, `custom`
- `service` (str, required): nama service yang affected
- `severity` (str, required): `SEV1`, `SEV2`, `SEV3`, `SEV4`
- `datadog_alert` (dict, required): metadata alert dari Datadog
- `mock_context_type` (str, optional): override mock context yang digunakan

**Response 200 (TriggerResponse):**

```json
{
  "incident_id": "INC-2026-001",
  "status": "investigating",
  "message": "Investigation started. Check /incident/INC-2026-001 for status.",
  "estimated_duration_seconds": 30
}
```

**Response 400 (BadRequest):**

```json
{
  "error": "Invalid incident_type: 'outagee'. Must be one of: outage, billing, database, slo_breach, custom"
}
```

**Response 500 (ServerError):**

```json
{
  "error": "Failed to start investigation",
  "detail": "OpenRouter API timeout after 30s"
}
```

---

### 2. Get Incident Detail

```
GET /incident/{incident_id}
```

**Response 200 (IncidentDetail):**

```json
{
  "id": "INC-2026-001",
  "incident_type": "outage",
  "service": "payment-api",
  "severity": "SEV1",
  "status": "resolved",
  "datadog_url": "https://app.datadoghq.com/monitors/12345",
  "created_at": "2026-06-14T10:00:00Z",
  "closed_at": "2026-06-14T10:45:00Z",
  "duration_seconds": 2700,
  "findings": [
    {
      "agent_name": "infrastructure",
      "finding": "Deployment v2.1.5 correlated with CPU spike and pod crash loops. Likely resource limit regression.",
      "confidence": 92,
      "suggested_action": "rollback deployment to v2.1.4",
      "evidence": ["k8s_deployment_event", "cpu_metrics"],
      "created_at": "2026-06-14T10:00:15Z"
    }
    // ... 3 more findings
  ],
  "recommendation": {
    "recommendation": "rollback deployment to v2.1.4",
    "root_cause": "Deployment v2.1.5 introduced CPU regression due to unbounded goroutines.",
    "confidence": 94,
    "risk": "LOW",
    "approval_required": true,
    "affected_users": 14500,
    "revenue_exposure": "$25000/hour",
    "recovery_time_estimate": "15 minutes",
    "created_at": "2026-06-14T10:00:30Z"
  },
  "approval": {
    "status": "approved",
    "approver": "engineer@company.com",
    "responded_at": "2026-06-14T10:05:00Z"
  }
}
```

**Response 404:**

```json
{
  "error": "Incident not found",
  "incident_id": "INC-999-999"
}
```

---

### 3. Get Incident Timeline

```
GET /incident/{incident_id}/timeline
```

**Response 200:**

```json
{
  "incident_id": "INC-2026-001",
  "events": [
    {
      "actor": "system",
      "event": "Incident created via trigger",
      "timestamp": "2026-06-14T10:00:00Z",
      "metadata": {"source": "api", "trigger_type": "outage"}
    },
    {
      "actor": "infrastructure_agent",
      "event": "Investigation completed: deployment correlation found",
      "timestamp": "2026-06-14T10:00:15Z",
      "metadata": {"confidence": 92}
    },
    {
      "actor": "decision_engine",
      "event": "Recommendation generated: rollback deployment",
      "timestamp": "2026-06-14T10:00:30Z",
      "metadata": {"confidence": 94, "risk": "LOW"}
    },
    {
      "actor": "system",
      "event": "Approval requested via Slack",
      "timestamp": "2026-06-14T10:00:30Z",
      "metadata": {"approval_required": true}
    },
    {
      "actor": "human",
      "event": "Action approved",
      "timestamp": "2026-06-14T10:05:00Z",
      "metadata": {"approver": "engineer@company.com"}
    }
  ]
}
```

---

### 4. List Active Incidents

```
GET /incidents/active
```

**Response 200:**

```json
{
  "count": 3,
  "incidents": [
    {
      "id": "INC-2026-001",
      "service": "payment-api",
      "severity": "SEV1",
      "status": "investigating",
      "age_minutes": 12
    }
  ]
}
```

---

### 5. Submit Approval

```
POST /incident/{incident_id}/approve
```

**Request Body (ApprovalRequest):**

```json
{
  "approver": "engineer@company.com",
  "action": "rollback deployment",
  "approved": true,
  "reason": "Confidence 94%, rollback is safe"
}
```

**Response 200 (ApprovalResponse):**

```json
{
  "incident_id": "INC-2026-001",
  "status": "approved",
  "approver": "engineer@company.com",
  "action": "rollback deployment",
  "timestamp": "2026-06-14T10:05:00Z",
  "message": "Action approved. Incident updated to approved status."
}
```

**Response 409 (Already Approved/Rejected):**

```json
{
  "error": "Approval already exists for this incident",
  "current_status": "approved"
}
```

---

### 6. Resolve Incident

```
POST /incident/{incident_id}/resolve
```

**Request Body:**

```json
{
  "resolution": "Rollback deployment v2.1.5 to v2.1.4 completed. Latency returned to baseline.",
  "closed_by": "ai-system"
}
```

**Response 200:**

```json
{
  "incident_id": "INC-2026-001",
  "status": "resolved",
  "resolution": "Rollback deployment v2.1.5 to v2.1.4 completed...",
  "duration_seconds": 2700,
  "rca_generated": true,
  "rca_preview": "Root cause: v2.1.5 introduced unbounded goroutine pool..."
}
```

---

## 7.3 OpenAPI Schema

Generated automatically oleh FastAPI. Akses via:

```
http://localhost:8000/docs       # Swagger UI
http://localhost:8000/redoc      # ReDoc
http://localhost:8000/openapi.json # Raw schema
```

---

## 7.4 Error Response Standard

Semua error responses mengikuti format:

```json
{
  "error": "Human-readable error message",
  "detail": "Optional technical detail",
  "code": "ERROR_CODE",
  "timestamp": "2026-06-14T10:00:00Z"
}
```

Error codes:
- `INVALID_PAYLOAD`: 400 Bad Request
- `INCIDENT_NOT_FOUND`: 404 Not Found
- `APPROVAL_CONFLICT`: 409 Conflict
- `LLM_TIMEOUT`: 504 Gateway Timeout
- `DATABASE_ERROR`: 503 Service Unavailable

---

## 7.5 Postman / cURL Examples

### Trigger Outage

```bash
# Outage scenario
curl -X POST http://localhost:8000/incident/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "type": "outage",
    "service": "payment-api",
    "severity": "SEV1",
    "datadog_alert": {
      "metric": "avg:p99_latency",
      "threshold": 500,
      "value": 2100,
      "url": "https://app.datadoghq.com/monitors/12345"
    },
    "mock_context_type": "outage"
  }'
```

### Trigger Billing Anomaly

```bash
# Billing anomaly scenario
curl -X POST http://localhost:8000/incident/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "type": "billing",
    "service": "payment-api",
    "severity": "SEV2",
    "datadog_alert": {
      "metric": "aws.daily_cost",
      "expected": 300,
      "actual": 1200,
      "url": "https://app.datadoghq.com/monitors/67890"
    },
    "mock_context_type": "billing"
  }'
```

### Check Status

```bash
curl http://localhost:8000/incident/INC-2026-001
```

### Approve Action

```bash
curl -X POST http://localhost:8000/incident/INC-2026-001/approve \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "engineer@company.com",
    "approved": true,
    "reason": "Rollback is safe, confidence 94%"
  }'
```

### Resolve

```bash
curl -X POST http://localhost:8000/incident/INC-2026-001/resolve \
  -H "Content-Type: application/json" \
  -d '{
    "resolution": "Deployment rollback completed successfully.",
    "closed_by": "ai-system"
  }'
```

---

## 7.6 Async Architecture Detail

Investigation di-trigger secara **async background**:

1. API menerima trigger → return immediately dengan `incident_id`
2. Background task (FastAPI `BackgroundTasks`) menjalankan LangGraph state machine
3. Client poll `GET /incident/{id}` untuk cek progress
4. Alternatif: webhook callback (opsional)

```python
# In router
@router.post("/incident/trigger")
async def trigger_incident(
    payload: TriggerPayload,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db)
):
    incident = create_incident(db, payload)
    background_tasks.add_task(run_investigation, incident.id)
    return {"incident_id": incident.id, "status": "investigating"}
```

---

> Next: baca `08-datadog-metrics.md` untuk observability strategy detail.
