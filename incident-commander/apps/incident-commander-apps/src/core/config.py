from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    # Datadog settings
    dd_api_key: str = ""
    dd_app_key: str = ""
    dd_site: str = "datadoghq.com"
    dd_env: str = "development"
    dd_service: str = "ai-incident-commander"
    dd_version: str = "1.0.0"

    # LLM Provider (openrouter or bedrock)
    llm_provider: str = "openrouter"

    # OpenRouter settings
    openrouter_api_key: str = ""
    openrouter_model: str = "anthropic/claude-3.5-sonnet"

    # AWS / Bedrock settings
    aws_access_key_id: str = ""
    aws_secret_access_key: str = ""
    aws_default_region: str = "us-east-1"
    bedrock_model_id: str = "anthropic.claude-3-5-sonnet-20240620-v1:0"

    # Slack settings
    slack_bot_token: str = ""
    slack_app_token: str = ""
    slack_signing_secret: str = ""

    # Database
    database_url: str = ""

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

settings = Settings()
