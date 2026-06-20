from langchain_aws import ChatBedrock
from langchain_core.messages import SystemMessage
from src.agents.state import IncidentState
from src.agents.prompts import APP_INTELLIGENCE_PROMPT
from src.mcp.tools import query_logs, query_traces
from src.core.config import settings

def call_application_intelligence(state: IncidentState) -> dict:
    llm = ChatBedrock(
        model_id="anthropic.claude-3-sonnet-20240229-v1:0",
        region_name=settings.aws_default_region
    )
    llm_with_tools = llm.bind_tools([query_logs, query_traces])
    
    messages = [SystemMessage(content=APP_INTELLIGENCE_PROMPT)] + state["messages"]
    response = llm_with_tools.invoke(messages)
    
    return {"messages": [response]}
