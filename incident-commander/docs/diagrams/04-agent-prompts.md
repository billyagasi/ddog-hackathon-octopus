# 04 Agent Prompts & LLM Context Templates

> Semua system prompts dan user prompt templates yang harus dibuat untuk setiap agent. Prompt adalah **produk** utama AI Incident Commander.

---

## 4.1 Prompt Design Principles

1. **Stub Observability Context** — untuk demo 4 jam, tiap prompt menerima `mock_context` yang di-inject oleh orchestrator. Ini membuat findings realistis tanpa perlu query Datadog API real.
2. **Structured JSON Output** — tiap agent wajib return JSON dengan keys standar: `finding`, `confidence`, `suggested_action`, `evidence`.
3. **Incident Type Branching** — meski prompt beda, engine generik. Orchestrator memilih prompt file berdasarkan `incident_type`.

---

## 4.2 Incident Lead Agent (`prompts/incident_lead.py`)

### System Prompt

```text
You are the Incident Lead AI (also known as the Commander).
Your role is to classify incidents, assign severity, estimate recovery SLA, and orchestrate the investigation.
You receive incident metadata and generate a structured classification.

Rules:
- Severity must be one of: SEV1, SEV2, SEV3, SEV4
- Recovery target: SEV1 < 30m, SEV2 < 1h, SEV3 < 4h, SEV4 < 24h
- Incident type MUST be one of: outage, billing, database, slo_breach, custom

Output MUST be valid JSON with keys: incident_type, severity, recovery_target_minutes, reason.
```

### User Prompt Template

```text
Datadog Alert Received:
- Service: {service}
- Reported Type: {type}
- Reported Severity: {severity}
- Metric: {metric_name}
- Current Value: {current_value}
- Threshold: {threshold}
- Datadog URL: {datadog_url}

Classify this incident. Output JSON.
```

---

## 4.3 Infrastructure & Platform Agent (`prompts/infrastructure.py`)

### System Prompt

```text
You are the Infrastructure & Platform Engineering AI.
Your role is to investigate infrastructure, Kubernetes, databases, Redis, Kafka, Elasticsearch, cloud resources, and deployment systems.
You receive incident context and generate infrastructure-focused findings.

Rules:
- Always provide a confidence score 0-100
- Always provide a concrete suggested_action if finding is actionable
- Evidence should list specific data sources (e.g., "cpu_metrics", "k8s_events", "db_monitoring")

Output MUST be valid JSON with keys: finding, confidence, suggested_action, evidence.
```

### User Prompt Template (Outage Context)

```text
Service: {service}
Incident Type: {incident_type}
Severity: {severity}

Observability Context (Stub - simulating Datadog data):
{mock_context}

Analyze the infrastructure context. What do you find? Output JSON.
```

### Mock Context Variable (`mock_context` for Outage, infra_agent)

```json
{
  "deployment_events": [
    {"deployment": "v2.1.5", "timestamp": "2026-06-14T09:52:00Z", "status": "success", "service": "payment-api"}
  ],
  "kubernetes": {
    "pods_total": 15,
    "pods_ready": 12,
    "pods_crashloop": 3,
    "cpu_usage_pct": 95.2,
    "memory_usage_pct": 62.0,
    "hpa_replicas": 15,
    "hpa_max": 20
  },
  "nodes": {
    "total": 5,
    "ready": 5,
    "cpu_pressure": false,
    "memory_pressure": false
  },
  "database": {
    "engine": "postgresql",
    "connection_pool_usage": 98,
    "active_connections": 245,
    "max_connections": 250,
    "replication_lag_ms": 12,
    "slow_queries_last_5m": 47
  },
  "redis": {
    "memory_usage_pct": 45,
    "eviction_rate": 0,
    "connections": 120
  }
}
```

### Mock Context Variable (`mock_context` for Billing, infra_agent)

```json
{
  "compute": {
    "ec2_instances": 5,
    "eks_nodes_yesterday": 5,
    "eks_nodes_today": 20
  },
  "kubernetes": {
    "payment_api_replicas": 40,
    "payment_api_replicas_baseline": 5,
    "hpa_events": [
      {"timestamp": "2026-06-14T02:00:00Z", "action": "scale_up", "from": 5, "to": 40}
    ]
  },
  "database": {
    "aurora_instance_class": "db.r6.4xlarge",
    "aurora_instance_class_yesterday": "db.r6.large",
    "aurora_storage_gb": 500,
    "cost_contribution_pct": 65
  },
  "storage": {
    "s3_daily_increase_gb": 45,
    "ebs_daily_increase_gb": 12
  },
  "cost_summary": {
    "expected_daily_usd": 300,
    "current_daily_usd": 1200,
    "delta_usd": 900
  }
}
```

---

## 4.4 Application Support Agent (`prompts/application.py`)

### System Prompt

```text
You are the Application Support AI.
Your role is to investigate application-level issues using logs, traces, APM, and service dependencies.
- Analyze error patterns, exceptions, timeouts
- Identify latency bottlenecks via traces
- Map service dependencies (upstream/downstream)
- Suggest application-level remediation

Rules:
- Confidence score 0-100
- Provide root_cause_hypothesis if clear
- Evidence: list error patterns, trace IDs, log signatures

Output MUST be valid JSON with keys: finding, confidence, suggested_action, evidence, root_cause_hypothesis.
```

### User Prompt Template (Outage)

```text
Service: {service}
Incident Type: {incident_type}
Severity: {severity}

Observability Context (Application Layer):
{mock_context}

What is the application-level root cause? Output JSON.
```

### Mock Context (Outage, app_agent)

```json
{
  "logs": {
    "error_count_last_10m": 4500,
    "error_types": [
      {"type": "connection_timeout", "count": 3800},
      {"type": "context_deadline_exceeded", "count": 600},
      {"type": "pq: connection refused", "count": 100}
    ],
    "stack_trace_signature": "database/sql.(*DB).conn() -> pq.DialTimeout() -> net.Dialer.DialContext()",
    "log_volume_increase_pct": 300
  },
  "traces": {
    "p99_latency_ms": 2100,
    "p99_baseline_ms": 85,
    "slowest_endpoint": "/v1/payments/process",
    "trace_error_rate_pct": 34.5,
    "failing_span": "db_query: SELECT * FROM transactions WHERE ...",
    "avg_query_time_ms": 1800
  },
  "apm": {
    "service": "payment-api",
    "request_rate_rps": 450,
    "error_rate_pct": 34.5,
    "throughput_change_pct": -25
  },
  "dependencies": {
    "downstream": ["postgres-payment", "redis-cache", "payment-webhook"],
    "upstream": ["api-gateway", "web-frontend"],
    "affected_downstream": ["postgres-payment"]
  }
}
```

### Mock Context (Billing, app_agent)

```json
{
  "logs": {
    "retry_count_total": 8000000,
    "retry_count_baseline": 500000,
    "retry_pattern": "Exponential backoff infinite retry: max retries exceeded, reconnecting...",
    "error_loop_signature": "payment retry: attempt 50/∞, backing off 30s",
    "top_error_message": "connection refused to webhook processor"
  },
  "traces": {
    "expensive_endpoint": "/v1/payments/retry",
    "request_count_spike_pct": 900,
    "trace_hot_path": "payment-api → payment-retry-service → webhook-processor → webhook-processor",
    "runaway_request_indicator": true
  },
  "apm": {
    "service": "payment-api",
    "request_rate_rps": 12000,
    "request_rate_baseline_rps": 1200,
    "error_rate_pct": 18.5,
    "throughput_change_pct": 900
  }
}
```

---

## 4.5 Change Correlation Agent (`prompts/change_correlation.py`)

### System Prompt

```text
You are the Change Correlation AI.
Your role is to correlate the incident timing with recent changes: deployments, config changes, feature flags, infrastructure modifications.
Calculate a probability score for each change category.

Rules:
- Change probability 0.0-100.0 for each category
- Time window considered: 24 hours before incident
- Always include the most recent change event with timestamp

Output MUST be valid JSON with keys:
  findings: [{category, probability, description, timestamp}],
  highest_probability_category,
  overall_confidence.
```

### User Prompt Template

```text
Service: {service}
Incident Time: {incident_time}
Incident Type: {incident_type}

Observability Context (Change Events):
{mock_context}

What change is most likely causing this incident? Output JSON.
```

### Mock Context (Outage, change_agent)

```json
{
  "deployment_events": [
    {"service": "payment-api", "version": "v2.1.5", "timestamp": "2026-06-14T09:52:00Z", "deployed_by": "argo-cd", "git_commit": "a1b2c3d"}
  ],
  "config_changes": [
    {"service": "payment-api", "key": "DB_POOL_SIZE", "old_value": "50", "new_value": "10", "timestamp": "2026-06-14T09:50:00Z"}
  ],
  "infrastructure_changes": [],
  "feature_flags": [],
  "incident_start_time": "2026-06-14T10:00:00Z"
}
```

### Mock Context (Billing, change_agent)

```json
{
  "deployment_events": [],
  "config_changes": [
    {"service": "payment-api", "key": "HPA_MAX_REPLICAS", "old_value": "10", "new_value": "50", "timestamp": "2026-06-13T22:00:00Z"},
    {"service": "payment-api", "key": "RETRY_MAX_ATTEMPTS", "old_value": "3", "new_value": "-1", "timestamp": "2026-06-14T01:00:00Z"}
  ],
  "infrastructure_changes": [
    {"resource": "aurora-postgres", "change": "instance_class_modified", "old": "db.r6.large", "new": "db.r6.4xlarge", "timestamp": "2026-06-13T20:00:00Z"}
  ],
  "feature_flags": [],
  "incident_start_time": "2026-06-14T10:00:00Z"
}
```

---

## 4.6 Business Impact Agent (`prompts/business_impact.py`)

### System Prompt

```text
You are the Business Impact AI.
Your role is to translate technical findings into business impact: affected users, failed transactions, revenue exposure, SLA risk.

Rules:
- Revenue exposure: provide currency per hour or per day
- Affected users: estimate based on traffic patterns
- SLA risk: HIGH if availability < 99.9%, CRITICAL if < 99.0%
- Always project monthly impact for billing incidents

Output MUST be valid JSON with keys: affected_users, failed_transactions, revenue_exposure, sla_risk, business_impact_level, projected_monthly_loss, confidence.
```

### User Prompt Template (Outage)

```text
Service: {service}
Incident Type: {incident_type}
Severity: {severity}

Technical Findings Summary:
{findings_summary}

Calculate business impact. Output JSON.
```

### Mock Context (Outage, business_agent)

```json
{
  "traffic": {
    "daily_active_users": 450000,
    "request_rate_rps": 450,
    "error_rate_pct": 34.5,
    "avg_revenue_per_transaction_usd": 1.75,
    "sla_target_pct": 99.95,
    "current_availability_pct": 65.5
  },
  "incident_duration_estimate_minutes": 30
}
```

### Mock Context (Billing, business_agent)

```json
{
  "cost_data": {
    "expected_daily_usd": 300,
    "current_daily_usd": 1200,
    "delta_usd": 900,
    "primary_contributor": "aurora-postgres",
    "contribution_pct": 65
  },
  "historical_comparison": {
    "weekly_avg_usd": 310,
    "monthly_avg_usd": 330
  }
}
```

---

## 4.7 Decision Engine (`prompts/decision_engine.py`)

### System Prompt

```text
You are the Decision Engine AI.
Your role is to aggregate all agent findings, synthesize a root cause, generate a recommended action, assess risk, and determine if human approval is required.

You are the final decision maker before human approval.

Rules:
1. Root cause must synthesize ALL agent findings, not just one
2. Recommended action must be concrete and executable
3. Confidence score 0-100 based on consistency of findings
4. Risk assessment: LOW / MEDIUM / HIGH / CRITICAL
5. Approval required if action modifies production state (deployment rollback, scaling, DB failover)
6. Approval NOT required for read-only actions or cache purges

Output MUST be valid JSON with keys:
  root_cause, contributing_factors, recommended_action, confidence, risk, approval_required, affected_users, revenue_exposure, recovery_time_estimate.
```

### User Prompt Template

```text
Incident: {incident_id}
Service: {service}
Type: {incident_type}
Severity: {severity}

Agent Findings (collected from parallel investigation):

[INFRASTRUCTURE AGENT]
{infra_finding}

[APPLICATION AGENT]
{app_finding}

[CHANGE CORRELATION AGENT]
{change_finding}

[BUSINESS IMPACT AGENT]
{business_finding}

Synthesize findings. Determine root cause, recommendation, and risk. Output JSON.
```

---

## 4.8 Prompt File Mapping Table

| Incident Type | infra_agent | app_agent | change_agent | business_agent | decision_engine |
|--------------|-------------|-----------|--------------|----------------|-----------------|
| outage | `infrastructure.py` (outage ctx) | `application.py` (outage ctx) | `change_correlation.py` (outage ctx) | `business_impact.py` (outage ctx) | `decision_engine.py` (all findings) |
| billing | `infrastructure.py` (billing ctx) | `application.py` (billing ctx) | `change_correlation.py` (billing ctx) | `business_impact.py` (billing ctx) | `decision_engine.py` (all findings) |
| database | `infrastructure.py` (database ctx - to be created) | `application.py` (database ctx - to be created) | `change_correlation.py` | `business_impact.py` | `decision_engine.py` |
| slo_breach | `infrastructure.py` | `application.py` | `change_correlation.py` | `business_impact.py` (SLO focus) | `decision_engine.py` |
| custom | `infrastructure.py` (generic ctx) | `application.py` (generic ctx) | `change_correlation.py` | `business_impact.py` (generic) | `decision_engine.py` |

> **Note:** Untuk demo, hanya `outage` dan `billing` yang memiliki mock_context lengkap. Tipe lain dapat fallback ke generic context. Hal ini menunjukkan **flexibility** engine tanpa hardcoded per use case.

---

## 4.9 JSON Output Schema Validation

Setiap agent output harus LOL valid conform kepada schema berikut:

```python
AGENT_OUTPUT_SCHEMA = {
    "required": ["finding", "confidence"],
    "properties": {
        "finding": {"type": "string", "minLength": 20},
        "confidence": {"type": "integer", "minimum": 0, "maximum": 100},
        "suggested_action": {"type": "string"},
        "evidence": {"type": "array", "items": {"type": "string"}}
    }
}

# Decision Engine extended schema
DECISION_OUTPUT_SCHEMA = {
    "required": ["root_cause", "recommended_action", "confidence", "risk", "approval_required"],
    "properties": {
        "root_cause": {"type": "string", "minLength": 20},
        "contributing_factors": {"type": "array", "items": {"type": "string"}},
        "recommended_action": {"type": "string"},
        "confidence": {"type": "integer", "minimum": 0, "maximum": 100},
        "risk": {"type": "string", "enum": ["LOW", "MEDIUM", "HIGH", "CRITICAL"]},
        "approval_required": {"type": "boolean"},
        "affected_users": {"type": "integer"},
        "revenue_exposure": {"type": "string"},
        "recovery_time_estimate": {"type": "string"}
    }
}
```

---

> Next: baca `05-database-schema.md` untuk DDL detail.
