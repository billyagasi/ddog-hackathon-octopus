# 00 Quick Start Guide

> Dokumen ini adalah anchor untuk seluruh implementasi demo AI Incident Commander.

---

## TL;DR (4 Jam)

Tujuan: buat demo end-to-end yang:
1. FastAPI trigger lewat HTTP
2. LangGraph orkestrasi 4 AI agent paralel
3. Setiap agent dipanggil via OpenRouter LLM
4. Findings di-aggregate jadi decision & recommendation
5. Semua output disimpan di Postgres
6. Slack thread dibangkitkan via Webhook
7. Datadog metrics & logs autoflush ke agent lokal

Langkah tercepat:
1. Docker Compose up (Postgres + Datadog Agent)
2. Jalankan `docker compose up --build`
3. Trigger scenario via `curl` ke `/incident/trigger`
4. Cek Slack thread + Datadog dashboard metrics

---

## Tech Stack Lock-in

| Kategori | Tool | Alasan |
|---------|------|--------|
| LLM Router | OpenRouter (Claude Sonnet / GPT-4o) | Aman, cepat, API key mudah, nanti switch Bedrock |
| AI Framework | LangGraph (state machine orchestration) | Native parallelism, DAG minimal, ReAct traceable |
| API | FastAPI + Uvicorn | Async-friendly, auto-docs, threaded agent callback support |
| DB | PostgreSQL (Alpine) via Docker | ACID untuk incident state, findings, timeline, approval |
| Knowledge (opt) | Qdrant (via Docker) | Similar incident search (demo opsional) |
| Slack | Incoming Webhook URL | Tanpa OAuth, langsung posting thread message |
| Datadog | DogStatsD + structured stdout | DD Agent lokal scrape metrics via UDP; log via stdout agent |
| Base Image | `python:3.11-slim` | Ringan, library modern tersedia |

---

## Environment Variables (`.env`)

```bash
# API & LLM
OPENROUTER_API_KEY=sk-or-...
LLM_MODEL=anthropic/claude-sonnet-4-20250514

# Database
DATABASE_URL=postgresql://postgres:postgres@db:5432/incident_commander
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=incident_commander

# Slack
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/XXXXX/YYYYY/ZZZZZ

# Datadog
DD_API_KEY=<datadog-api-key>
DD_AGENT_HOST=datadog-agent
DD_DOGSTATSD_PORT=8125
DD_ENV=demo
DD_SERVICE=ai-incident-commander
```

---

## Checklist Pre-Flight

- [ ] Pastikan Docker & Docker Compose terinstall
- [ ] Pastikan `docker compose ps` bisa run
- [ ] Export OpenRouter key
- [ ] Export Datadog API Key
- [ ] Buat Slack Incoming Webhook (Workspace yang kamu kontrol)
- [ ] Siapkan 2 terminal: satu untuk docker, satu untuk trigger

---

## Architecture Mindset

1. **Prompt is the Product** — setiap agent beda `system_prompt`, tapi core engine sama
2. **Stub Data di Prompt** — untuk 4 jam demo, kita inject mock observability context ke prompt LLM agar findings realistis tanpa query Datadog API real
3. **Fleksibel Via Tags** — setiap agent punya tag `incident_type`, `service`, `severity`, jadi generic engine tetap support berbagai use case
4. **Datadog-Native** — metrics & logs emitted sebagai first-class citizen demo Datadog

---

## Deliverable Demo

Setelah selesai, harus bisa demo ini di Slack thread + Datadog dashboard:

| Step | Hasil |
|------|-------|
| Trigger POST outage | `/incident/trigger` return `{"incident_id":"INC-001"}` |
| 4 Agent merespon | Slack thread punya messages: Infra, App, Change, Business |
| Decision Engine | Slack final message: Root Cause + Recommended Action |
| Datadog Metrics | `aic.incident.created`, `aic.agent.confidence`, `aic.incident.duration` muncul |
| Datadog Log | Structured JSON log terindex dengan `incident_id`, `agent`, `finding` |

---

> Next: baca `01-architecture.md` untuk deep dive diagram.
