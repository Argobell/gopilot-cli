from abc import ABC, abstractmethod
from typing import Any

from src.schema import LLMResponse, Message


class LLMClientBase(ABC):
    """Abstract base class for LLM clients.
    
    This class defines the interface that all LLM clients must implement
    """

    def __init__(
        self,
        api_key: str,
        api_base: str,
        model: str,
    ):
        """Initialize the LLM client.
        
        Args:
            api_key: API key for authentication
            api_base: Base URL for the API
            model: Model name to use
        """
        self.api_key = api_key
        self.api_base = api_base
        self.model = model
    
    @abstractmethod
    async def generate(
        self,
        messages: list[Message],
        tools: list[Any] | None = None,
    ) -> LLMResponse:
        """Generate response from LLM.
        
        Args:
            messages: List of conversation messages
            tools: Optional list of Tool objects or dicts
        
        Returns:
            LLMResponse containing the generated content, thinking, and tool calls
        """
        pass

    @abstractmethod
    def _prepare_request(
        self,
        messages: list[Message],
        tools: list[Any] | None = None,
    ) -> dict[str, Any]:
        """Prepare the request payload for the LLM API.
        
        Args:
            messages: List of conversation messages
            tools: Optional list of available tools
        Returns:
            Dictionary representing the request payload
        """
        pass

    @abstractmethod
    def _convert_messages(self, messages: list[Message]) -> tuple[str | None, list[dict[str, Any]]]:
        """Convert internal message format to API-specific format.

        Args:
            messages: List of internal Message objects

        Returns:
            Tuple of (system_message, api_messages)
        """
        pass
