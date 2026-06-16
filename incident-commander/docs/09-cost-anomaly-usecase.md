# AWS Cost Anomaly Use Case

## Overview

AI Incident Commander is capable of investigating cloud cost anomalies using Datadog Cloud Cost Management, Watchdog, and MCP-powered AI investigation.

The objective is to identify the root cause of unexpected cloud spending and recommend corrective actions before financial impact escalates.

---

# Scenario

Expected AWS Cost

```text
$300/day
```

Current AWS Cost

```text
$1,200/day
```

Increase

```text
+300%
```

---

# Detection

Datadog Watchdog detects an abnormal increase in AWS spending.

```text
Service:
payment-platform

Severity:
High

Cost Increase:
+300%
```

---

# Investigation

## Incident Lead AI

Launches cost anomaly investigation.

Requests findings from:

- Infrastructure Operations AI
- Application Intelligence AI
- Service Management AI

---

## Infrastructure Operations AI

### Findings

Aurora Cluster Scale Event

```text
db.r6.large

↓

db.r6.8xlarge
```

Timestamp

```text
09:42 AM
```

---

### Findings

Unexpected Resource Growth

```text
Aurora CPU

42%

↓

96%
```

Confidence

```text
88%
```

---

## Application Intelligence AI

### Findings

Retry Storm Detected

Evidence

```text
Request Retries

Normal:
1,200/hour

Current:
47,000/hour
```

---

### Findings

Connection Timeout Errors

Evidence

```text
Timeout Errors

+1,850%
```

Confidence

```text
92%
```

---

## Service Management AI

### Findings

Projected Monthly Impact

```text
$27,000/month
```

---

### Findings

Business Risk

```text
High
```

---

# Root Cause

Retry storm caused abnormal database load.

Database autoscaling increased infrastructure costs.

Cloud spending increased significantly.

---

# Recommendation

## Immediate Actions

- Reduce Aurora capacity
- Fix retry policy
- Limit retry attempts

---

## Long-Term Actions

- Implement retry budget
- Add cost anomaly guardrails
- Add autoscaling thresholds

---

# Outcome

Estimated Savings

```text
$18,000/month
```

Incident Status

```text
Resolved
```