# 05 PostgreSQL Database Schema

> DDL lengkap untuk PostgreSQL. Tabel-tabel ini dipakai oleh SQLAlchemy ORM.

---

## 5.1 Schema Overview

```text
                    ┌─────────────┐
                    │  incidents  │
                    │  (master)   │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
    ┌─────────┐     ┌──────────┐     ┌───────────┐
    │findings │     │timeline  │     │recommenda-│
    │         │     │          │     │tions      │
    └─────────┘     └──────────┘     └───────────┘
         │                 │                 │
         ▼                 ▼                 ▼
    ┌─────────┐     ┌──────────┐     ┌───────────┐
    │approvals│     │   rca    │     │knowledge  │
    │         │     │          │     │(Qdrant)   │
    └─────────┘     └──────────┘     └───────────┘
```

---

## 5.2 DDL Statements

### Tabel: `incidents`

```sql
CREATE TABLE IF NOT EXISTS incidents (
    id VARCHAR(50) PRIMARY KEY,
    incident_type VARCHAR(50) NOT NULL,
    service VARCHAR(100) NOT NULL,
    severity VARCHAR(10) NOT NULL CHECK (severity IN ('SEV1','SEV2','SEV3','SEV4')),
    status VARCHAR(50) NOT NULL DEFAULT 'investigating' 
        CHECK (status IN ('investigating','pending_approval','approved','resolved','closed')),
    datadog_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_incidents_type ON incidents(incident_type);
CREATE INDEX idx_incidents_service ON incidents(service);
CREATE INDEX idx_incidents_status ON incidents(status);
CREATE INDEX idx_incidents_created_at ON incidents(created_at);
```

---

### Tabel: `findings`

```sql
CREATE TABLE IF NOT EXISTS findings (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(50) NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    agent_name VARCHAR(50) NOT NULL 
        CHECK (agent_name IN ('infrastructure','application','change_correlation','business_impact','decision_engine')),
    finding TEXT NOT NULL,
    confidence INTEGER NOT NULL CHECK (confidence >= 0 AND confidence <= 100),
    suggested_action TEXT,
    evidence JSONB DEFAULT '[]',
    source VARCHAR(100) DEFAULT 'ai_agent',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_findings_incident ON findings(incident_id);
CREATE INDEX idx_findings_agent ON findings(agent_name);
CREATE INDEX idx_findings_confidence ON findings(confidence);
```

---

### Tabel: `timeline`

```sql
CREATE TABLE IF NOT EXISTS timeline (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(50) NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    actor VARCHAR(50) NOT NULL,
    event TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_timeline_incident ON timeline(incident_id);
CREATE INDEX idx_timeline_timestamp ON timeline(timestamp);
```

---

### Tabel: `recommendations`

```sql
CREATE TABLE IF NOT EXISTS recommendations (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(50) NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    recommendation TEXT NOT NULL,
    root_cause TEXT,
    confidence INTEGER CHECK (confidence >= 0 AND confidence <= 100),
    risk VARCHAR(20) CHECK (risk IN ('LOW','MEDIUM','HIGH','CRITICAL')),
    approval_required BOOLEAN DEFAULT FALSE,
    affected_users INTEGER,
    revenue_exposure VARCHAR(100),
    recovery_time_estimate VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_recommendations_incident ON recommendations(incident_id);
```

---

### Tabel: `approvals`

```sql
CREATE TABLE IF NOT EXISTS approvals (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(50) NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    approver VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending','approved','rejected')),
    action_type VARCHAR(100),
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE,
    reason TEXT
);

CREATE INDEX idx_approvals_incident ON approvals(incident_id);
CREATE INDEX idx_approvals_status ON approvals(status);
```

---

### Tabel: `rca`

```sql
CREATE TABLE IF NOT EXISTS rca (
    incident_id VARCHAR(50) PRIMARY KEY REFERENCES incidents(id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    root_cause TEXT NOT NULL,
    resolution TEXT,
    contributing_factors JSONB DEFAULT '[]',
    lessons_learned TEXT,
    preventive_actions JSONB DEFAULT '[]',
    generated_by VARCHAR(50) DEFAULT 'ai_agent',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rca_incident ON rca(incident_id);
```

---

## 5.3 Entity Relationship Diagram (ERD)

```text
┌────────────────────────────────────────────────────────────────────┐
│ incidents                                                          │
├────────────────────────────────────────────────────────────────────┤
│ PK id                    VARCHAR(50)                                 │
│    incident_type         VARCHAR(50)                               │
│    service               VARCHAR(100)                              │
│    severity              VARCHAR(10) CHECK SEV1..SEV4            │
│    status                VARCHAR(50) CHECK investigating..closed │
│    datadog_url           TEXT                                        │
│    created_at            TIMESTAMP                                   │
│    closed_at             TIMESTAMP NULL                              │
│    updated_at            TIMESTAMP                                   │
└────────────────────────────────────────────────────────────────────┘
    │
    │ 1 ─────── N
    │
    ▼
┌────────────────────────────────────────────────────────────────────┐
│ findings                                                           │
├────────────────────────────────────────────────────────────────────┤
│ PK id                    SERIAL                                      │
│ FK incident_id           VARCHAR(50) → incidents.id                │
│    agent_name            VARCHAR(50) CHECK agent enum              │
│    finding               TEXT NOT NULL                                │
│    confidence            INTEGER CHECK 0-100                         │
│    suggested_action      TEXT                                        │
│    evidence              JSONB                                        │
│    source                VARCHAR(100)                               │
│    created_at            TIMESTAMP                                    │
└────────────────────────────────────────────────────────────────────┘
    │
    │ 1 ─────── N
    │
    ▼
┌────────────────────────────────────────────────────────────────────┐
│ timeline                                                           │
├────────────────────────────────────────────────────────────────────┤
│ PK id                    SERIAL                                      │
│ FK incident_id           VARCHAR(50) → incidents.id                │
│    actor                 VARCHAR(50)                                 │
│    event                 TEXT NOT NULL                                │
│    metadata              JSONB                                        │
│    timestamp             TIMESTAMP                                    │
└────────────────────────────────────────────────────────────────────┘
    │
    │ 1 ─────── 1 (opsional)
    │
    ▼
┌────────────────────────────────────────────────────────────────────┐
│ recommendations                                                    │
├────────────────────────────────────────────────────────────────────┤
│ PK id                    SERIAL                                      │
│ FK incident_id           VARCHAR(50) → incidents.id                │
│    recommendation        TEXT NOT NULL                                │
│    root_cause            TEXT                                        │
│    confidence            INTEGER CHECK 0-100                         │
│    risk                  VARCHAR(20) CHECK LOW..CRITICAL           │
│    approval_required     BOOLEAN DEFAULT FALSE                     │
│    affected_users        INTEGER                                     │
│    revenue_exposure      VARCHAR(100)                               │
│    recovery_time_estimate VARCHAR(50)                               │
│    created_at            TIMESTAMP                                    │
└────────────────────────────────────────────────────────────────────┘
    │
    │ 1 ─────── 1..N
    │
    ▼
┌────────────────────────────────────────────────────────────────────┐
│ approvals                                                          │
├────────────────────────────────────────────────────────────────────┤
│ PK id                    SERIAL                                      │
│ FK incident_id           VARCHAR(50) → incidents.id                │
│    approver              VARCHAR(100)                                │
│    status                VARCHAR(20) CHECK pending..rejected       │
│    action_type           VARCHAR(100)                               │
│    requested_at          TIMESTAMP                                    │
│    responded_at          TIMESTAMP                                  │
│    reason                TEXT                                        │
└────────────────────────────────────────────────────────────────────┘
    │
    │ 1 ─────── 1 (opsional)
    │
    ▼
┌────────────────────────────────────────────────────────────────────┐
│ rca                                                                │
├────────────────────────────────────────────────────────────────────┤
│ PK incident_id           VARCHAR(50) → incidents.id                │
│    summary               TEXT NOT NULL                                │
│    root_cause            TEXT NOT NULL                                │
│    resolution            TEXT                                        │
│    contributing_factors  JSONB                                        │
│    lessons_learned       TEXT                                        │
│    preventive_actions    JSONB                                        │
│    generated_by          VARCHAR(50)                                  │
│    created_at            TIMESTAMP                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## 5.4 SQLAlchemy Model Sync Strategy

1. Gunakan `alembic` untuk migration (opsional, bisa skip untuk demo)
2. Atau gunakan `Base.metadata.create_all(engine)` di startup aplikasi:

```python
# db.py
from app.models.base import Base

engine = create_engine(DATABASE_URL)
Base.metadata.create_all(bind=engine)
```

3. Untuk demo, gunakan opsi 2 (auto-create) agar tidak perlu setup alembic.

---

## 5.5 Sample Query Patterns

### Get Full Incident Detail

```sql
SELECT 
    i.*,
    json_agg(f.* ORDER BY f.created_at) as findings,
    json_agg(t.* ORDER BY t.timestamp) as timeline,
    r.*
FROM incidents i
LEFT JOIN findings f ON i.id = f.incident_id
LEFT JOIN timeline t ON i.id = t.incident_id
LEFT JOIN recommendations r ON i.id = r.incident_id
WHERE i.id = 'INC-2026-001'
GROUP BY i.id, r.id;
```

### Get Recent Incidents

```sql
SELECT id, incident_type, service, severity, status, 
       EXTRACT(EPOCH FROM (NOW() - created_at))/60 as age_minutes
FROM incidents
WHERE status != 'closed'
ORDER BY created_at DESC
LIMIT 20;
```

### Get Agent Confidence Distribution

```sql
SELECT agent_name, AVG(confidence) as avg_confidence, COUNT(*) as count
FROM findings
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY agent_name;
```

---

> Next: baca `06-docker-compose.md` untuk Docker Compose spec.
