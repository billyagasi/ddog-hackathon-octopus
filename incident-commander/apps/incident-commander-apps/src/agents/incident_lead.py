from langchain_aws import ChatBedrock
from langchain_core.messages import SystemMessage
from src.agents.state import IncidentState
from src.agents.prompts import INCIDENT_LEAD_PROMPT
from src.mcp.tools import get_incident
from src.core.config import settings

def call_incident_lead(state: IncidentState) -> dict:
    llm = ChatBedrock(
        model_id="anthropic.claude-3-sonnet-20240229-v1:0",
        region_name=settings.aws_default_region
    )
    llm_with_tools = llm.bind_tools([get_incident])
    
    messages = [SystemMessage(content=INCIDENT_LEAD_PROMPT)] + state["messages"]
    response = llm_with_tools.invoke(messages)
    
    # Simple state update mapping
    return {"messages": [response]}
