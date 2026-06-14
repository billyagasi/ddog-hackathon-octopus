# 03 Complete File Structure Blueprint

> Blueprint lengkap semua file yang harus di-generate/di-implementasikan oleh opencode.

---

## Directory Tree (Full)

```text
ai-incident-commander/
│
├── .env                                 # Environment variables (copy dari template)
├── .env.template                        # Template untuk .env
├── docker-compose.yml                   # Docker Compose: app, postgres, datadog-agent
├── Dockerfile                           # Dockerfile untuk FastAPI app
├── requirements.txt                     # Python dependencies
├── README.md                            # Cara run demo
│
├── scripts/
│   ├── trigger_outage.sh                # curl untuk trigger scenario 1
│   ├── trigger_billing.sh               # curl untuk trigger scenario 2
│   ├── check_incident.sh                # curl GET /incident/{id}
│   ├── approve_action.sh                # curl POST /incident/{id}/approve
│   └── run_demo.sh                      # Script otomatis: trigger → sleep → check timeline
│
├── app/
│   ├── __init__.py
│   ├── main.py                          # FastAPI app initialization, routers mount
│   ├── config.py                        # Pydantic Settings, env var loader
│   ├── dependencies.py                  # FastAPI dependency injection (DB session, etc)
│   │
│   ├── models/
│   │   ├── __init__.py
│   │   ├── base.py                      # SQLAlchemy declarative base
│   │   ├── incident.py                  # Tabel: incidents
│   │   ├── finding.py                   # Tabel: findings
│   │   ├── timeline.py                  # Tabel: timeline
│   │   ├── recommendation.py            # Tabel: recommendations
│   │   ├── approval.py                  # Tabel: approvals
│   │   └── rca.py                       # Tabel: rca (post-resolution)
│   │
│   ├── schemas/
│   │   ├── __init__.py
│   │   ├── incident.py                  # Pydantic models: IncidentCreate, IncidentOut
│   │   ├── finding.py                   # Pydantic models: FindingOut
│   │   ├── recommendation.py            # Pydantic models: RecommendationOut
│   │   ├── timeline.py                  # Pydantic models: TimelineEvent
│   │   ├── approval.py                  # Pydantic models: ApprovalRequest
│   │   ├── trigger.py                   # Pydantic model: TriggerPayload
│   │   └── decision.py                  # Pydantic model: DecisionOut
│   │
│   ├── db.py                            # SQLAlchemy engine, sessionmaker, get_db()
│   │
│   ├── services/
│   │   ├── __init__.py
│   │   ├── incident_service.py          # CRUD logic untuk incidents
│   │   ├── finding_service.py           # CRUD logic untuk findings
│   │   ├── recommendation_service.py    # CRUD logic untuk recommendations
│   │   ├── timeline_service.py          # CRUD logic untuk timeline
│   │   └── approval_service.py          # CRUD logic untuk approvals
│   │
│   ├── agents/
│   │   ├── __init__.py
│   │   ├── base.py                      # BaseAgent class: LLM call, JSON parsing, error handling
│   │   ├── prompts/
│   │   │   ├── __init__.py
│   │   │   ├── infrastructure.py        # System prompt + user prompt template infra agent
│   │   │   ├── application.py           # System prompt + user prompt template app agent
│   │   │   ├── change_correlation.py    # System prompt + user prompt template change agent
│   │   │   ├── business_impact.py       # System prompt + user prompt template business agent
│   │   │   ├── decision_engine.py       # System prompt + user prompt template decision engine
│   │   │   └── incident_lead.py         # System prompt + user prompt template incident lead
│   │   │
│   │   ├── infrastructure_agent.py      # InfraAgent extends BaseAgent
│   │   ├── application_agent.py         # AppAgent extends BaseAgent
│   │   ├── change_correlation_agent.py  # ChangeAgent extends BaseAgent
│   │   ├── business_impact_agent.py     # BusinessAgent extends BaseAgent
│   │   ├── decision_engine.py          # DecisionEngine extends BaseAgent
│   │   └── incident_lead.py              # IncidentLead orchestrates, extends BaseAgent
│   │
│   ├── orchestrator/
│   │   ├── __init__.py
│   │   ├── state.py                     # IncidentState TypedDict (LangGraph state)
│   │   ├── graph.py                     # graph builder: nodes + edges + conditional routing
│   │   ├── nodes.py                     # Individual node functions (incident_start, infra_node, app_node, dll)
│   │   └── router.py                    # Conditional edge: approval_required? → wait / skip
│   │
│   ├── integrations/
│   │   ├── __init__.py
│   │   ├── slack_client.py              # SlackWebhookClient: post_message, post_thread, format blocks
│   │   └── datadog_client.py            # DatadogMetricsClient: gauge, count, timer, histogram
│   │
│   └── routers/
│       ├── __init__.py
│       ├── incident.py                  # FastAPI routes: POST /trigger, GET /{id}, GET /{id}/timeline
│       └── approval.py                  # FastAPI routes: POST /{id}/approve, GET /{id}/approval-status
│
└── tests/
    ├── __init__.py
    ├── conftest.py                      # Pytest fixtures: db_session, test_client, mock_llm
    ├── test_trigger.py                  # Test: trigger POST creates incident + findings
    ├── test_orchestrator.py             # Test: LangGraph state machine runs end-to-end
    ├── test_agents.py                   # Test: masing-masing agent returns valid JSON
    └── test_datadog.py                  # Test: metrics emitted correctly
```

---

## File Descriptions & Opencode Prompts

> Untuk setiap file di bawah ini, buatkan prompt generasi yang spesifik untuk opencode.

### 1. `app/config.py`

**Purpose:** Pydantic Settings untuk load semua env var.

**Key fields:**
- `OPENROUTER_API_KEY: str`
- `LLM_MODEL: str = "anthropic/claude-sonnet-4-20250514"`
- `DATABASE_URL: str`
- `SLACK_WEBHOOK_URL: str`
- `DD_AGENT_HOST: str = "datadog-agent"`
- `DD_DOGSTATSD_PORT: int = 8125`

---

### 2. `app/models/*.py`

**Purpose:** SQLAlchemy ORM models per tabel.

**Key tables:**

```python
# models/incident.py
class Incident(Base):
    __tablename__ = "incidents"
    id = Column(String(50), primary_key=True)  # INC-YYYY-NNN
    incident_type = Column(String(50))         # outage|billing|custom
    service = Column(String(100))
    severity = Column(String(10))              # SEV1|SEV2|SEV3|SEV4
    status = Column(String(50))                # investigating|pending_approval|approved|resolved|closed
    datadog_url = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    closed_at = Column(DateTime, nullable=True)
    # relationships: findings, recommendations, timeline, approvals, rca

# models/finding.py  
class Finding(Base):
    __tablename__ = "findings"
    id = Column(Integer, primary_key=True, autoincrement=True)
    incident_id = Column(String(50), ForeignKey("incidents.id"), index=True)
    agent_name = Column(String(50))            # infrastructure|application|change|business
    finding = Column(Text)
    confidence = Column(Integer)               # 0-100
    suggested_action = Column(Text, nullable=True)
    evidence = Column(JSON, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

# models/timeline.py
class TimelineEvent(Base):
    __tablename__ = "timeline"
    id = Column(Integer, primary_key=True, autoincrement=True)
    incident_id = Column(String(50), ForeignKey("incidents.id"), index=True)
    actor = Column(String(50))                 # system|infrastructure_agent|application_agent|human
    event = Column(Text)
    timestamp = Column(DateTime, default=datetime.utcnow)

# models/recommendation.py
class Recommendation(Base):
    __tablename__ = "recommendations"
    id = Column(Integer, primary_key=True, autoincrement=True)
    incident_id = Column(String(50), ForeignKey("incidents.id"), index=True)
    recommendation = Column(Text)
    root_cause = Column(Text)
    confidence = Column(Integer)
    risk = Column(String(20))                  # LOW|MEDIUM|HIGH|CRITICAL
    approval_required = Column(Boolean, default=False)
    affected_users = Column(Integer, nullable=True)
    revenue_exposure = Column(String(100), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

# models/approval.py
class Approval(Base):
    __tablename__ = "approvals"
    id = Column(Integer, primary_key=True, autoincrement=True)
    incident_id = Column(String(50), ForeignKey("incidents.id"), index=True)
    approver = Column(String(100), nullable=True)
    status = Column(String(20))                # pending|approved|rejected
    requested_at = Column(DateTime, default=datetime.utcnow)
    responded_at = Column(DateTime, nullable=True)

# models/rca.py
class RCA(Base):
    __tablename__ = "rca"
    incident_id = Column(String(50), ForeignKey("incidents.id"), primary_key=True)
    summary = Column(Text)
    root_cause = Column(Text)
    resolution = Column(Text)
    lessons_learned = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
```

---

### 3. `app/agents/base.py`

**Purpose:** Abstract base class semua agent.

```python
class BaseAgent:
    def __init__(self, name: str, api_key: str, model: str, db: Session):
        self.name = name
        self.api_key = api_key
        self.model = model
        self.db = db
    
    async def call_llm(self, system_prompt: str, user_prompt: str, mock_context: dict) -> dict:
        """Call OpenRouter API, parse JSON, return dict."""
        pass
    
    async def execute(self, incident_id: str, service: str, incident_type: str, severity: str) -> dict:
        """Main entry point."""
        pass
    
    async def save_finding(self, incident_id: str, result: dict):
        """Save finding ke DB."""
        pass
```

---

### 4. `app/agents/prompts/infrastructure.py`

**Purpose:** System prompt + user prompt template untuk Infrastructure Agent.

**Key requirement:** Prompt harus menghasilkan JSON valid dengan keys: `finding`, `confidence`, `suggested_action`, `evidence`.

**Mock context injection:** Prompt memiliki placeholder `{mock_context}` yang diisi oleh orchestrator.

---

### 5. `app/orchestrator/state.py`

**Purpose:** LangGraph TypedDict untuk state machine.

```python
class IncidentState(TypedDict):
    incident_id: str
    service: str
    incident_type: str
    severity: str
    status: str
    datadog_url: Optional[str]
    findings: List[dict]
    recommendation: Optional[dict]
    approval_required: bool
    approved: bool
    created_at: str
    slack_thread_ts: Optional[str]
```

---

### 6. `app/orchestrator/graph.py`

**Purpose:** Build LangGraph state machine DAG.

**Key functions:**
- `build_graph() -> StateGraph`
- `add_nodes()`
- `add_edges()`
- `add_conditional_edge()` untuk approval gate

---

### 7. `app/integrations/slack_client.py`

**Purpose:** Simple slack webhook client.

```python
class SlackWebhookClient:
    def __init__(self, webhook_url: str):
        self.webhook_url = webhook_url
    
    async def post_message(self, text: str, thread_ts: Optional[str] = None) -> dict:
        """POST ke Slack Webhook URL."""
        pass
    
    async def post_incident_header(self, incident: dict) -> str:
        """Post header, return thread timestamp."""
        pass
    
    async def post_finding(self, thread_ts: str, finding: dict) -> None:
        """Reply finding ke thread."""
        pass
    
    async def post_decision(self, thread_ts: str, decision: dict) -> None:
        """Post final decision."""
        pass
```

> For demo: thread dapat dibuat via **unified message** (semua findings + decision dalam 1 POST), karena Incoming Webhook tidak native support thread_ts tanpa Slack App + Bot Token.

---

### 8. `app/integrations/datadog_client.py`

**Purpose:** Emit metrics via DogStatsD.

```python
class DatadogMetricsClient:
    def __init__(self, host: str, port: int):
        self.statsd = DogStatsd(host=host, port=port)
    
    def incident_created(self, incident_type: str, severity: str):
        self.statsd.increment("aic.incident.created", tags=[f"type:{incident_type}", f"severity:{severity}"])
    
    def agent_executed(self, agent_name: str, duration_ms: float, confidence: int):
        self.statsd.timing("aic.agent.execution.duration", duration_ms, tags=[f"agent:{agent_name}"])
        self.statsd.gauge("aic.agent.confidence", confidence, tags=[f"agent:{agent_name}"])
    
    def recommendation_generated(self, confidence: int, risk: str):
        self.statsd.gauge("aic.recommendation.confidence", confidence, tags=[f"risk:{risk}"])
        self.statsd.increment("aic.incident.approval.required", tags=[f"risk:{risk}"])
    
    def incident_closed(self, duration_seconds: int):
        self.statsd.histogram("aic.incident.duration", duration_seconds)
```

> Note: untuk demo lokal, DogStatsD agent harus di Docker Compose bersama app.

---

### 9. `app/routers/incident.py`

**Routes:**

```python
@router.post("/incident/trigger")
async def trigger_incident(payload: TriggerPayload, db: Session = Depends(get_db)):
    """Create incident, start LangGraph orchestration."""
    pass

@router.get("/incident/{incident_id}")
async def get_incident(incident_id: str, db: Session = Depends(get_db)):
    """Return full incident with findings and timeline."""
    pass

@router.get("/incident/{incident_id}/timeline")
async def get_timeline(incident_id: str, db: Session = Depends(get_db)):
    pass
```

---

### 10. `app/routers/approval.py`

**Routes:**

```python
@router.post("/incident/{incident_id}/approve")
async def approve_action(incident_id: str, request: ApprovalRequest, db: Session = Depends(get_db)):
    """Update approval status."""
    pass

@router.get("/incident/{incident_id}/approval-status")
async def get_approval_status(incident_id: str, db: Session = Depends(get_db)):
    pass
```

---

### 11. `docker-compose.yml`

**Services:**
- `db`: `postgres:15-alpine`, port 5432, volume untuk persist
- `datadog-agent`: `gcr.io/datadoghq/agent:latest`, mount docker socket, env `DD_API_KEY`, DogStatsD UDP expose
- `app`: Build dari `Dockerfile`, depends_on `db`, `datadog-agent`

---

### 12. `Dockerfile`

**Base:** `python:3.11-slim`
**Steps:**
- WORKDIR /app
- COPY requirements.txt + pip install
- COPY app/
- CMD uvicorn app.main:app --host 0.0.0.0 --port 8000

---

### 13. `requirements.txt`

Key packages:
```
fastapi==0.115.0
uvicorn[standard]==0.30.0
sqlalchemy==2.0.30
alembic==1.13.0
psycopg2-binary==2.9.9
pydantic==2.7.0
pydantic-settings==2.2.0
httpx==0.27.0
langgraph==0.0.65
langchain-core==0.1.0
openpy
python-json-logger==2.0.7
datadog==0.49.1
pytest==8.2.0
pytest-asyncio==0.23.0
```

---

### 14. `scripts/trigger_outage.sh`

```bash
#!/bin/bash
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
    }
  }'
```

### 15. `scripts/trigger_billing.sh`

```bash
#!/bin/bash
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
    }
  }'
```

---

### 16. `README.md`

Sections:
1. Quick Start (docker compose up)
2. Environment Setup (.env dari template)
3. Trigger Scenarios (scripts)
4. Check Results (GET /incident/{id})
5. View Datadog Metrics
6. View Slack Thread

---

## Opencode Generation Priorities

**Prioritas 1 (Hour 1-1.5):**
1. `docker-compose.yml`, `Dockerfile`, `requirements.txt`, `.env.template`
2. `app/config.py`, `app/db.py`
3. `app/models/*.py` (semua tabel)
4. `app/schemas/*.py` (semua Pydantic)

**Prioritas 2 (Hour 1.5-2.5):**
5. `app/agents/base.py`, `app/agents/prompts/*.py`, `app/agents/*_agent.py`
6. `app/orchestrator/state.py`, `app/orchestrator/graph.py`, `app/orchestrator/nodes.py`, `app/orchestrator/router.py`

**Prioritas 3 (Hour 2.5-3.5):**
7. `app/integrations/slack_client.py`, `app/integrations/datadog_client.py`
8. `app/services/*.py`
9. `app/routers/incident.py`, `app/routers/approval.py`
10. `app/main.py`

**Prioritas 4 (Hour 3.5-4):**
11. `scripts/*.sh`
12. `README.md`
13. Test: `trigger_outage.sh` + `trigger_billing.sh` end-to-end

---

> Next: baca `04-agent-prompts.md` untuk semua LLM prompts yang harus dibuat.
