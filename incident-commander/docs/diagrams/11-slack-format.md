# 11 Slack Message Templates & Format

> Format lengkap message yang diposting ke Slack untuk setiap fase incident.

---

## 11.1 Slack Integration Strategy

**Approach untuk Demo 4 Jam:**
- Gunakan **Slack Incoming Webhook URL** (bukan Slack Bot/App)
- Keuntungan: tidak perlu OAuth, token, scope, atau Slack App setup
- Limitation: tidak support thread_ts (tidak bisa native threads)
- **Workaround:** Kirim semua message sebagai 1 unified long message

**Alternatif (jika punya waktu tambahan):**
- Setup Slack Bot App dengan `chat:write`, `chat:write.public`
- Dapat thread_ts, interactive buttons, reactions

---

## 11.2 Message Templates

### Template A: Incident Header

```markdown
🔴 **INC-2026-001** | `payment-api` | **SEV1**

**Status:** Investigating
**Type:** Production Outage
**Severity:** SEV1 (< 30 min recovery target)
**Started:** 2026-06-14 10:00 UTC
**Datadog:** <https://app.datadoghq.com/monitors/12345|View Monitor>
─────────────────────────────
```

### Template B: Agent Finding

```markdown
🤖 **Infrastructure & Platform AI**
> Deployment v2.1.5 correlated with CPU spike and pod crash loops. Likely resource limit regression.
> 
> **Confidence:** 92%
> **Suggested Action:** Rollback deployment to v2.1.4
> **Evidence:** `k8s_deployment_event`, `cpu_metrics`
```

### Template C: Business Impact

```markdown
💰 **Business Impact AI**
> **Affected Users:** 14,500
> **Failed Transactions:** 4,500
> **Revenue Exposure:** $25,000/hour
> **SLA Breach Risk:** 🔴 HIGH
> **Current Availability:** 65.5% (Target: 99.95%)
```

### Template D: Decision & Recommendation

```markdown
👨‍✈️ **Incident Lead AI Decision**

**Root Cause:**
> Deployment v2.1.5 introduced CPU regression due to unbounded goroutines.

**Contributing Factors:**
• No resource limit on new worker pool
• HPA max replicas reached

**Recommended Action:**
> 🔄 Rollback deployment to v2.1.4

**Confidence:** 94%
**Risk:** 🟢 LOW
**Estimated Recovery:** 15 minutes

⚠️ **APPROVAL REQUIRED**
Reply with `APPROVE` or visit API endpoint.
```

### Template E: Approved & Executed

```markdown
✅ **APPROVED** by engineer@company.com (10:05 UTC)

**Action:** Rollback deployment to v2.1.4
**Status:** Executing...

🤖 **Executing via Datadog Workflow Automation...**
```

### Template F: Resolution & RCA

```markdown
✅ **RESOLVED** (10:45 UTC)
**Duration:** 45 minutes

**Resolution:**
> Rollback deployment v2.1.5 to v2.1.4 completed successfully. Latency returned to baseline.

**RCA Summary:**
• **Root Cause:** v2.1.5 introduced unbounded goroutine pool
• **Trigger:** Increased connection load → DB pool exhaustion
• **Fix:** Rollback + add resource limits in v2.1.6
• **Prevention:** Add load test gate in CI/CD pipeline

📊 View full RCA: <https://app.datadoghq.com/ai-incident-commander/INC-001|Dashboard>
```

---

## 11.3 Unified Message Format (for Incoming Webhook)

Karena Incoming Webhook tidak support multiple messages/threads, kirim 1 message per major state change:

```python
def format_unified_message(incident, findings, recommendation, approval_status=None):
    parts = [
        format_header(incident),
        "\n",
    ]
    
    for finding in findings:
        parts.append(format_agent_finding(finding))
        parts.append("\n")
    
    parts.append(format_recommendation(recommendation))
    
    if approval_status:
        parts.append(format_approval_status(approval_status))
    
    return "\n".join(parts)
```

---

## 11.4 Billing Anomaly Specific Format

```markdown
🟡 **INC-2026-044** | `payment-api` | **SEV2**

**Status:** Investigating
**Type:** AWS Billing Anomaly
**Expected Daily Cost:** $300 | **Current:** $1,200 (+300%)
─────────────────────────────

🏗️ **Infrastructure & Platform AI**
> Aurora PostgreSQL upgraded from db.r6.large to db.r6.4xlarge
> Cost Contribution: 65%
> Confidence: 92%

🤖 **Application Support AI**
> Retry storm detected. 8M additional requests generated.
> Confidence: 90%

💰 **Business Impact AI**
> Additional Daily Cost: $900
> Projected Monthly Impact: $27,000
> Business Impact: 🔴 HIGH

👨‍✈️ **Incident Lead AI Decision**

**Root Cause:** Aurora scale-up + retry storm

**Recommended Actions:**
1. 🔄 Reduce Aurora instance size to db.r6.large
2. 🔧 Fix retry configuration (max retries: 3)
3. 📊 Review HPA max replicas

**Confidence:** 94%
**Risk:** 🟡 MEDIUM

⚠️ **APPROVAL REQUIRED**
```

---

## 11.5 Slack Payload Structure

```json
{
  "text": "🔴 INC-2026-001 | payment-api | SEV1\n... (full message)",
  "username": "AI Incident Commander",
  "icon_emoji": ":robot_face:",
  "attachments": [
    {
      "color": "danger",
      "fields": [
        {
          "title": "Status",
          "value": "Investigating",
          "short": true
        },
        {
          "title": "Severity",
          "value": "SEV1",
          "short": true
        },
        {
          "title": "Affected Users",
          "value": "14,500",
          "short": true
        },
        {
          "title": "Revenue Exposure",
          "value": "$25,000/hour",
          "short": true
        }
      ],
      "footer": "AI Incident Commander",
      "ts": 1718359200
    }
  ]
}
```

---

## 11.6 Color Coding

| Severity/Status | Slack Color |
|---------------|-------------|
| SEV1 / Active | 🔴 `#FF0000` (danger) |
| SEV2 / Warning | 🟡 `#FFAA00` (warning) |
| SEV3 / Info | 🔵 `#36A64F` (good) |
| SEV4 / Resolved | ⚪ `#808080` (#808080) |
| Approved | 🟢 `#36A64F` (good) |
| Rejected | 🔴 `#FF0000` (danger) |

---

## 11.7 Datadog Incident Integration (Opsional)

Jika ingin buat Datadog Incident juga:

```python
# Create Datadog Incident via API
datadog_api.create_incident(
    title=f"[{severity}] {service} - {incident_type}",
    description=message_text,
    severity=severity,  # SEV1, SEV2, dll.
    service=service,
    tags=["source:ai-incident-commander"]
)
```

---

> Next: baca `12-runbook-opencode.md` untuk runbook eksekusi opencode.
