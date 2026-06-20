from datadog_api_client import ApiClient, Configuration
from src.core.config import settings

def get_datadog_client() -> ApiClient:
    configuration = Configuration()
    # Configure API key authorization
    configuration.api_key["apiKeyAuth"] = settings.dd_api_key
    # Configure APP key authorization
    configuration.api_key["appKeyAuth"] = settings.dd_app_key
    configuration.server_variables["site"] = settings.dd_site

    return ApiClient(configuration)

# Placeholder client for Bedrock if needed directly, though LangChain handles it usually
def get_bedrock_client():
    import boto3
    return boto3.client(
        service_name="bedrock-runtime",
        region_name=settings.aws_default_region,
        aws_access_key_id=settings.aws_access_key_id,
        aws_secret_access_key=settings.aws_secret_access_key,
    )
