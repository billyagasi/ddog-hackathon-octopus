# docs/04-observability.md

# Observability Design

AI Incident Commander is fully observable through Datadog.

## LLM Observability

Tracked via Datadog LLM Observability.

### Metrics

aic.llm.tokens.prompt

aic.llm.tokens.completion

aic.llm.tokens.total

aic.llm.cost.usd

aic.llm.latency

aic.llm.errors

Tags:

* model
* incident_type
* agent

---

## Agent Metrics

aic.agent.execution.count

aic.agent.execution.duration

aic.agent.execution.error

aic.agent.confidence

Tags:

* agent
* incident_type

---

## Incident Metrics

aic.incident.created

aic.incident.closed

aic.incident.duration

aic.incident.approval

aic.incident.automation.executed

aic.incident.cost_saved

---

## SLO Metrics

aic.slo.breach.detected

aic.error_budget.remaining

aic.error_budget.burn_rate

---

## Distributed Tracing

Every investigation becomes a distributed trace.

```text
Incident

├── Incident Lead AI
├── Infrastructure AI
├── Application AI
├── Change Correlation AI
├── Business Impact AI
└── Confidence Engine
```

Visualized directly in Datadog APM.

---

## Structured Logs

Example:

```json
{
  "incident_id":"INC-001",
  "agent":"change-correlation",
  "finding":"deployment correlation",
  "confidence":92,
  "source":"datadog-events"
}
```
