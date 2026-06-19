from typing import TypedDict, Annotated, Sequence
import operator
from langchain_core.messages import BaseMessage

class IncidentState(TypedDict):
    messages: Annotated[Sequence[BaseMessage], operator.add]
    incident_id: str
    service_name: str
    timeline: list[str]
    findings: list[str]
    confidence_score: int
    risk_score: str
    recommendation: str
    status: str
    approval_required: bool
    approved: bool
