"""LLM client wrapper that supports OpenAI provider.

This module provides a unified interface for OpenAI provider
through a single LLMClient class.
"""

import logging

from src.schema import LLMProvider, LLMResponse, Message
from src.llm.base import LLMClientBase
from src.llm.openai_client import OpenAIClient

logger = logging.getLogger(__name__)


class LLMClient:
    """LLM Client wrapper supporting OpenAI provider.

    This class provides a unified interface for OpenAI provider.
    It automatically instantiates the correct underlying client based on
    the provider parameter and appends the appropriate API endpoint suffix.

    Supported provider:
    - openai: Appends /v1 to api_base
    """

    def __init__(
        self,
        api_key: str,
        provider: LLMProvider = LLMProvider.OPENAI,
        api_base: str = "http://localhost:8080",
        model: str = "gpt-oss",
    ):
        """Initialize LLM client with specified provider.

        Args:
            api_key: API key for authentication
            provider: LLM provider (openai)
            api_base: Base URL for the API (default: http://localhost:8080)
                     Will be automatically suffixed with /v1 based on provider
            model: Model name to use
        """
        self.provider = provider
        self.api_key = api_key
        self.model = model

        # Append provider-specific suffix to api_base
        full_api_base = f"{api_base.rstrip('/')}/v1"

        self.api_base = full_api_base

        # Instantiate the appropriate client
        self._client: LLMClientBase
        if provider == LLMProvider.OPENAI:
            self._client = OpenAIClient(
                api_key=api_key,
                api_base=full_api_base,
                model=model,
            )
        else:
            raise ValueError(f"Unsupported provider: {provider}")

        logger.info(
            "Initialized LLM client with provider: %s, api_base: %s",
            provider,
            full_api_base,
        )

    async def generate(
        self,
        messages: list[Message],
        tools: list | None = None,
    ) -> LLMResponse:
        """Generate response from LLM.

        Args:
            messages: List of conversation messages
            tools: Optional list of Tool objects or dicts

        Returns:
            LLMResponse containing the generated content
        """
        return await self._client.generate(messages, tools)
