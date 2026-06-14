# docs/02-agent-design.md

# Agent Design

## Incident Lead AI

Responsibilities:

* Incident ownership
* Agent orchestration
* Severity assignment
* Recommendation generation
* Approval workflow

Inputs:

* Watchdog Findings
* Datadog Incident Context
* Service Catalog

Outputs:

* Incident Summary
* Final Recommendation
* Confidence Score

---

## Infrastructure & Platform AI

Responsibilities:

* Kubernetes
* Infrastructure
* Database Monitoring
* Cloud Resources
* Cost Analysis

Datadog Sources:

* Infrastructure Monitoring
* Database Monitoring
* Cloud Cost Management
* Watchdog

Outputs:

* Infrastructure Findings
* Cost Findings
* Platform Findings

---

## Application Support AI

Responsibilities:

* Log Analysis
* Trace Analysis
* Dependency Analysis
* Service Analysis

Datadog Sources:

* APM
* Logs
* Traces
* Service Catalog

Outputs:

* Root Cause Hypothesis
* Application Findings

---

## Change Correlation AI

Responsibilities:

* Deployment Analysis
* Change Tracking Analysis
* Feature Flag Analysis
* Configuration Drift Analysis

Datadog Sources:

* Events
* Deployment Tracking
* Change Tracking

Outputs:

* Change Probability Score

Example:

Deployment:
92%

Infrastructure:
5%

Database:
3%

---

## Business Impact AI

Responsibilities:

* User Impact Analysis
* Revenue Impact Analysis
* SLO Analysis
* Error Budget Analysis

Datadog Sources:

* Watchdog Impact Analysis
* SLO Management
* Service Catalog

Outputs:

* Revenue Exposure
* User Impact
* Error Budget Risk
* Executive Summary

---

## Confidence Engine

Responsibilities:

* Aggregate confidence from all agents
* Weight Watchdog confidence
* Weight Bits AI findings
* Generate final recommendation confidence

Example:

Watchdog:
88

Bits AI:
91

Agents:
93

Final:
91

