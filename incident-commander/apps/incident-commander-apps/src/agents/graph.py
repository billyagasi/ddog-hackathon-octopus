import operator
from typing import Annotated
from langgraph.graph import StateGraph, END
from langgraph.prebuilt import ToolNode

from src.agents.state import IncidentState
from src.agents.incident_lead import call_incident_lead
from src.agents.infrastructure import call_infrastructure_ops
from src.agents.application import call_application_intelligence
from src.agents.service import call_service_management
from src.mcp.tools import ALL_TOOLS

def create_incident_graph() -> StateGraph:
    workflow = StateGraph(IncidentState)

    # Add Nodes
    workflow.add_node("incident_lead", call_incident_lead)
    workflow.add_node("infrastructure_ops", call_infrastructure_ops)
    workflow.add_node("application_intelligence", call_application_intelligence)
    workflow.add_node("service_management", call_service_management)
    
    # Tool Node for MCP
    tool_node = ToolNode(ALL_TOOLS)
    workflow.add_node("tools", tool_node)

    # Workflow Definition
    workflow.set_entry_point("incident_lead")

    # In a full implementation, you would use conditional edges to route to tools
    # or to sub-agents. For simplicity in this demo structure:
    workflow.add_edge("incident_lead", "infrastructure_ops")
    workflow.add_edge("infrastructure_ops", "application_intelligence")
    workflow.add_edge("application_intelligence", "service_management")
    workflow.add_edge("service_management", END)
    
    return workflow.compile()
