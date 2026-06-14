# 08 Datadog Metrics & Logging Strategy

> Cara meng-emit metrics, logs, dan traces ke Datadog untuk demo observability end-to-end.

---

## 8.1 Datadog Agent Configuration

DogStatsD UDP port `8125` di-expose dari container `datadog-agent`.
App harus knows host dan port via environment variable.

---

## 8.2 Metrics Catalog

### Incident Lifecycle Metrics

| Metric Name | Type | Tags | Description |
|------------|------|------|-------------|
| `aic.incident.created` | count | `type`, `severity`, `service` | Jumlah incident dibuat |
| `aic.incident.closed` | count | `type`, `severity` | Jumlah incident ditutup |
| `aic.incident.duration` | histogram/gauge | `type`, `severity` | Durasi incident dalam detik |
| `aic.incident.approval.required` | count | `severity`, `risk` | Jumlah incident yang butuh approval |
| `aic.incident.approval.time_to_approve` | timer | `severity` | Waktu dari request → approve |
| `aic.incident.automation.executed` | count | `action_type` | Jumlah action dieksekusi |

### Agent Performance Metrics

| Metric Name | Type | Tags | Description |
|------------|------|------|-------------|
| `aic.agent.execution.count` | count | `agent` | Jumlah eksekusi per agent |
| `aic.agent.execution.duration` | timer | `agent`, `incident_type` | Latency LLM call per agent |
| `aic.agent.confidence` | gauge | `agent`, `incident_type` | Confidence score per agent |
| `aic.agent.error` | count | `agent`, `error_type` | Jumlah error LLM |

### LLM Observability Metrics

| Metric Name | Type | Tags | Description |
|------------|------|------|-------------|
| `aic.llm.tokens.prompt` | count | `model`, `agent` | Token count prompt |
| `aic.llm.tokens.completion` | count | `model`, `agent` | Token count completion |
| `aic.llm.tokens.total` | count | `model`, `agent` | Total token |
| `aic.llm.cost.usd` | gauge | `model` | Estimasi cost per call |
| `aic.llm.latency` | timer | `model`, `agent` | Latency call LLM |
| `aic.llm.errors` | count | `model`, `error_code` | LLM API errors |

### Business Impact Metrics

| Metric Name | Type | Tags | Description |
|------------|------|------|-------------|
| `aic.business.affected_users` | gauge | `service`, `incident_type` | Estimasi affected users |
| `aic.business.revenue_exposure` | gauge | `service`, `incident_type` | Revenue exposure USD/hour |
| `aic.slo.breach.detected` | count | `service`, `slo_name` | Jumlah SLO breach |
| `aic.error_budget.burn_rate` | gauge | `service` | Error budget burn rate |

### Cost Metrics

| Metric Name | Type | Tags | Description |
|------------|------|------|-------------|
| `aic.cost.expected_daily` | gauge | `service` | Expected daily cost |
| `aic.cost.actual_daily` | gauge | `service` | Actual daily cost |
| `aic.cost.delta` | gauge | `service`, `resource` | Delta cost USD |
| `aic.cost.projected_monthly` | gauge | `service` | Projected monthly impact |

---

## 8.3 Structured JSON Log Format

Semua log ditulis ke stdout. Datadog Agent scrape dari container log.

```json
{
  "timestamp": "2026-06-14T10:00:15.000Z",
  "service": "ai-incident-commander",
  "env": "demo",
  "level": "INFO",
  "logger_name": "app.agents.infrastructure",
  "message": "Agent investigation completed",
  "incident_id": "INC-2026-001",
  "agent": "infrastructure",
  "incident_type": "outage",
  "severity": "SEV1",
  "finding": "Deployment v2.1.5 correlated with CPU spike and pod crash loops...",
  "confidence": 92,
  "suggested_action": "rollback deployment to v2.1.4",
  "evidence": ["k8s_deployment_event", "cpu_metrics"],
  "execution_duration_ms": 8245,
  "llm_model": "anthropic/claude-sonnet-4-20250514",
  "llm_tokens_prompt": 1250,
  "llm_tokens_completion": 342,
  "llm_cost_usd": 0.0042
}
```

**Log levels:**
- `INFO`: General flow (orchestrator state transitions)
- `DEBUG`: Detail agent reasoning (opsional)
- `WARN`: Skipped agent, timeout, retry
- `ERROR`: LLM call failed, DB error, Slack posting error

---

## 8.4 Python Implementation Sample

```python
# app/integrations/datadog_client.py

from datadog import DogStatsd
import logging
import json
import time
from typing import Optional

logger = logging.getLogger("ai_incident_commander.datadog")

class DatadogMetricsClient:
    def __init__(self, host: str = "datadog-agent", port: int = 8125):
        self.statsd = DogStatsd(host=host, port=port)
    
    # ─── Incident Metrics ─────────────────────
    
    def incident_created(self, incident_type: str, severity: str, service: str):
        tags = [f"type:{incident_type}", f"severity:{severity}", f"service:{service}"]
        self.statsd.increment("aic.incident.created", tags=tags)
        self._audit_log("incident_created", incident_type=incident_type, severity=severity)
    
    def incident_closed(self, incident_id: str, incident_type: str, 
                        severity: str, duration_seconds: int):
        tags = [f"type:{incident_type}", f"severity:{severity}"]
        self.statsd.increment("aic.incident.closed", tags=tags)
        self.statsd.histogram("aic.incident.duration", duration_seconds, tags=tags)
        self._audit_log("incident_closed", incident_id=incident_id, 
                         duration_seconds=duration_seconds)
    
    def approval_required(self, incident_id: str, risk: str, severity: str):
        tags = [f"risk:{risk}", f"severity:{severity}"]
        self.statsd.increment("aic.incident.approval.required", tags=tags)
    
    def approval_resolved(self, incident_id: str, duration_seconds: int, 
                          approved: bool):
        tags = [f"approved:{approved}"]
        self.statsd.histogram("aic.incident.approval.time_to_approve", 
                              duration_seconds, tags=tags)
    
    # ─── Agent Metrics ──────────────────────────
    
    def agent_executed(self, agent_name: str, incident_type: str, 
                       duration_ms: float, confidence: int):
        tags = [f"agent:{agent_name}", f"incident_type:{incident_type}"]
        self.statsd.increment("aic.agent.execution.count", tags=tags)
        self.statsd.timing("aic.agent.execution.duration", duration_ms, tags=tags)
        self.statsd.gauge("aic.agent.confidence", confidence, tags=tags)
    
    def agent_error(self, agent_name: str, error_type: str, 
                    incident_type: str):
        tags = [f"agent:{agent_name}", f"error_type:{error_type}", 
                f"incident_type:{incident_type}"]
        self.statsd.increment("aic.agent.error", tags=tags)
    
    # ─── LLM Metrics ────────────────────────────
    
    def llm_call(self, model: str, agent_name: str, duration_ms: float,
                 tokens_prompt: int, tokens_completion: int, cost_usd: float):
        tags = [f"model:{model}", f"agent:{agent_name}"]
        self.statsd.timing("aic.llm.latency", duration_ms, tags=tags)
        self.statsd.increment("aic.llm.tokens.prompt", tokens_prompt, tags=tags)
        self.statsd.increment("aic.llm.tokens.completion", tokens_completion, tags=tags)
        self.statsd.increment("aic.llm.tokens.total", tokens_prompt + tokens_completion, tags=tags)
        self.statsd.gauge("aic.llm.cost.usd", cost_usd, tags=tags)
    
    def llm_error(self, model: str, error_code: str):
        tags = [f"model:{model}", f"error_code:{error_code}"]
        self.statsd.increment("aic.llm.errors", tags=tags)
    
    # ─── Business Impact Metrics ────────────────
    
    def business_impact(self, service: str, incident_type: str,
                        affected_users: int, revenue_exposure: float):
        tags = [f"service:{service}", f"incident_type:{incident_type}"]
        self.statsd.gauge("aic.business.affected_users", affected_users, tags=tags)
        self.statsd.gauge("aic.business.revenue_exposure", revenue_exposure, tags=tags)
    
    # ─── Cost Metrics ───────────────────────────
    
    def cost_anomaly(self, service: str, resource: str,
                     expected_usd: float, actual_usd: float):
        tags = [f"service:{service}", f"resource:{resource}"]
        self.statsd.gauge("aic.cost.expected_daily", expected_usd, tags=tags)
        self.statsd.gauge("aic.cost.actual_daily", actual_usd, tags=tags)
        self.statsd.gauge("aic.cost.delta", actual_usd - expected_usd, tags=tags)
    
    def cost_projected(self, service: str, projected_monthly_usd: float):
        tags = [f"service:{service}"]
        self.statsd.gauge("aic.cost.projected_monthly", projected_monthly_usd, tags=tags)
    
    # ─── Audit Logger ───────────────────────────
    
    def _audit_log(self, event: str, **kwargs):
        log_entry = {
            "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            "service": "ai-incident-commander",
            "env": "demo",
            "level": "INFO",
            "event": event,
            **kwargs
        }
        logger.info(json.dumps(log_entry))
```

---

## 8.5 Datadog Dashboard Setup (Post-Demo)

Setelah demo berjalan, buat dashboard di Datadog web UI:

### Widget 1: Incident Overview
- Query: `sum:aic.incident.created{*}` (group by `type`, `severity`)
- Type: Timeseries

### Widget 2: Agent Confidence
- Query: `avg:aic.agent.confidence{*}` (group by `agent`)
- Type: Timeseries

### Widget 3: LLM Latency
- Query: `avg:aic.llm.latency{*}` (group by `model`)
- Type: Heatmap

### Widget 4: Incident Duration
- Query: `avg:aic.incident.duration{*}` (group by `severity`)
- Type: Top List

### Widget 5: Revenue Exposure
- Query: `avg:aic.business.revenue_exposure{*}` (group by `service`)
- Type: Timeseries

### Widget 6: Cost Delta
- Query: `avg:aic.cost.delta{*}` (group by `resource`)
- Type: Bar chart

---

## 8.6 Log Indexing & Alerting

**Status Check Log Query (Datadog Log Explorer):**

```
service:ai-incident-commander
@event:incident_created OR @event:incident_closed
```

**Generate Alert:**

```
avg:aic.agent.confidence{service:payment-api}
```
Alert condition: `< 70` (confidence rendah)

---

> Next: baca `09-llm-integration.md` untuk OpenRouter integration detail.
