# Observability Design

AI Incident Commander is fully observable through Datadog.

---

## LLM Observability

Tracked using Datadog LLM Observability.

### Metrics

aic.llm.tokens.prompt

aic.llm.tokens.completion

aic.llm.tokens.total

aic.llm.cost.usd

aic.llm.latency

aic.llm.error

### Tags

- model
- incident_type
- agent

---

## MCP Metrics

### Metrics

aic.mcp.calls

aic.mcp.duration

aic.mcp.errors

### Tags

- tool
- agent

---

## Agent Metrics

aic.agent.execution.count

aic.agent.execution.duration

aic.agent.execution.error

aic.agent.confidence

### Tags

- agent
- incident_type

---

## Incident Metrics

aic.incident.created

aic.incident.closed

aic.incident.mttr

aic.incident.approval

aic.incident.automation.executed

aic.incident.cost_saved

---

## Reliability Metrics

aic.slo.breach.detected

aic.error_budget.remaining

aic.error_budget.burn_rate

---

## Distributed Tracing

Every investigation becomes a distributed trace.

```text
Incident

├── Incident Lead AI
├── Infrastructure Operations AI
├── Application Intelligence AI
├── Service Management AI
├── Decision Engine
└── Workflow Execution
```

---

## Structured Logs

All findings are logged as structured events.

```json
{
  "incident_id":"INC-001",
  "agent":"application-intelligence",
  "finding":"database timeout",
  "confidence":94,
  "source":"apm-trace"
}
```