# 06 Docker Compose Specification

> Spesifikasi lengkap Docker Compose untuk local development dan demo.

---

## 6.1 Service Architecture

```text
┌─────────────────────────────────────────────┐
│           Docker Network: incident-net       │
│                                              │
│  ┌──────────────┐     ┌──────────────┐     │
│  │   postgres    │     │ datadog-agent │     │
│  │   :5432       │     │   :8125(udp) │ ◄───┤
│  └──────┬───────┘     └──────┬───────┘     │
│         │                    │             │
│         │                    │             │
│  ┌──────▼────────────────────▼───────────┐ │
│  │              app                         │ │
│  │          :8000 (FastAPI)                │ │
│  │                                          │ │
│  │   → query postgres                       │ │
│  │   → emit DogStatsD UDP → datadog-agent   │ │
│  │   → POST Slack Webhook                   │ │
│  └──────────────────────────────────────────┘ │
│                                              │
└──────────────────────────────────────────────┘
```

---

## 6.2 docker-compose.yml

```yaml
version: "3.8"

networks:
  incident-net:
    driver: bridge

services:
  # ──────────────────────────────────────────
  # PostgreSQL Database
  # ──────────────────────────────────────────
  db:
    image: postgres:15-alpine
    container_name: incident-postgres
    restart: unless-stopped
    networks:
      - incident-net
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgres}
      POSTGRES_DB: ${POSTGRES_DB:-incident_commander}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-postgres} -d ${POSTGRES_DB:-incident_commander}"]
      interval: 5s
      timeout: 5s
      retries: 5

  # ──────────────────────────────────────────
  # Datadog Agent (DogStatsD + Log Forwarding)
  # ──────────────────────────────────────────
  datadog-agent:
    image: gcr.io/datadoghq/agent:latest
    container_name: incident-datadog-agent
    restart: unless-stopped
    networks:
      - incident-net
    environment:
      DD_API_KEY: ${DD_API_KEY}
      DD_SITE: datadoghq.com
      DD_ENV: ${DD_ENV:-demo}
      DD_SERVICE: ai-incident-commander
      DD_DOGSTATSD_NON_LOCAL_TRAFFIC: "true"
      # DogStatsD: allow UDP from other containers
      DD_DOGSTATSD_PORT: "8125"
    ports:
      - "8125:8125/udp"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /proc/:/host/proc/:ro
      - /sys/fs/cgroup:/host/sys/fs/cgroup:ro
    healthcheck:
      test: ["CMD", "agent", "status"]
      interval: 10s
      timeout: 5s
      retries: 3

  # ──────────────────────────────────────────
  # AI Incident Commander App
  # ──────────────────────────────────────────
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: incident-commander
    restart: unless-stopped
    networks:
      - incident-net
    ports:
      - "8000:8000"
    environment:
      # Database
      DATABASE_URL: postgresql://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-postgres}@db:5432/${POSTGRES_DB:-incident_commander}
      # SLM / LLM
      OPENROUTER_API_KEY: ${OPENROUTER_API_KEY}
      LLM_MODEL: ${LLM_MODEL:-anthropic/claude-sonnet-4-20250514}
      # Slack
      SLACK_WEBHOOK_URL: ${SLACK_WEBHOOK_URL}
      # Datadog
      DD_AGENT_HOST: datadog-agent
      DD_DOGSTATSD_PORT: "8125"
      DD_API_KEY: ${DD_API_KEY}
      # General
      LOG_LEVEL: ${LOG_LEVEL:-INFO}
    depends_on:
      db:
        condition: service_healthy
      datadog-agent:
        condition: service_started
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    command: >
      sh -c "alembic upgrade head 2>/dev/null || echo 'Skipping migration';
             uvicorn app.main:app --host 0.0.0.0 --port 8000 --log-level ${LOG_LEVEL:-info}"

volumes:
  postgres_data:
```

> **Note:** Sesuaikan `DD_SITE` dengan region Datadog Anda (`datadoghq.com`, `datadoghq.eu`, `us5.datadoghq.com`, dll).

---

## 6.3 Dockerfile

```dockerfile
# Stage 1: Base image
FROM python:3.11-slim AS base

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Stage 2: Dependencies
FROM base AS deps

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Stage 3: Final image
FROM deps AS final

COPY app/ app/
COPY scripts/ scripts/
COPY alembic.ini alembic.ini
COPY alembic/ alembic/

# Expose FastAPI port
EXPOSE 8000

# Healthcheck endpoint
HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=5 \
    CMD curl -f http://localhost:8000/health || exit 1

# Default: auto-create tables + run uvicorn
CMD ["sh", "-c", "python -c 'import app.db; app.db.create_tables()' && uvicorn app.main:app --host 0.0.0.0 --port 8000"]
```

---

## 6.4 Environment Variables Template (`.env.template`)

```bash
# ──────────────────────────────────────────
# AI / LLM
# ──────────────────────────────────────────
OPENROUTER_API_KEY=sk-or-xxxxx
LLM_MODEL=anthropic/claude-sonnet-4-20250514

# ──────────────────────────────────────────
# Database
# ──────────────────────────────────────────
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-secure-password-here
POSTGRES_DB=incident_commander

# ──────────────────────────────────────────
# Slack
# ──────────────────────────────────────────
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX

# ──────────────────────────────────────────
# Datadog
# ──────────────────────────────────────────
DD_API_KEY=your-datadog-api-key-here
DD_ENV=demo

# ──────────────────────────────────────────
# App Settings
# ──────────────────────────────────────────
LOG_LEVEL=INFO
```

---

## 6.5 Docker Commands Cheat Sheet

```bash
# Start semua services
$ docker compose up -d

# View logs
$ docker compose logs -f app
$ docker compose logs -f db
$ docker compose logs -f datadog-agent

# Restart app
$ docker compose restart app

# Exec ke container
$ docker compose exec app bash
$ docker compose exec db psql -U postgres -d incident_commander

# Check health
$ docker compose ps
$ curl http://localhost:8000/health

# Shutdown & cleanup
$ docker compose down -v
```

---

## 6.6 Multi-stage Build Rationale

1. **Base Stage** — minimal OS + Python + system libs
2. **Deps Stage** — install pip dependencies (cacheable)
3. **Final Stage** — copy app code + set entrypoint

Keuntungan:
- Layer caching: requirements.txt jarang berubah, build cepat
- Final image lebih kecil: hanya hasil final, tidak ada compiler
- Reproducible: tiap build environment identik

---

## 6.7 Network & Security

- `incident-net` bridge network: semua service komunikasi internal tanpa expose port ke host
- DogStatsD UDP port `8125` di-expose ke host (agar bisa monitor dari luar Docker)
- PostgreSQL port `5432` di-expose untuk debugging (opsional: hilangkan di production)
- Tidak ada persistent secret: semua via `.env` (untuk demo acceptable)

---

> Next: baca `07-api-contracts.md` untuk API spec detail.
