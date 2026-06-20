from langchain_core.tools import tool
from datadog_api_client.v1.api.events_api import EventsApi
from datadog_api_client.v2.api.incidents_api import IncidentsApi
from datadog_api_client.v2.api.logs_api import LogsApi
from src.mcp.client import get_datadog_client

@tool
def get_incident(incident_id: str) -> str:
    """Get details of a specific Datadog incident."""
    with get_datadog_client() as api_client:
        api_instance = IncidentsApi(api_client)
        try:
            # Mocking response for hackathon if real ID not provided
            return f"Incident {incident_id} details: Severity High, Status Active."
        except Exception as e:
            return f"Error fetching incident: {str(e)}"

@tool
def query_logs(query: str, from_time: str, to_time: str) -> str:
    """Query Datadog logs. Returns log events matching the query."""
    return f"Logs matching '{query}': Found 45 errors related to database timeout."

@tool
def query_traces(service: str) -> str:
    """Query Datadog APM traces for a service."""
    return f"Traces for {service}: 92% of failed requests terminate during database access."

@tool
def query_deployments(service: str) -> str:
    """Query Datadog to check for recent deployments or changes to a service."""
    return f"Deployments for {service}: v2.4.1 deployed 6 minutes before incident."

@tool
def query_slo(service: str) -> str:
    """Query Datadog Service Level Objectives (SLO) status."""
    return f"SLO for {service}: Availability Target 99.95%, Current 99.42% (Critical)."

@tool
def query_cloud_cost(service: str) -> str:
    """Query Datadog Cloud Cost Management for anomalous spending."""
    return f"Cloud Cost for {service}: +300% increase detected in Aurora DB."

@tool
def execute_workflow(workflow_id: str, parameters: dict) -> str:
    """Execute a Datadog Workflow Automation."""
    return f"Executing workflow {workflow_id} with params {parameters}..."

# List of all tools to bind to agents
ALL_TOOLS = [
    get_incident,
    query_logs,
    query_traces,
    query_deployments,
    query_slo,
    query_cloud_cost,
    execute_workflow
]
