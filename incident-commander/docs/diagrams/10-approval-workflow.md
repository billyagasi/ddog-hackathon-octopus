# 10 Approval Workflow Design

> Detail flow approval: kapan diperlukan, bagaimana direquest, dan bagaimana disetujui.

---

## 10.1 Approval Categories

### Auto-Approved (No Human Intervention)

| Action Type | Risk | Why Auto-Approved |
|------------|------|-------------------|
| Read-only investigation | None | Tidak ada side effect |
| Cache flush/invalidate | Low | Reversible, no data loss |
| Log level toggle | Low | Reversible |
| Pod restart (single) | Low | Kubes akan reschedule |

### Approval Required (Human Must Approve)

| Action Type | Risk | Why Approval |
|------------|------|-------------|
| Deployment rollback | Medium | Menargetkan production state |
| Scale deployment (Up/Down) | Medium | Cost & resource impact |
| Database failover | High | Data integrity risk |
| Traffic shift | Medium | User-facing routing change |
| Config change (infra) | Medium | Affects multiple services |

### Restricted (Requires Additional Authorization)

| Action Type | Risk | Why Restricted |
|------------|------|---------------|
| Data deletion | Critical | Irreversible |
| Database restore | High | Potentially data overwrite |
| Infrastructure destruction | Critical | Irreversible |

---

## 10.2 Approval Decision Matrix

```text
┌─────────────────────────────────────────────────────────────┐
│                     Decision Engine Output                    │
│                                                               │
│   Recommended Action                                          │
│         │                                                     │
│         ▼                                                     │
│   ┌─────────────────────────┐                                 │
│   │  Is action in restricted │                                │
│   │  list?                   │                                │
│   │                          │                                │
│   │   YES → BLOCK + Escalate │                                │
│   │         (not allowed)    │                                │
│   │                          │                                │
│   │   NO  → Continue         │                                │
│   └────────────┬─────────────┘                                │
│                │                                              │
│                ▼                                              │
│   ┌─────────────────────────┐                                 │
│   │  Is action in approval  │                                 │
│   │  required list?         │                                 │
│   │                          │                                 │
│   │   YES → Request Approval  │                               │
│   │         Set status:         │                               │
│   │         pending_approval    │                               │
│   │         Slack notification  │                               │
│   │         Wait for human      │                               │
│   │                          │                                 │
│   │   NO  → Auto-Approve       │                               │
│   │         Set status: approved │                              │
│   │         Execute action       │                              │
│   └────────────────────────────┘                               │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

---

## 10.3 Approval State Machine

```text
    ┌──────────┐
    │ DECISION │
    │ ENGINE   │
    └────┬─────┘
         │
         │ approval_required = true
         ▼
    ┌───────────────────────────┐
    │  STATUS: pending_approval  │
    │  - Insert row approvals    │
    │    status = 'pending'      │
    │  - Add timeline entry      │
    │  - Post Slack "Approval    │
    │    Required"               │
    │  - Emit metric             │
    │    approval.required       │
    └─────────────┬─────────────┘
                  │
         ┌────────┴────────┐
         │                 │
         │                 │
         ▼                 ▼
  ┌─────────────┐   ┌─────────────┐
  │  APPROVED   │   │  REJECTED   │
  │  (POST)     │   │  (POST)     │
  └──────┬──────┘   └──────┬──────┘
         │                 │
         ▼                 ▼
  ┌─────────────┐   ┌─────────────┐
  │ STATUS:     │   │ STATUS:     │
  │ approved    │   │ investigating│
  │ Execute     │   │ (re-investigate│
  │ action      │   │ or close)   │
  │ Emit metric │   │ Emit metric │
  │ approval    │   │ approval    │
  │ .resolved   │   │ .rejected   │
  └─────────────┘   └─────────────┘
         │
         ▼
  ┌─────────────┐
  │ Resolution  │
  │ - Generate  │
  │   RCA       │
  │ - Update KB │
  │ - Close     │
  └─────────────┘
```

---

## 10.4 Slack Approval Message Format

```text
🔴 INC-2026-001 | payment-api | SEV1

👨‍✈️ Incident Lead AI Decision:

Recommended Action: Rollback deployment to v2.1.4
Confidence: 94% | Risk: LOW

⚠️ APPROVAL REQUIRED

Execute this action?

[ APPROVE ]  [ REJECT ]  [ VIEW DETAILS ]
```

> **Note:** Incoming Webhook tidak support interactive buttons. Untuk demo 4 jam, gunakan approach:
> 1. Slack post berisi "Reply with: APPROVE or REJECT"
> 2. Atau: gunakan `/incident/{id}/approve` API endpoint via cURL

---

## 10.5 API Approval Flow

### Request Approval (Auto-generated by system)

```
POST /incident/{id}/approve-request (internal)
```

```json
{
  "action_type": "rollback_deployment",
  "action_detail": "Rollback payment-api from v2.1.5 to v2.1.4",
  "requested_by": "decision_engine"
}
```

### Submit Approval (by human)

```
POST /incident/{id}/approve
```

```json
{
  "approver": "engineer@company.com",
  "approved": true,
  "reason": "Confidence 94%, rollback is safe"
}
```

### Response

```json
{
  "incident_id": "INC-2026-001",
  "approval_id": 1,
  "status": "approved",
  "approver": "engineer@company.com",
  "responded_at": "2026-06-14T10:05:00Z",
  "next_step": "Execute rollback_deployment via workflow automation"
}
```

---

## 10.6 Approval Table Schema

```sql
CREATE TYPE approval_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE approvals (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(50) REFERENCES incidents(id),
    action_type VARCHAR(100),
    action_detail TEXT,
    requested_by VARCHAR(50) DEFAULT 'decision_engine',
    approver VARCHAR(100),
    status approval_status DEFAULT 'pending',
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE,
    reason TEXT
);
```

---

## 10.7 Decision → Action Mapping

```text
Recommended Action                  │ Risk │ Approval │ Action Type
────────────────────────────────────┼──────┼──────────┼─────────────
Rollback deployment                 │ MED  │ REQUIRED │ rollback_deployment
Scale deployment up                 │ MED  │ REQUIRED │ scale_up
Scale deployment down               │ LOW  │ AUTO     │ scale_down
Restart pod                         │ LOW  │ AUTO     │ restart_pod
Restart deployment                  │ MED  │ REQUIRED │ restart_deployment
Database failover                   │ HIGH │ REQUIRED │ db_failover
Traffic shift                       │ MED  │ REQUIRED │ traffic_shift
Cache flush                         │ LOW  │ AUTO     │ cache_flush
Run runbook                         │ LOW  │ AUTO     │ run_runbook
Reduce Aurora instance              │ MED  │ REQUIRED │ resize_aurora
Fix retry config                    │ MED  │ REQUIRED │ update_config
```

---

## 10.8 Security Considerations

1. **Approval must be logged** — tiap approval direkam dengan timestamp, approver, dan action
2. **No unsigned webhook** — validate Slack signature (opsional untuk demo)
3. **RBAC** — hanya users dengan role `incident_responder` yang bisa approve (future enhancement)
4. **Audit trail** — immutable, tidak bisa dihapus

---

> Next: baca `11-slack-format.md` untuk Slack message templates.
