import logging
from fastapi import FastAPI
from src.core.config import settings

# OpenTelemetry configuration
from opentelemetry import trace
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from ddtrace import patch_all

# Patch all libraries with Datadog's ddtrace for APM and LLM Observability
patch_all()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="AI Incident Commander",
    description="Datadog-Native Autonomous Incident Response Platform",
    version=settings.dd_version,
)

# Instrument FastAPI with OpenTelemetry
FastAPIInstrumentor.instrument_app(app)

@app.get("/health")
def health_check():
    return {"status": "ok", "service": settings.dd_service}

# Placeholder for webhook endpoints from Datadog Watchdog or Slack
@app.post("/webhook/datadog")
def datadog_webhook(payload: dict):
    logger.info(f"Received Datadog webhook: {payload}")
    # TODO: Trigger LangGraph workflow
    return {"status": "accepted"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("src.main:app", host="0.0.0.0", port=8000, reload=True)
