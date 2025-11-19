"""Base Tool classes."""

from typing import Any

from pydantic import BaseModel


# 工具输出结果数据模型
class ToolResult(BaseModel):
    """Tool execution result."""

    success: bool
    content: str = ""
    error: str | None = None


# 基础工具类
class Tool:
    """Base class for all tools."""

    @property
    def name(self) -> str:
        """Tool name."""
        raise NotImplementedError

    @property
    def description(self) -> str:
        """Tool description."""
        raise NotImplementedError

    @property
    def parameters(self) -> dict[str, Any]:
        """Tool parameters schema.(JSON Schema)"""
        raise NotImplementedError

    async def execute(self, *args, **kwargs) -> ToolResult:
        """Execute the tool with arbitrary arguments."""
        raise NotImplementedError

    def to_openai_schema(self) -> dict[str, Any]:
        """Convert the tool to OpenAI tool schema."""
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": self.parameters,
            },
        }
