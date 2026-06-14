# docs/06-slack-war-room.md

# Slack War Room

Every Datadog alert creates a dedicated Slack thread.

Example:

INC-2026-001

Service:
payment-api

Severity:
SEV1

Status:
Investigating

---

## Participants

### Incident Lead AI

Coordinates investigation.

### Infrastructure AI

Investigates infrastructure.

### Application AI

Investigates application.

### Incident Management AI

Analyzes business impact.

---

## Approval Workflow

Actions requiring approval:

* Rollback deployment
* Scale deployment
* Database failover

Approval happens directly inside Slack.
