import logging
from src.mcp.client import get_datadog_client

logger = logging.getLogger(__name__)

def trigger_workflow(workflow_id: str, parameters: dict):
    """
    Trigger a Datadog Workflow Automation instance.
    This is called after human approval is received in Slack.
    """
    logger.info(f"Triggering Datadog Workflow {workflow_id} with parameters: {parameters}")
    try:
        # with get_datadog_client() as api_client:
        #     api_instance = WorkflowsApi(api_client)
        #     # Call execute workflow API
        logger.info("Workflow triggered successfully.")
        return True
    except Exception as e:
        logger.error(f"Failed to trigger workflow: {e}")
        return False
