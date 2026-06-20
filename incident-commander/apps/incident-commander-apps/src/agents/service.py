from langchain_aws import ChatBedrock
from langchain_core.messages import SystemMessage
from src.agents.state import IncidentState
from src.agents.prompts import SERVICE_MANAGEMENT_PROMPT
from src.mcp.tools import query_slo
from src.core.config import settings

def call_service_management(state: IncidentState) -> dict:
    llm = ChatBedrock(
        model_id="anthropic.claude-3-sonnet-20240229-v1:0",
        region_name=settings.aws_default_region
    )
    llm_with_tools = llm.bind_tools([query_slo])
    
    messages = [SystemMessage(content=SERVICE_MANAGEMENT_PROMPT)] + state["messages"]
    response = llm_with_tools.invoke(messages)
    
    return {"messages": [response]}
