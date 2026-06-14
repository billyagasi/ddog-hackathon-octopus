# docs/01-architecture.md

# High Level Architecture

```text
Datadog Platform

├── Watchdog
├── Watchdog RCA
├── Watchdog Impact Analysis
├── Bits AI
├── Incident Management
├── Service Catalog
├── SLO Management
├── Cloud Cost Management
├── Workflow Automation
├── APM
├── Logs
├── Traces
├── Database Monitoring
└── Events / Change Tracking

                     │
                     ▼

            FastAPI Alert Gateway

                     │
                     ▼

            LangGraph Orchestrator

                     │

 ┌──────────┬────────────┬────────────┬────────────┬────────────┐
 ▼          ▼            ▼            ▼            ▼

Incident   Infra &      App         Change      Business
Lead AI    Platform AI  AI          AI          Impact AI

                     │
                     ▼

              Decision Engine

                     │
                     ▼

          Risk & Confidence Engine

                     │
                     ▼

            Human Approval Layer

                     │
                     ▼

      Datadog Workflow Automation

                     │
                     ▼

             Slack Incident Room

                     │
                     ▼

               PostgreSQL

                     │
                     ▼

            Qdrant Knowledge Base

---------------------------------------------------------

OpenTelemetry

      │

      ▼

Datadog LLM Observability

Datadog APM

Datadog Metrics

Datadog Logs
```

## Core Components

### Datadog Intelligence Layer

* Watchdog
* Watchdog RCA
* Watchdog Impact Analysis
* Bits AI
* Service Catalog
* SLO Management
* Cloud Cost Management
* Incident Management
* Workflow Automation

### AI Orchestration Layer

* FastAPI
* LangGraph
* OpenRouter

### Collaboration Layer

* Slack
* Datadog Incident Management

### Data Layer

* PostgreSQL
* Qdrant

## Architecture Principle

AI Incident Commander does not replace Datadog intelligence.

Datadog provides:

* Detection
* Correlation
* Root Cause Signals
* Cost Signals
* Service Context

AI Incident Commander provides:

* Multi-agent Investigation
* Decision Making
* Risk Analysis
* Human Approval
* Knowledge Retention
