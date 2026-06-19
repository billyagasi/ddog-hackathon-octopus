import logging
from langchain_aws import ChatBedrock
from langchain_core.messages import HumanMessage
from src.core.config import settings

logger = logging.getLogger(__name__)

def generate_rca(incident_state: dict) -> str:
    """
    Generate an Auto RCA document based on the incident state using Amazon Bedrock.
    """
    llm = ChatBedrock(
        model_id="anthropic.claude-3-sonnet-20240229-v1:0",
        region_name=settings.aws_default_region
    )

    prompt = f"""
    Generate a highly detailed Root Cause Analysis (RCA) report based on the following incident data.
    The report MUST contain these sections:
    1. Incident Summary
    2. Executive Summary
    3. Detailed Timeline
    4. Agent Findings
    5. Hypothesis Evolution
    6. Decision Analysis
    7. Approval History
    8. Remediation Actions
    9. Recovery Analysis
    10. Root Cause Analysis
    11. Contributing Factors
    12. Business Impact
    13. Preventive Actions
    14. Lessons Learned

    Incident Data:
    ID: {incident_state.get('incident_id')}
    Service: {incident_state.get('service_name')}
    Findings: {incident_state.get('findings')}
    Recommendation: {incident_state.get('recommendation')}
    Timeline: {incident_state.get('timeline')}
    """

    try:
        response = llm.invoke([HumanMessage(content=prompt)])
        logger.info(f"RCA generated for {incident_state.get('incident_id')}")
        return response.content
    except Exception as e:
        logger.error(f"Error generating RCA: {e}")
        return f"Failed to generate RCA: {str(e)}"
