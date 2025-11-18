# Gopilot-CLI 开发文档

## 1. 项目简介

### 1.1 项目概述

Gopilot-CLI 是一个统一的 LLM（大语言模型）客户端接口库，提供简洁、易用的 API 来与各种 LLM 提供商进行交互。该项目采用 Python 3.12+ 开发，使用 OpenAI SDK 和 Pydantic 构建，具有良好的类型安全和数据验证能力。

### 1.2 核心特性

- **统一接口**：提供抽象基类，支持多种 LLM 提供商
- **OpenAI 支持**：完整的 OpenAI Provider 实现
- **推理内容支持**：支持模型的思考过程（reasoning_details）
- **工具调用**：支持 function/tool calling 能力
- **类型安全**：基于 Pydantic 的数据模型和验证
- **自动端点处理**：自动为不同提供商添加合适的 API 端点后缀

### 1.3 技术栈

- Python 3.12+
- OpenAI SDK (>=2.8.1)
- Pydantic (>=2.12.4)

## 2. 项目结构

```
gopilot-cli/
├── src/
│   ├── __init__.py              # 包初始化，导出主要 API
│   ├── llm/
│   │   ├── __init__.py          # LLM 模块初始化
│   │   ├── base.py              # 抽象基类定义
│   │   ├── openai_client.py     # OpenAI 提供商实现
│   │   └── llm_wrapper.py       # LLM 客户端包装器
│   └── schema/
│       ├── __init__.py          # Schema 模块初始化
│       └── schema.py            # 数据模型定义
├── pyproject.toml               # 项目配置和依赖
└── README.md                    # 项目说明
```

### 2.1 核心文件说明

- **`src/llm/base.py`**：定义了 `LLMClientBase` 抽象基类，是所有 LLM 客户端的接口规范
- **`src/llm/openai_client.py`**：`OpenAIClient` 类，实现 OpenAI Provider 的具体逻辑
- **`src/llm/llm_wrapper.py`**：`LLMClient` 统一包装器，根据 provider 参数自动选择合适的客户端
- **`src/schema/schema.py`**：定义 `Message`、`LLMResponse`、`ToolCall` 等 Pydantic 数据模型

## 3. 快速开始

### 3.1 安装依赖

```bash
# 克隆项目
git clone <repository-url>
cd gopilot-cli

# 安装依赖
uv sync --dev
```

### 3.2 基本使用

```python
import asyncio
from src import LLMClient, Message, LLMProvider

async def main():
    # 初始化客户端
    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    # 创建消息
    messages = [
        Message(role="system", content="你是一个有用的助手。"),
        Message(role="user", content="你好，请介绍一下自己。")
    ]

    # 生成响应
    response = await client.generate(messages)
    print(response.content)

asyncio.run(main())
```

### 3.3 环境要求

- Python 3.12 或更高版本
- 有效的 API 密钥
- 支持 OpenAI 兼容 API 的后端服务（如 llama.cpp、OpenAI API 等）

## 4. 使用示例

### 4.1 基本对话

```python
from src import LLMClient, Message

client = LLMClient(
    api_key="sk-xxxxx",
    provider=LLMProvider.OPENAI,
    api_base="https://api.openai.com",
    model="gpt-4"
)

messages = [
    Message(role="system", content="你是一个专业的 Python 开发者。"),
    Message(role="user", content="请写一个快速排序算法。")
]

response = await client.generate(messages)
print("回答:", response.content)
```

### 4.2 多轮对话

```python
messages = [
    Message(role="system", content="你是一个有用的助手。"),
    Message(role="user", content="1+1等于多少？"),
    Message(role="assistant", content="1+1等于2。"),
    Message(role="user", content="那乘以2呢？")
]

response = await client.generate(messages)
print("回答:", response.content)
```

### 4.3 使用工具调用

```python
def get_weather(city: str) -> dict:
    """获取指定城市的天气信息"""
    # 模拟 API 调用
    return {"city": city, "temperature": "25°C", "weather": "晴朗"}

# 定义工具
tools = [
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "获取城市的天气信息",
            "parameters": {
                "type": "object",
                "properties": {
                    "city": {
                        "type": "string",
                        "description": "城市名称"
                    }
                },
                "required": ["city"]
            }
        }
    }
]

messages = [
    Message(role="user", content="北京的天气怎么样？")
]

response = await client.generate(messages, tools=tools)

# 检查是否有工具调用
if response.tool_calls:
    for tool_call in response.tool_calls:
        if tool_call.function.name == "get_weather":
            city = tool_call.function.arguments["city"]
            result = get_weather(city)

            # 发送工具结果回模型
            messages.append(
                Message(
                    role="tool",
                    tool_call_id=tool_call.id,
                    content=f"北京天气：{result['temperature']}，{result['weather']}"
                )
            )

    # 再次生成，获取最终回答
    final_response = await client.generate(messages)
    print("最终回答:", final_response.content)
```

### 4.4 推理内容提取

```python
messages = [
    Message(role="user", content="解决这个数学问题：15 * 23 = ?")
]

response = await client.generate(messages)

print("思考过程:", response.thinking)
print("最终答案:", response.content)
```

### 4.5 处理工具调用的完整流程

```python
async def handle_tool_calls(client, messages, tools):
    """处理工具调用的完整流程"""
    response = await client.generate(messages, tools=tools)

    # 如果有工具调用
    if response.tool_calls:
        # 执行工具调用
        for tool_call in response.tool_calls:
            tool_result = await execute_tool(tool_call)

            # 添加工具结果到消息历史
            messages.append(
                Message(
                    role="tool",
                    tool_call_id=tool_call.id,
                    content=str(tool_result)
                )
            )

        # 再次调用模型，获取最终响应
        final_response = await client.generate(messages)
        return final_response

    return response

async def execute_tool(tool_call):
    """执行工具调用"""
    if tool_call.function.name == "your_tool_name":
        # 处理工具逻辑
        return {"result": "tool execution result"}
    return None
```

## 5. 开发指南

### 5.1 架构设计

项目采用 **策略模式** 和 **工厂模式** 的设计：

1. **抽象层**：`LLMClientBase` 定义统一的接口规范
2. **实现层**：`OpenAIClient` 实现具体的提供商逻辑
3. **包装层**：`LLMClient` 根据 provider 参数选择合适的客户端

### 5.2 添加新的 LLM 提供商

要添加新的提供商，需要：

1. 在 `LLMProvider` 枚举中添加新的提供商类型

```python
# src/schema/schema.py
class LLMProvider(str, Enum):
    OPENAI = "openai"
    ANTHROPIC = "anthropic"  # 新增
```

2. 创建新的客户端类，继承 `LLMClientBase`

```python
# src/llm/anthropic_client.py
from src.llm.base import LLMClientBase
from src.schema import LLMResponse, Message

class AnthropicClient(LLMClientBase):
    def __init__(self, api_key: str, api_base: str, model: str):
        super().__init__(api_key, api_base, model)

    async def generate(
        self,
        messages: list[Message],
        tools: list[Any] | None = None,
    ) -> LLMResponse:
        # 实现 Anthropic 特定的逻辑
        pass

    def _prepare_request(self, ...):
        # 实现请求准备
        pass

    def _convert_messages(self, messages: list[Message]):
        # 实现消息转换
        pass
```

3. 在 `LLMClient` 中添加对新提供商的支持

```python
# src/llm/llm_wrapper.py
from src.llm.anthropic_client import AnthropicClient

if provider == LLMProvider.ANTHROPIC:
    self._client = AnthropicClient(
        api_key=api_key,
        api_base=full_api_base,
        model=model,
    )
```

### 5.3 自定义工具调用

可以通过多种方式定义工具：

1. **字典格式**（适合简单工具）：

```python
tools = [
    {
        "type": "function",
        "function": {
            "name": "calculator",
            "description": "执行基本数学运算",
            "parameters": {
                "type": "object",
                "properties": {
                    "expression": {"type": "string"}
                }
            }
        }
    }
]
```

2. **对象格式**（适合复杂工具）：

```python
class CalculatorTool:
    def to_openai_schema(self):
        return {
            "type": "function",
            "function": {
                "name": "calculate",
                "description": "执行数学计算",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "a": {"type": "number"},
                        "b": {"type": "number"},
                        "operation": {"type": "string", "enum": ["+", "-", "*", "/"]}
                    }
                }
            }
        }

tools = [CalculatorTool()]
```

### 5.4 错误处理

```python
try:
    response = await client.generate(messages)
    if response.tool_calls:
        # 处理工具调用
        pass
except Exception as e:
    print(f"请求失败: {e}")
    # 处理错误
```

### 5.5 最佳实践

1. **消息历史管理**：保留完整的对话历史，包括推理内容
2. **工具调用处理**：确保工具调用的 ID 正确传递
3. **资源管理**：使用 `async with` 或确保客户端正确关闭
4. **类型注解**：始终使用类型注解以获得更好的 IDE 支持
5. **错误处理**：为 API 调用添加适当的错误处理和重试逻辑

## 6. API 参考

### 6.1 LLMClient

主要的统一客户端类。

#### 6.1.1 初始化参数

```python
LLMClient(
    api_key: str,                    # API 密钥
    provider: LLMProvider = LLMProvider.OPENAI,  # 提供商类型
    api_base: str = "http://localhost:8080",     # API 基础地址
    model: str = "gpt-oss"           # 模型名称
)
```

#### 6.1.2 方法

**`async generate(messages: list[Message], tools: list | None = None) -> LLMResponse`**

生成 LLM 响应。

- `messages`: 对话消息列表
- `tools`: 可选的工具列表
- 返回：`LLMResponse` 对象

### 6.2 OpenAIClient

OpenAI 提供商的实现类。

#### 6.2.1 初始化参数

```python
OpenAIClient(
    api_key: str,
    api_base: str = "http://localhost:8080/v1",
    model: str = "gpt-oss"
)
```

### 6.3 数据模型

#### 6.3.1 Message

聊天消息模型。

```python
Message(
    role: str,                                    # 角色：system, user, assistant, tool
    content: str | list[dict[str, Any]],          # 消息内容
    thinking: str | None = None,                  # 推理内容（仅 assistant）
    tool_calls: list[ToolCall] | None = None,     # 工具调用列表
    tool_call_id: str | None = None,              # 工具调用 ID
    name: str | None = None                       # 工具名称
)
```

#### 6.3.2 LLMResponse

LLM 响应模型。

```python
LLMResponse(
    content: str,                                 # 文本内容
    thinking: str | None = None,                  # 推理内容
    tool_calls: list[ToolCall] | None = None,     # 工具调用列表
    finish_reason: str                            # 完成原因
)
```

#### 6.3.3 ToolCall

工具调用模型。

```python
ToolCall(
    id: str,                                      # 调用 ID
    type: str,                                    # 类型（通常为 "function"）
    function: FunctionCall                        # 函数调用详情
)
```

#### 6.3.4 FunctionCall

函数调用模型。

```python
FunctionCall(
    name: str,                                    # 函数名称
    arguments: dict[str, Any]                     # 函数参数
)
```

#### 6.3.5 LLMProvider

LLM 提供商枚举。

```python
LLMProvider.OPENAI  # OpenAI 提供商
```

### 6.4 常用用例

#### 6.4.1 初始化 OpenAI 客户端

```python
from src import LLMClient, LLMProvider

client = LLMClient(
    api_key="your-api-key",
    provider=LLMProvider.OPENAI,
    api_base="http://localhost:8080",
    model="gpt-oss"
)
```

#### 6.4.2 创建系统消息

```python
from src import Message

system_msg = Message(
    role="system",
    content="你是一个有用的AI助手。"
)
```

#### 6.4.3 创建用户消息

```python
user_msg = Message(
    role="user",
    content="请解释一下什么是机器学习。"
)
```

#### 6.4.4 处理响应

```python
response = await client.generate(messages)

# 访问文本内容
print(response.content)

# 访问推理内容
if response.thinking:
    print("思考过程:", response.thinking)

# 访问工具调用
if response.tool_calls:
    for tool_call in response.tool_calls:
        print(f"调用工具: {tool_call.function.name}")
```

### 6.5 注意事项

1. **API 密钥安全**：不要将 API 密钥硬编码在代码中，使用环境变量
2. **并发请求**：`LLMClient` 是线程安全的，可以同时处理多个请求
3. **资源释放**：如果使用大量客户端实例，考虑实现上下文管理器
4. **版本兼容性**：确保 OpenAI SDK 版本兼容（>=2.8.1）
5. **端点后缀**：库会自动为 API base 添加 `/v1` 后缀，无需手动添加

---

## 总结

Gopilot-CLI 提供了简洁、统一的 LLM 客户端接口，支持 OpenAI 提供商、工具调用和推理内容提取。通过本开发文档，您可以快速上手并在项目中集成和使用该库。如有问题，请参考示例代码或提交 Issue。

**许可证**：请根据项目实际情况添加许可证信息。

**贡献**：欢迎提交 Pull Request 来改进项目！

**更新日志**：请查看项目中的 CHANGELOG.md 文件了解版本更新信息。
