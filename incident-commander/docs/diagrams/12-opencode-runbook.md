# 12 Runbook: Menggunakan Opencode untuk Generate Code

> Step-by-step runbook untuk mengeksekusi implementasi menggunakan opencode sebagai code generator.

---

## 12.1 Prerequisites

Sebelum mulai:
1. Pastikan Docker & Docker Compose terinstall
2. Pastikan `.env` file terisi dengan:
   - `OPENROUTER_API_KEY`
   - `DD_API_KEY`
   - `SLACK_WEBHOOK_URL`
3. Pastikan opencode running dan bisa access working directory

---

## 12.2 Phase 1: Infrastructure & Base (Est. 30-45 menit)

### Step 1.1: Generate Docker Files

**Prompt untuk opencode:**

```
Generate docker-compose.yml, Dockerfile, requirements.txt, and .env.template for AI Incident Commander:

Requirements:
- docker-compose.yml with 3 services: postgres (5432), datadog-agent (DogStatsD UDP 8125), and app (8000)
- Dockerfile: python:3.11-slim, multi-stage build with gcc, libpq-dev, curl
- requirements.txt: fastapi, uvicorn, sqlalchemy, alembic, psycopg2-binary, pydantic, pydantic-settings, httpx, langgraph, langchain-core, datadog, python-json-logger, pytest, pytest-asyncio
- .env.template: OPENROUTER_API_KEY, LLM_MODEL, POSTGRES_USER/PASSWORD/DB, SLACK_WEBHOOK_URL, DD_API_KEY, DD_ENV

Refer to file: /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/06-docker-compose.md for spec details.
```

### Step 1.2: Generate Database Models

**Prompt untuk opencode:**

```
Generate SQLAlchemy ORM models for PostgreSQL:

File structure: app/models/base.py, app/models/incident.py, app/models/finding.py, app/models/timeline.py, app/models/recommendation.py, app/models/approval.py, app/models/rca.py

Refer to file: /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/05-database-schema.md for full DDL and ERD.

Write app/db.py for SQLAlchemy engine, SessionLocal, and get_db().
Write app/config.py as Pydantic Settings with env vars.
```

---

## 12.3 Phase 2: Agents & LLM (Est. 45-60 menit)

### Step 2.1: Generate Agent Base + LLM Client

**Prompt untuk opencode:**

```
Generate app/agents/base.py with:
- BaseAgent class abstract with execute() method
- LLMClient class using httpx to call OpenRouter API (https://openrouter.ai/api/v1/chat/completions)
- call() method: send system_prompt + user_prompt → return JSON dict with keys: success, data, tokens_prompt, tokens_completion, cost_usd, latency_ms
- call_with_retry() method: retry up to 2 times on INVALID_JSON
- Error handling: Timeout, JSONDecodeError, HTTPStatusError

Refer to file: /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/09-llm-integration.md
```

### Step 2.2: Generate Agent Prompts

**Prompt untuk opencode:**

```
Generate app/agents/prompts/ directory with these files:
- infrastructure.py: System prompt + user prompt template + mock_context dicts for outage and billing
- application.py: System prompt + user prompt template + mock_context dicts for outage and billing  
- change_correlation.py: System prompt + user prompt template + mock_context dicts
- business_impact.py: System prompt + user prompt template + mock_context dicts
- decision_engine.py: System prompt + user prompt template for aggregate all findings
- incident_lead.py: System prompt + user prompt for classification

Each prompt must output valid JSON. Return format defined in each prompt.

Refer to file: /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/04-agent-prompts.md for exact prompts and mock_contexts.
```

### Step 2.3: Generate Agent Implementations

**Prompt untuk opencode:**

```
Generate agent implementations:
- app/agents/infrastructure_agent.py: extends BaseAgent, uses prompts/infrastructure.py mock_context
- app/agents/application_agent.py: same pattern for application
- app/agents/change_correlation_agent.py
- app/agents/business_impact_agent.py
- app/agents/decision_engine.py: takes all findings, generates final recommendation
- app/agents/incident_lead.py: classification and orchestration

Pattern: each agent receives incident_id, service, type, severity → calls LLM with appropriate prompt → parses JSON → saves finding to DB via finding_service.
```

---

## 12.4 Phase 3: Orchestrator (Est. 30-45 menit)

### Step 3.1: Generate LangGraph State Machine

**Prompt untuk opencode:**

```
Generate LangGraph orchestrator in app/orchestrator/:
- state.py: IncidentState TypedDict with fields: incident_id, service, type, severity, status, findings[], recommendation, approval_required, approved, created_at
- nodes.py: node functions (incident_start, infrastructure_node, application_node, change_correlation_node, business_impact_node, decision_engine_node, approval_gate_node, slack_notify_node)
- router.py: conditional edges (if approval_required → wait, else → resolve)
- graph.py: build_graph() returns compiled StateGraph with parallel fan-out and conditional routing

LangGraph version: 0.0.65
Parallel execution: infrastructure + application + change + business agents run in parallel
Conditional routing: decision_engine → approval_gate → (if approved) → slack_notify

Refer to file: /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/01-architecture.md and 02-data-flow.md for architecture details.
```

---

## 12.5 Phase 4: API & Services (Est. 30-45 menit)

### Step 4.1: Generate Services & Routers

**Prompt untuk opencode:**

```
Generate app/services/ and app/routers/:

Services (CRUD logic):
- incident_service.py: create_incident, get_incident, list_active_incidents, resolve_incident
- finding_service.py: create_finding, get_findings_by_incident
- recommendation_service.py: create_recommendation, get_recommendation
- timeline_service.py: add_timeline_event, get_timeline
- approval_service.py: create_approval_request, submit_approval

Routers (FastAPI):
- routers/incident.py: POST /incident/trigger, GET /incident/{id}, GET /incident/{id}/timeline, GET /incidents/active, POST /incident/{id}/resolve
- routers/approval.py: POST /incident/{id}/approve, GET /incident/{id}/approval-status

main.py: FastAPI app instance, mount routers, health check endpoint

Refer to file: /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/07-api-contracts.md for full API spec.
```

---

## 12.6 Phase 5: Integrations (Est. 20-30 menit)

### Step 5.1: Generate Slack & Datadog Clients

**Prompt untuk opencode:**

```
Generate integrations:
- app/integrations/slack_client.py: SlackWebhookClient class using httpx POST to SLACK_WEBHOOK_URL. Methods: post_message, post_incident_header, post_findings, post_decision. Format messages using templates from diagrams/11-slack-format.md

- app/integrations/datadog_client.py: DatadogMetricsClient using datadog.DogStatsd. Methods: incident_created, incident_closed, agent_executed, agent_error, llm_call, llm_error, business_impact, cost_anomaly. Also emit structured JSON logs via python logging.

Refer to files: 
- /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/08-datadog-metrics.md
- /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/11-slack-format.md
```

---

## 12.7 Phase 6: Scripts & Test (Est. 15-20 menit)

### Step 6.1: Generate Demo Scripts

**Prompt untuk opencode:**

```
Generate scripts/ directory:
- trigger_outage.sh: curl POST to /incident/trigger with outage payload
- trigger_billing.sh: curl POST with billing payload
- check_incident.sh: curl GET /incident/{id}
- approve_action.sh: curl POST /incident/{id}/approve
- run_demo.sh: automated script that triggers outage, waits 40s, checks status
```

### Step 6.2: Generate Tests (if time permits)

**Prompt untuk opencode:**

```
Generate tests/:
- conftest.py: pytest fixtures (db_session, test_client, mock_llm)
- test_trigger.py: test POST /incident/trigger creates incident with correct fields
- test_orchestrator.py: test LangGraph runs end-to-end with mocked LLM
```

---

## 12.8 Phase 7: Run & Debug (Est. 30-60 menit)

### Step 7.1: Build & Compose

```bash
# Terminal 1
docker compose up --build

# Terminal 2 (wait for healthy)
docker compose ps
curl http://localhost:8000/health
```

### Step 7.2: Trigger Scenarios

```bash
# Terminal 2
chmod +x scripts/*.sh
./scripts/trigger_outage.sh
sleep 45
./scripts/check_incident.sh INC-2026-001
```

### Step 7.3: Verify Outputs

1. **Postgres:** `docker compose exec db psql -U postgres -d incident_commander -c "SELECT * FROM incidents;"`
2. **Datadog:** Check DD dashboard for `aic.incident.created` metrics
3. **Slack:** Check Slack channel for unified message

---

## 12.9 Opencode Prompt Best Practices

### Prompt Structure yang Efektif

```
Context: [Referensi ke file spec]
Goal: [Apa yang mau dibuat]
Requirements: [Acceptance criteria]
Constraints: [Limitations: framework version, library, dll]
Output: [File path dan expected content]
```

### Contoh Prompt untuk LangGraph Orchestrator

```
Generate app/orchestrator/graph.py that builds a LangGraph state machine for AI Incident Commander.

Context: Refer to /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/01-architecture.md

Requirements:
1. Use LangGraph 0.0.65
2. State: IncidentState (TypedDict) with fields incident_id, service, type, severity, status, findings[], recommendation, approval_required, approved
3. Nodes: incident_start → parallel(infra_node, app_node, change_node, business_node) → decision_node → conditional_router → slack_node
4. Parallel execution using langgraph fan-out
5. Conditional routing: if approval_required → END (wait), else → slack_notify
6. Each node runs in async context
7. State updates use typed dict updates

Write graph.py and nodes.py in app/orchestrator/.
```

---

## 12.10 Troubleshooting Quick Guide

| Problem | Check | Fix |
|---------|-------|-----|
| Postgres connect refused | Is db container healthy? | `docker compose up -d db && wait 10s` |
| OpenRouter timeout | Is API key valid? | Check OpenRouter dashboard, rate limit |
| Slack not received | Is webhook URL correct? | Test via curl manual |
| Datadog metrics missing | Is datadog-agent container running? | Check `DD_API_KEY`, container logs |
| LLM returns invalid JSON | Is temperature too high? | Set temp to 0.1, add validation |
| LangGraph hangs | Are all nodes returning state? | Check all nodes return dict |

---

## 12.11 File Priority for Demo

**Must have (MVP):**
- [ ] docker-compose.yml, Dockerfile, requirements.txt
- [ ] app/config.py, app/db.py
- [ ] app/models/*.py
- [ ] app/agents/base.py, app/agents/prompts/*.py, app/agents/*_agent.py
- [ ] app/orchestrator/*.py
- [ ] app/services/*.py
- [ ] app/routers/*.py, app/main.py
- [ ] app/integrations/slack_client.py, app/integrations/datadog_client.py
- [ ] scripts/*.sh

**Nice to have:**
- [ ] app/integrations/datadog_api.py (real Datadog API queries)
- [ ] app/orchestrator/loopback.py (human approval loopback)
- [ ] Knowledge base (Qdrant)
- [ ] RCA auto-generator
- [ ] Cached LLM responses (Redis)

---

## 12.12 Time Allocation Strategy

| Phase | Time | Priority |
|-------|------|----------|
| 1. Infrastructure | 30-45 min | Critical |
| 2. Agents + LLM | 45-60 min | Critical |
| 3. Orchestrator | 30-45 min | Critical |
| 4. API + Services | 30-45 min | Critical |
| 5. Integrations | 20-30 min | High |
| 6. Scripts | 15-20 min | High |
| 7. Run & Debug | 30-60 min | Critical |
| **Buffer** | **15-30 min** | — |
| **Total** | **~4 jam** | — |

---

## 12.13 Success Criteria

Demo berhasil jika:
1. ✅ Trigger outage → incident_id terbentuk
2. ✅ 4 agent berjalan paralel (log terlihat)
3. ✅ Decision engine menghasilkan recommendation
4. ✅ Postgres berisi incident + findings + recommendations
5. ✅ Slack terima unified message
6. ✅ Datadog metrics `aic.incident.created` muncul di dashboard
7. ✅ Trigger billing → proses sama, context berbeda (proving flexibility)

---

> This concludes the implementation runbook. Good luck with the 4-hour demo!
