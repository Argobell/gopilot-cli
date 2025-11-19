"""Basic test for llama.cpp deployment.

This test verifies that the gopilot-cli library works correctly
with a llama.cpp server running on localhost:8080.
"""

import pytest
import asyncio
from src import LLMClient, Message, LLMProvider


@pytest.mark.asyncio
async def test_basic_chat():
    """Test basic chat functionality with llama.cpp server."""
    # Initialize client with llama.cpp server configuration
    client = LLMClient(
        api_key="dummy-key",  # llama.cpp typically doesn't require real API key
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss",  # Your model name
    )

    # Create a simple message
    messages = [
        Message(role="user", content="Hello! Please respond with 'Test successful'")
    ]

    # Generate response
    response = await client.generate(messages)

    # Verify response
    assert response.content is not None, "Response content should not be None"
    assert len(response.content) > 0, "Response content should not be empty"
    assert isinstance(response.content, str), "Response content should be a string"
    print(f"✓ Basic chat test passed!")
    print(f"  Response: {response.content}")


@pytest.mark.asyncio
async def test_multi_turn_conversation():
    """Test multi-turn conversation."""
    client = LLMClient(
        api_key="dummy-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss",
    )

    messages = [
        Message(role="user", content="What is 2+2?"),
        Message(role="assistant", content="2+2 equals 4."),
        Message(role="user", content="What about 4+4?"),
    ]

    response = await client.generate(messages)

    assert response.content is not None
    assert len(response.content) > 0
    print(f"✓ Multi-turn conversation test passed!")
    print(f"  Response: {response.content}")


@pytest.mark.asyncio
async def test_system_message():
    """Test that system messages are handled correctly."""
    client = LLMClient(
        api_key="dummy-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss",
    )

    messages = [
        Message(
            role="system",
            content="You are a helpful assistant that always responds with 'YES'.",
        ),
        Message(role="user", content="Are you working?"),
    ]

    response = await client.generate(messages)

    assert response.content is not None
    assert len(response.content) > 0
    print(f"✓ System message test passed!")
    print(f"  Response: {response.content}")


@pytest.mark.asyncio
async def test_tool_call_support():
    """Test tool calling functionality if supported by your llama.cpp build."""
    client = LLMClient(
        api_key="dummy-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss",
    )

    # Simple tool definition
    tools = [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "Get weather information for a city",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "city": {"type": "string", "description": "City name"}
                    },
                    "required": ["city"],
                },
            },
        }
    ]

    messages = [Message(role="user", content="What is the weather in Beijing?")]

    response = await client.generate(messages, tools=tools)

    # Check if response has tool calls or normal content
    assert response.content is not None or response.tool_calls is not None
    print(f"✓ Tool call test passed!")
    if response.tool_calls:
        print(f"  Tool calls: {len(response.tool_calls)}")
        for tool_call in response.tool_calls:
            print(f"    - {tool_call.function.name}")
    else:
        print(f"  Response: {response.content}")


if __name__ == "__main__":
    # Run tests manually without pytest
    print("Running tests with llama.cpp server...")
    print("Make sure your server is running on http://localhost:8080\n")

    async def run_tests():
        try:
            print("=" * 60)
            await test_basic_chat()
            print()

            print("=" * 60)
            await test_multi_turn_conversation()
            print()

            print("=" * 60)
            await test_system_message()
            print()

            print("=" * 60)
            await test_tool_call_support()
            print()

            print("=" * 60)
            print("\n✅ All tests completed!")
        except Exception as e:
            print(f"\n❌ Test failed with error: {e}")
            import traceback

            traceback.print_exc()

    asyncio.run(run_tests())
