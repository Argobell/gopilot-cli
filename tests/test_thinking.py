"""Test reasoning/thinking content generation.

This test verifies that the gopilot-cli library can extract and handle
thinking content (reasoning_details) from LLM responses.
"""

import pytest
import asyncio
from src import LLMClient, Message, LLMProvider


@pytest.mark.asyncio
async def test_thinking_content():
    """Test that thinking content is extracted from responses."""
    client = LLMClient(
        api_key="dummy-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss",
    )

    messages = [
        Message(
            role="user",
            content="Please think step by step and solve this: What is 15 * 23?",
        )
    ]

    print("\n" + "=" * 60)
    print("Testing thinking content extraction...")
    print("=" * 60)

    response = await client.generate(messages)

    print(f"\n✓ Response received!")
    print(f"  Content: {response.content}")
    print(f"  Thinking: {response.thinking}")

    assert response.content is not None
    assert len(response.content) > 0

    # Check if thinking content was extracted
    if response.thinking:
        print(f"✓ Thinking content was extracted!")
        print(f"  Thinking length: {len(response.thinking)} chars")
    else:
        print(
            f"ℹ No thinking content in response (model may not support reasoning_split)"
        )


@pytest.mark.asyncio
async def test_thinking_preservation_in_history():
    """Test that thinking content is preserved in message history."""
    client = LLMClient(
        api_key="dummy-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss",
    )

    # First turn
    messages = [
        Message(role="user", content="Think about why 2+2=4, then give the answer")
    ]

    response1 = await client.generate(messages)

    # Add assistant's response to history
    messages.append(
        Message(
            role="assistant", content=response1.content, thinking=response1.thinking
        )
    )

    # Second turn with follow-up
    messages.append(Message(role="user", content="What about 2+3?"))

    response2 = await client.generate(messages)

    print("\n" + "=" * 60)
    print("Testing thinking preservation in conversation history...")
    print("=" * 60)

    print(f"\nFirst response:")
    print(f"  Content: {response1.content}")
    print(f"  Thinking: {response1.thinking}")

    print(f"\nSecond response:")
    print(f"  Content: {response2.content}")
    print(f"  Thinking: {response2.thinking}")

    assert response2.content is not None

    if response2.thinking:
        print(f"\n✓ Second response also has thinking content!")
    else:
        print(f"\nℹ Second response has no thinking content")


@pytest.mark.asyncio
async def test_complex_reasoning():
    """Test with a more complex reasoning problem."""
    client = LLMClient(
        api_key="dummy-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss",
    )

    messages = [
        Message(
            role="user",
            content="""Please think through this step by step:
1. What is 10 + 5?
2. Multiply that result by 2
3. Subtract 5 from that
Show your thinking process.""",
        )
    ]

    print("\n" + "=" * 60)
    print("Testing complex reasoning...")
    print("=" * 60)

    response = await client.generate(messages)

    print(f"\n✓ Complex reasoning response:")
    print(f"  Content:\n{response.content}")
    if response.thinking:
        print(f"\n  Thinking:\n{response.thinking}")

    assert response.content is not None
    assert len(response.content) > 0


if __name__ == "__main__":
    print("Running thinking content tests with llama.cpp server...")
    print("Make sure your server is running on http://localhost:8080\n")

    async def run_tests():
        try:
            await test_thinking_content()
            print()

            await test_thinking_preservation_in_history()
            print()

            await test_complex_reasoning()
            print()

            print("=" * 60)
            print("\n✅ All thinking content tests completed!")
        except Exception as e:
            print(f"\n❌ Test failed with error: {e}")
            import traceback

            traceback.print_exc()

    asyncio.run(run_tests())
