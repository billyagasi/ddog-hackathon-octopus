# 06-slack-war-room.md

# Slack War Room

## Overview

Every incident automatically creates a dedicated Slack War Room.

The War Room acts as the primary collaboration space between:

* Incident Responders
* Platform Engineers
* SRE Teams
* Service Owners
* AI Incident Commander

The objective is to centralize all investigation findings, recommendations, approvals, and remediation actions into a single operational channel.

---

# Automatic Channel Creation

When a Datadog incident is detected:

```text
Watchdog Alert

↓

Incident Created

↓

Slack War Room Created

↓

AI Agents Join

↓

Investigation Begins
```

Example:

```text
#inc-sev1-payment-api-2026-001
```

---

# War Room Participants

## Incident Lead AI

Role:

Incident Commander

Responsibilities:

* Incident coordination
* Realtime updates
* Recommendation publishing
* Approval coordination
* Executive summaries

---

## Infrastructure Operations AI

Provides:

* Infrastructure findings
* Deployment findings
* Platform findings
* Cost findings

---

## Application Intelligence AI

Provides:

* Log findings
* Trace findings
* APM findings
* Root cause hypotheses

---

## Service Management AI

Provides:

* User impact
* Revenue impact
* SLO status
* Runbook recommendations
* RCA summaries

---

# Realtime Investigation Feed

The War Room receives investigation updates continuously.

Example:

```text
[Infrastructure Operations AI]

Deployment detected 6 minutes before incident.

Confidence: 84%
```

```text
[Application Intelligence AI]

Database timeout observed in 92% of traces.

Confidence: 91%
```

```text
[Service Management AI]

Affected Users: 14,523

Revenue Exposure:
$25,000/hour
```

---

# Recommendation Updates

As confidence increases, recommendations are updated.

Example:

```text
Current Recommendation

Rollback Deployment

Confidence: 94%

Risk: Low
```

---

# Approval Workflow

Certain actions require human approval.

Examples:

* Rollback Deployment
* Restart Service
* Scale Service
* Database Failover
* Execute Runbook

---

# Approval Example

```text
Recommendation

Rollback deployment payment-api:v2.4.1

Confidence:
94%

Risk:
Low

Approve?
```

Actions:

* Approve
* Reject
* Request Investigation

---

# Executive Status Updates

Incident Lead AI periodically publishes executive summaries.

Example:

```text
SEV1 Payment API Incident

Status:
Investigating

Affected Users:
14,523

Revenue Exposure:
$25,000/hour

Current Recommendation:
Rollback Deployment

Confidence:
94%
```

---

# Resolution Summary

Once resolved:

```text
Incident Resolved

Duration:
18 minutes

Root Cause:
Database timeout introduced by deployment v2.4.1

Resolution:
Rollback deployment

Affected Users:
14,523

Estimated Revenue Exposure:
$7,500
```

---

# Benefits

* Faster collaboration
* Centralized investigation
* Human-in-the-loop approvals
* Executive visibility
* Complete incident timeline
