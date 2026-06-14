# 📋 AI Incident Commander — Strategy Plan Index

> Daftar lengkap semua dokumen strategi implementasi yang telah dibuat.

---

## 🎯 Alur Pembacaan & Prioritas

Urutkanlah pembacaan sesuai nomor agar pemahaman bertahap:

| No | File | Fokus | Prioritas |
|----|------|-------|-----------|
| 00 | `00-quick-start.md` | TL;DR 4 jam, stack lock-in, checklist | P0 |
| 01 | `01-architecture.md` | Diagram arsitektur high level & alur data | P0 |
| 02 | `02-data-flow.md` | Sequence diagrams detail per step | P0 |
| 03 | `03-file-structure.md` | Blueprint tree lengkap semua file | P0 |
| 04 | `04-agent-prompts.md` | Semua LLM prompts & mock contexts | P0 |
| 05 | `05-database-schema.md` | DDL PostgreSQL & ERD | P0 |
| 06 | `06-docker-compose.md` | Docker Compose, Dockerfile, env | P0 |
| 07 | `07-api-contracts.md` | API endpoints & request/response specs | P0 |
| 08 | `08-datadog-metrics.md` | Metrics, logs, dashboard spec | P0 |
| 09 | `09-llm-integration.md` | OpenRouter spec & switch ke Bedrock | P0 |
| 10 | `10-approval-workflow.md` | State machine approval | P1 |
| 11 | `11-slack-format.md` | Slack message templates | P1 |
| 12 | `12-opencode-runbook.md` | Runbook eksekusi opencode step-by-step | **Kunci** |

---

## 🚀 Perbedaan Dokumen Lama vs Dokumen Baru

| Dokumen Lama (concept) | Dokumen Baru (implementation-ready) |
|-------------------------|--------------------------------------|
| `docs/00-overview.md` → Konsep bisnis | `diagrams/00-quick-start.md` → Execute plan |
| `docs/01-architecture.md` → Arsitektur konseptual | `diagrams/01-architecture.md` → Diagram teknis lengkap |
| `docs/02-agent-design.md` → Tanggung jawab agent | `diagrams/04-agent-prompts.md` → Prompts + mock context siap copy-paste |
| `docs/03-workflow.md` → Workflow linear | `diagrams/02-data-flow.md` → Sequence diagram mendetail |
| `docs/04-observability.md` → Metric list | `diagrams/08-datadog-metrics.md` → Implementasi DogStatsD + log format + cost |
| `docs/05-data-model.md` → Schema ringkas | `diagrams/05-database-schema.md` → DDL lengkap + indexes + FK |
| `docs/06-slack-war-room.md` → Fitur | `diagrams/11-slack-format.md` → Templates + payload structure |
| `docs/07-executive-dashboard.md` → Dashboard konsep | `diagrams/08-datadog-metrics.md` → Query spec + widget config |
| `docs/08-outage-usecase.md` → Skenario | `diagrams/04-agent-prompts.md` → Mock context + decision engine input |
| `docs/09-billing-anomaly-usecase.md` → Skenario | `diagrams/04-agent-prompts.md` → Mock context billing + projected cost |
| `docs/10-mvp-roadmap.md` → Roadmap fase | `diagrams/12-opencode-runbook.md` → Phase dengan time allocation + script |
| `docs/11-datadog-capabilities.md` → Capability mapping | `diagrams/12-opencode-runbook.md` → Concrete DogStatsD emission code |

---

## ⏱️ File Allocations (Rencana 4 Jam)

| File | Estimasi Waktu Generate | By Opencode Prompt |
|------|------------------------|-------------------|
| `docker-compose.yml` | 5 menit | `Generate docker-compose.yml...` |
| `Dockerfile` | 5 menit | `Generate Dockerfile...` |
| `requirements.txt` | 2 menit | `Generate requirements.txt...` |
| `app/config.py` | 5 menit | `Generate Pydantic Settings...` |
| `app/db.py` + models | 10 menit | `Generate SQLAlchemy models...` |
| `app/agents/base.py` | 10 menit | `Generate BaseAgent + LLMClient...` |
| `app/agents/prompts/*.py` | 15 menit | `Generate prompts with mock_context...` |
| `app/agents/*_agent.py` | 15 menit | `Generate agent implementations...` |
| `app/orchestrator/*.py` | 20 menit | `Generate LangGraph state machine...` |
| `app/services/*.py` | 15 menit | `Generate CRUD services...` |
| `app/routers/*.py` + `main.py` | 15 menit | `Generate FastAPI routers...` |
| `app/integrations/*.py` | 15 menit | `Generate Slack + Datadog client...` |
| `scripts/*.sh` | 5 menit | `Generate demo trigger scripts...` |
| Build & Test | 30-60 menit | Manual / opencode fix |
| **Total** | **~4 jam** | |

---

## ✅ Checklist Sebelum Mulai Generate

- [ ] Export `OPENROUTER_API_KEY`
- [ ] Export `DD_API_KEY`
- [ ] Setup Slack Incoming Webhook
- [ ] `docker --version` berfungsi
- [ ] `docker compose version` berfungsi
- [ ] Working directory kosong (kecuali `docs/`)

---

## 🔗 Referensi Lintas File

Saat prompt opencode, selalu referensikan file spesifik:

```
Refer to file: /mnt/c/Users/billy-sre/Documents/verihubs/code/lab/ddog-hackathon-octopus/incident-commander/diagrams/01-architecture.md
```

---

## 🎬 Langkah Berikutnya

Setelah membaca file ini, langkah berikutnya adalah:

1. **Baca `12-opencode-runbook.md`** — ini adalah runbook eksekusi
2. **Ikuti Phase 1-7** sesuai runbook
3. **Untuk setiap step**, gunakan prompt yang sudah disediakan di runbook
4. **Generate paralel** dimungkinkan (misal: Phase 1 & Phase 5 bisa di-generate bareng karena tidak saling depend)

---

**Strategy Plan siap di-execute. Semoga demo sukses!**
