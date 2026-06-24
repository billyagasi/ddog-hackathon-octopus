import logging
from typing import Any
from src.core.config import settings

logger = logging.getLogger(__name__)

def get_llm(**kwargs: Any) -> Any:
    """
    Returns an instance of a LangChain Chat Model based on configuration.
    Supports OpenRouter and AWS Bedrock.
    """
    provider = settings.llm_provider.lower()
    
    if provider == "openrouter":
        logger.info("Initializing OpenRouter LLM")
        from langchain_openai import ChatOpenAI
        
        # OpenRouter uses the OpenAI API standard
        return ChatOpenAI(
            base_url="https://openrouter.ai/api/v1",
            api_key=settings.openrouter_api_key,
            model=settings.openrouter_model,
            **kwargs
        )
        
    elif provider == "bedrock":
        logger.info("Initializing AWS Bedrock LLM")
        from langchain_aws import ChatBedrock
        
        return ChatBedrock(
            model_id=settings.bedrock_model_id,
            credentials_profile_name=None, # Typically rely on env vars or IAM roles
            region_name=settings.aws_default_region,
            **kwargs
        )
        
    else:
        raise ValueError(f"Unsupported LLM provider: {provider}")

if __name__ == "__main__":
    # Simple test for initialization
    import sys
    logging.basicConfig(level=logging.INFO)
    try:
        llm = get_llm()
        print(f"Successfully initialized: {llm}")
    except Exception as e:
        print(f"Error initializing LLM: {e}")
        sys.exit(1)
