import logging
from slack_bolt import App
from src.core.config import settings

logger = logging.getLogger(__name__)

slack_app = App(
    token=settings.slack_bot_token,
    signing_secret=settings.slack_signing_secret
)

def create_war_room(incident_id: str) -> str:
    """Create a dedicated Slack channel for the incident."""
    channel_name = f"inc-sev1-{incident_id.lower()}"
    try:
        # Mocking for hackathon context, typically:
        # response = slack_app.client.conversations_create(name=channel_name)
        # return response["channel"]["id"]
        logger.info(f"Created Slack channel: #{channel_name}")
        return channel_name
    except Exception as e:
        logger.error(f"Error creating Slack channel: {e}")
        return channel_name

def post_finding(channel: str, agent_name: str, finding: str, confidence: int):
    """Post an investigation finding to the War Room."""
    message = f"*[{agent_name}]*\n{finding}\n_Confidence: {confidence}%_"
    try:
        # slack_app.client.chat_postMessage(channel=channel, text=message)
        logger.info(f"Posted to {channel}: {message}")
    except Exception as e:
        logger.error(f"Error posting finding to Slack: {e}")

def request_approval(channel: str, recommendation: str, risk: str):
    """Send an interactive block message requesting human approval."""
    blocks = [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": f"*Recommendation*\n{recommendation}\n*Risk:* {risk}\n\n*Approve?*"
            }
        },
        {
            "type": "actions",
            "elements": [
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "Approve",
                        "emoji": True
                    },
                    "style": "primary",
                    "value": "approve_action",
                    "action_id": "approve_remediation"
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "Reject",
                        "emoji": True
                    },
                    "style": "danger",
                    "value": "reject_action",
                    "action_id": "reject_remediation"
                }
            ]
        }
    ]
    try:
        # slack_app.client.chat_postMessage(channel=channel, text="Approval Required", blocks=blocks)
        logger.info(f"Requested approval in {channel}")
    except Exception as e:
        logger.error(f"Error requesting approval: {e}")

# In a real Bolt app, you would define handlers for the action_ids:
@slack_app.action("approve_remediation")
def handle_approve(ack, body, logger):
    ack()
    logger.info(body)
    # Trigger Datadog Workflow Automation
    # Update channel message
