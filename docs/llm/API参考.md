# API 参考文档

完整的 API 文档，包含所有类、方法、参数、返回值的详细说明。

## 📋 目录

- [LLMClient 工厂类](#llmclient-工厂类)
- [LLMClientBase 抽象基类](#llmclientbase-抽象基类)
- [OpenAIClient 客户端](#openaiclient-客户端)
- [Message 对象](#message-对象)
- [LLMResponse 对象](#llmresponse-对象)
- [LLMProvider 枚举](#llmprovider-枚举)
- [ToolCall 对象](#toolcall-对象)
- [FunctionCall 对象](#functioncall-对象)
- [异常类型](#异常类型)

---

## LLMClient 工厂类

LLM 模块的推荐入口，工厂模式实现，根据提供商自动选择和实例化对应的客户端。

### 类定义

```python
class LLMClient:
    """LLM 工厂类，统一客户端入口"""
```

### 构造函数

#### `__init__()`

**定义**：
```python
def __init__(
    self,
    api_key: str,
    provider: LLMProvider = LLMProvider.OPENAI,
    api_base: str | None = None,
    model: str = "gpt-oss"
) -> None
```

**参数**：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `api_key` | `str` | - | **必需**。API 密钥，用于身份验证 |
| `provider` | `LLMProvider` | `LLMProvider.OPENAI` | 提供商类型，当前仅支持 `OPENAI` |
| `api_base` | `str \| None` | `None` | API 基础 URL。为 `None` 时使用默认值：<br/>- OpenAI: `"http://localhost:8080"` |
| `model` | `str` | `"gpt-oss"` | 模型名称，需与实际部署的模型匹配 |

**行为**：
1. 根据 `provider` 设置默认 API 地址（如果 `api_base` 为 `None`）
2. 自动添加 OpenAI API 路径后缀 `/v1`（如果未包含）
3. 验证提供商类型，不支持则抛出 `ValueError`
4. 实例化对应的具体客户端

**示例**：
```python
# 使用默认配置（连接本地 llama.cpp）
client = LLMClient(
    api_key="sk-1234567890",
    provider=LLMProvider.OPENAI
)
# 等价于：
# api_base="http://localhost:8080/v1", model="gpt-oss"

# 显式设置所有参数
client = LLMClient(
    api_key="sk-your-key",
    provider=LLMProvider.OPENAI,
    api_base="https://api.openai.com/v1",
    model="gpt-4o-mini"
)

# 连接本地 llama.cpp
client = LLMClient(
    api_key="sk-1234567890",
    provider=LLMProvider.OPENAI,
    api_base="http://localhost:8080",
    model="your-model-name"
)
```

### 实例方法

#### `generate()`

**定义**：
```python
async def generate(
    self,
    messages: list[Message],
    tools: list[Any] | None = None
) -> LLMResponse
```

**参数**：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `messages` | `list[Message]` | - | **必需**。对话消息列表，至少包含一个用户消息 |
| `tools` | `list[Any] \| None` | `None` | 可选。工具/函数调用列表，用于 Function Calling |

**返回值**：`LLMResponse` - LLM 生成的响应对象

**功能**：生成 LLM 响应

**调用流程**：
1. 将调用委托给内部的具体客户端实例
2. 具体客户端执行：
   - 转换消息格式
   - 准备请求参数
   - 执行 API 调用
   - 解析响应

**异常**：
- `ValueError`: 如果 `messages` 为空
- `ConnectionError`: 如果无法连接到 LLM 服务
- 其他提供商特定的异常

**示例**：
```python
import asyncio
from src.llm import LLMClient, LLMProvider
from src.schema import Message

async def example():
    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    # 基本对话
    messages = [
        Message(role="user", content="你好，请介绍一下自己")
    ]
    response = await client.generate(messages)

    print(response.content)
    print(f"完成原因: {response.finish_reason}")

    # 工具调用
    tools = [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "获取天气",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "city": {"type": "string"}
                    }
                }
            }
        }
    ]

    messages_with_tools = [
        Message(role="user", content="北京天气怎么样？")
    ]
    response = await client.generate(messages_with_tools, tools=tools)

    if response.tool_calls:
        for tool_call in response.tool_calls:
            print(f"调用工具: {tool_call.function.name}")
            print(f"参数: {tool_call.function.arguments}")

asyncio.run(example())
```

---

## LLMClientBase 抽象基类

所有 LLM 客户端的抽象基类，定义了必须实现的接口。

### 类定义

```python
class LLMClientBase:
    """LLM 客户端抽象基类"""
```

### 构造函数

#### `__init__()`

**定义**：
```python
def __init__(
    self,
    api_key: str,
    api_base: str,
    model: str
) -> None
```

**参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `api_key` | `str` | API 密钥 |
| `api_base` | `str` | API 基础 URL |
| `model` | `str` | 模型名称 |

**属性**：

| 属性 | 类型 | 说明 |
|------|------|------|
| `api_key` | `str` | API 密钥 |
| `api_base` | `str` | API 基础 URL |
| `model` | `str` | 模型名称 |

### 抽象方法

#### `generate()`

**定义**：
```python
async def generate(
    self,
    messages: list[Message],
    tools: list[Any] | None = None
) -> LLMResponse
```

**说明**：生成 LLM 响应的抽象方法，子类必须实现

**子类实现要求**：
- 转换消息格式
- 准备请求参数
- 执行 API 调用
- 解析响应
- 返回 `LLMResponse` 对象

#### `_prepare_request()`

**定义**：
```python
def _prepare_request(
    self,
    messages: list[Message],
    tools: list[Any] | None = None
) -> dict[str, Any]
```

**说明**：准备 API 请求参数字典

**返回**：`dict[str, Any]` - 请求参数字典，包含 `model`、`messages` 等

#### `_convert_messages()`

**定义**：
```python
def _convert_messages(
    self,
    messages: list[Message]
) -> tuple[str | None, list[dict[str, Any]]]
```

**说明**：转换消息格式为 API 特定格式

**返回**：`tuple[str | None, list[dict[str, Any]]]`
- 第一个元素：系统消息（如果存在）
- 第二个元素：转换后的消息列表

---

## OpenAIClient 客户端

OpenAI 协议的具体实现，兼容 llama.cpp 等本地部署。

### 类定义

```python
class OpenAIClient(LLMClientBase):
    """OpenAI 兼容客户端"""
```

### 构造函数

#### `__init__()`

**定义**：
```python
def __init__(
    self,
    api_key: str,
    api_base: str,
    model: str
) -> None
```

**参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `api_key` | `str` | API 密钥 |
| `api_base` | `str` | API 基础 URL（**必须**包含 `/v1` 后缀） |
| `model` | `str` | 模型名称 |

**行为**：
1. 调用父类初始化
2. 创建 `AsyncOpenAI` 客户端实例
3. 存储客户端引用以便后续调用

**注意**：与 `LLMClient` 不同，`OpenAIClient` 需要手动包含 `/v1` 后缀

**示例**：
```python
# ✅ 正确：包含 /v1 后缀
client = OpenAIClient(
    api_key="sk-1234567890",
    api_base="http://localhost:8080/v1",
    model="gpt-oss"
)

# ❌ 错误：缺少 /v1 后缀
# client = OpenAIClient(
#     api_key="sk-1234567890",
#     api_base="http://localhost:8080",
#     model="gpt-oss"
# )
```

### 实例方法

#### `_make_api_request()`

**定义**：
```python
async def _make_api_request(
    self,
    messages: list[dict[str, Any]],
    tools: list[Any] | None
) -> ChatCompletion
```

**参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `messages` | `list[dict[str, Any]]` | OpenAI 格式的消息列表 |
| `tools` | `list[Any] \| None` | 工具列表（可选） |

**返回值**：`ChatCompletion` - OpenAI API 响应对象

**功能**：执行核心的 API 调用

**关键特性**：
- 设置 `reasoning_split=True` 启用思考内容支持
- 自动添加工具调用参数（`tools`、`tool_choice`）
- 调用 `openai.chat.completions.create()`

**内部实现**：
```python
async def _make_api_request(self, messages, tools=None):
    kwargs = {
        "model": self.model,
        "messages": messages,
        "reasoning_split": True,  # 启用思考内容
    }

    if tools:
        kwargs["tools"] = tools
        kwargs["tool_choice"] = "auto"

    response = await self.client.chat.completions.create(**kwargs)
    return response
```

#### `_convert_tools()`

**定义**：
```python
def _convert_tools(
    self,
    tools: list[Any]
) -> list[dict[str, Any]]
```

**参数**：
| 参数 | 类型 | 说明 |
|------|------|------|
| `tools` | `list[Any]` | 工具列表 |

**返回值**：`list[dict[str, Any]]` - OpenAI 格式的工具列表

**功能**：转换工具格式为 OpenAI 格式

**支持的格式**：

1. **OpenAI 原生格式**（直接通过）：
```python
tools = [
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "获取天气",
            "parameters": {...}
        }
    }
]
```

2. **自定义对象**（调用 `to_openai_schema()`）：
```python
class WeatherTool:
    def to_openai_schema(self):
        return {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "获取天气",
                "parameters": {...}
            }
        }

tools = [WeatherTool()]
```

3. **Anthropic 格式**（自动转换）：
```python
tools = [
    {
        "name": "get_weather",
        "description": "获取天气",
        "input_schema": {...}
    }
]
# 自动转换为 OpenAI 格式
```

**异常**：
- `ValueError`: 如果工具格式不支持

#### `_convert_messages()`

**定义**：
```python
def _convert_messages(
    self,
    messages: list[Message]
) -> tuple[str | None, list[dict[str, Any]]]
```

**参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `messages` | `list[Message]` | 内部格式的消息列表 |

**返回值**：`tuple[str | None, list[dict[str, Any]]]`
- 系统消息字符串（如果有）
- OpenAI 格式的消息列表

**转换规则**：

| 输入格式 | 输出格式 | 特殊处理 |
|---------|---------|----------|
| `Message(role="system")` | 单独提取 | 存储为第一个返回值 |
| `Message(role="user")` | `{"role": "user", "content": ...}` | 直接映射 |
| `Message(role="assistant")` | `{"role": "assistant", "content": ..., "reasoning_content": ..., "tool_calls": ...}` | 处理 `thinking` 和 `tool_calls` |
| `Message(role="tool")` | `{"role": "tool", "content": ..., "tool_call_id": ..., "name": ...}` | 工具消息特殊处理 |

**示例**：
```python
messages = [
    Message(role="system", content="你是一个助手"),
    Message(
        role="assistant",
        content="你好",
        thinking="用户问候",
        tool_calls=[...]
    )
]

system_msg, api_msgs = client._convert_messages(messages)

# system_msg = "你是一个助手"
# api_msgs = [
#     {
#         "role": "assistant",
#         "content": "你好",
#         "reasoning_content": "用户问候",
#         "tool_calls": [...]
#     }
# ]
```

#### `_parse_response()`

**定义**：
```python
def _parse_response(
    self,
    response: ChatCompletion
) -> LLMResponse
```

**参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `response` | `ChatCompletion` | OpenAI API 响应对象 |

**返回值**：`LLMResponse` - 解析后的响应对象

**功能**：解析 OpenAI API 响应，提取内容、思考内容、工具调用等

**提取字段**：
- `content`: `response.choices[0].message.content`
- `thinking`: `response.choices[0].message.reasoning_content`（如果存在）
- `tool_calls`: 解析 `response.choices[0].message.tool_calls`（如果存在）
- `finish_reason`: `response.choices[0].finish_reason`

#### `generate()`

**定义**：
```python
async def generate(
    self,
    messages: list[Message],
    tools: list[Any] | None = None
) -> LLMResponse
```

**参数**：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `messages` | `list[Message]` | - | **必需**。对话消息列表 |
| `tools` | `list[Any] \| None` | `None` | 可选。工具列表 |

**返回值**：`LLMResponse` - 解析后的响应对象

**完整流程**：

```python
async def generate(self, messages, tools=None):
    # 1. 转换消息格式
    system_message, api_messages = self._convert_messages(messages)

    # 2. 准备请求参数（如果需要可以重写）
    request_data = self._prepare_request(api_messages, tools)

    # 3. 执行 API 调用
    response = await self._make_api_request(api_messages, tools)

    # 4. 解析响应
    result = self._parse_response(response)

    return result
```

---

## Message 对象

对话消息的数据模型，定义在 `src/schema.py` 中。

### 类定义

```python
class Message(BaseModel):
    """对话消息数据模型"""
```

### 字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `role` | `str` | - | **必需**。消息角色，取值：`"system"`、`"user"`、`"assistant"`、`"tool"` |
| `content` | `str \| list[dict]` | - | **必需**。消息内容，文本或多模态数据 |
| `thinking` | `str \| None` | `None` | 可选。思考内容（仅适用于 `assistant` 角色） |
| `tool_calls` | `list[ToolCall] \| None` | `None` | 可选。工具调用列表（仅适用于 `assistant` 角色） |
| `tool_call_id` | `str \| None` | `None` | 可选。工具调用 ID（仅适用于 `tool` 角色） |
| `name` | `str \| None` | `None` | 可选。工具名称（仅适用于 `tool` 角色） |

### 消息角色

#### `role="system"`

**用途**：设置 AI 的系统级指令和行为

**示例**：
```python
Message(
    role="system",
    content="你是一个专业的 Python 程序员，擅长编写优雅、高效的代码。"
)
```

**特点**：
- 可以在对话开始前设置 AI 的角色
- 通常作为第一条消息
- 不参与实际对话，但影响 AI 的行为

#### `role="user"`

**用途**：用户的输入消息

**示例**：
```python
Message(
    role="user",
    content="请写一个快速排序算法"
)
```

#### `role="assistant"`

**用途**：AI 的回复消息

**包含思考内容**：
```python
Message(
    role="assistant",
    content="以下是快速排序算法：",
    thinking="用户要求写快速排序，我需要提供一个完整的实现。"
)
```

**包含工具调用**：
```python
Message(
    role="assistant",
    content="我需要查询一下天气信息。",
    tool_calls=[
        ToolCall(
            id="call_123",
            type="function",
            function=FunctionCall(
                name="get_weather",
                arguments='{"city": "北京"}'
            )
        )
    ]
)
```

#### `role="tool"`

**用途**：工具执行的结果消息

**示例**：
```python
Message(
    role="tool",
    content='{"weather": "晴天", "temperature": 25}',
    tool_call_id="call_123",
    name="get_weather"
)
```

### 使用示例

#### 基本对话

```python
messages = [
    Message(role="user", content="你好")
]
```

#### 多轮对话

```python
messages = [
    Message(role="user", content="什么是机器学习？"),
    Message(role="assistant", content="机器学习是..."),
    Message(role="user", content="能详细说说吗？")
]
```

#### 带系统消息的对话

```python
messages = [
    Message(role="system", content="你是一个有用的助手"),
    Message(role="user", content="你好")
]
```

#### 完整的对话流

```python
messages = [
    # 系统消息
    Message(role="system", content="你是一个专业的 Python 程序员"),

    # 第1轮
    Message(role="user", content="请写一个计算斐波那契数列的函数"),

    # AI 回复（包含思考）
    Message(
        role="assistant",
        content="```python\ndef fibonacci(n):\n    ...\n```",
        thinking="用户要求写斐波那契函数，我应该提供一个递归或迭代的实现。"
    ),

    # 用户追问
    Message(role="user", content="能优化一下性能吗？"),

    # AI 回复（可能触发工具调用）
    Message(
        role="assistant",
        content="可以使用动态规划来优化...",
        tool_calls=[...]
    ),

    # 工具结果
    Message(
        role="tool",
        content='{"result": "优化后的代码"}',
        tool_call_id="call_123",
        name="optimize_code"
    )
]
```

---

## LLMResponse 对象

LLM 响应的数据模型。

### 类定义

```python
class LLMResponse(BaseModel):
    """LLM 响应数据模型"""
```

### 字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `content` | `str` | - | **必需**。AI 生成的文本内容 |
| `thinking` | `str \| None` | `None` | 可选。思考内容（reasoning_content），如果模型支持 |
| `tool_calls` | `list[ToolCall] \| None` | `None` | 可选。工具调用列表 |
| `finish_reason` | `str` | - | **必需**。完成原因，取值：`"stop"`、`"tool_calls"`、`"length"` 等 |

### 完成原因（finish_reason）

| 值 | 说明 |
|----|------|
| `"stop"` | 正常结束，AI 认为已完成回答 |
| `"tool_calls"` | 需要调用工具，工具调用完成后继续 |
| `"length"` | 达到最大长度限制 |
| `"content_filter"` | 内容被过滤 |
| `"function_call"` | 调用了函数（较老版本） |

### 使用示例

#### 基本响应

```python
response = await client.generate([Message(role="user", content="你好")])

print(response.content)  # "你好！很高兴为您服务！"
print(response.finish_reason)  # "stop"
```

#### 响应包含思考内容

```python
response = await client.generate([
    Message(role="user", content="请逐步思考 15 * 23")
])

if response.thinking:
    print("思考过程:")
    print(response.thinking)

print("最终答案:")
print(response.content)
```

**输出示例**：
```
思考过程:
让我逐步计算：
15 * 23 = 15 * (20 + 3) = 15*20 + 15*3 = 300 + 45 = 345

最终答案:
345
```

#### 响应包含工具调用

```python
response = await client.generate(
    [Message(role="user", content="北京天气怎么样？")],
    tools=[...]
)

if response.tool_calls:
    print("需要调用工具:")
    for tool_call in response.tool_calls:
        print(f"  工具: {tool_call.function.name}")
        print(f"  参数: {tool_call.function.arguments}")

        # 执行工具
        result = execute_tool(tool_call.function.name, tool_call.function.arguments)

        # 添加工具结果到对话
        messages.append(Message(
            role="tool",
            content=result,
            tool_call_id=tool_call.id,
            name=tool_call.function.name
        ))
```

---

## LLMProvider 枚举

LLM 提供商类型枚举。

### 定义

```python
class LLMProvider(str, Enum):
    """LLM 提供商枚举"""
    OPENAI = "openai"
```

### 当前支持的提供商

| 值 | 说明 |
|----|------|
| `"openai"` | OpenAI 兼容提供商（llama.cpp、本地部署、OpenAI API） |

### 使用示例

```python
from src.schema import LLMProvider

# 显式指定提供商
client = LLMClient(
    api_key="your-key",
    provider=LLMProvider.OPENAI
)

# 默认即为 OpenAI
client = LLMClient(
    api_key="your-key",
    # provider=LLMProvider.OPENAI  # 可省略
)
```

---

## ToolCall 对象

工具调用的数据模型。

### 类定义

```python
class ToolCall(BaseModel):
    """工具调用数据模型"""
```

### 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `str` | 工具调用的唯一 ID |
| `type` | `str` | 工具类型，通常为 `"function"` |
| `function` | `FunctionCall` | 工具调用详情 |

### 使用示例

```python
# 解析工具调用
if response.tool_calls:
    for tool_call in response.tool_calls:
        # 访问工具 ID
        print(f"工具调用 ID: {tool_call.id}")

        # 访问工具类型
        print(f"工具类型: {tool_call.type}")

        # 访问工具函数
        print(f"函数名: {tool_call.function.name}")
        print(f"参数: {tool_call.function.arguments}")
```

---

## FunctionCall 对象

函数调用的数据模型。

### 类定义

```python
class FunctionCall(BaseModel):
    """函数调用数据模型"""
```

### 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | `str` | 函数的名称 |
| `arguments` | `str` | 函数的参数（JSON 字符串格式） |

### 使用示例

```python
# 解析函数调用
tool_call = response.tool_calls[0]

function_name = tool_call.function.name
function_args_str = tool_call.function.arguments

# 解析 JSON 参数
import json
try:
    function_args = json.loads(function_args_str)
    print(f"解析后的参数: {function_args}")

    # 执行函数
    result = execute_function(function_name, **function_args)
except json.JSONDecodeError:
    print(f"参数解析失败: {function_args_str}")
```

**完整示例**：
```python
async def handle_tool_call(tool_call: ToolCall):
    """处理工具调用的完整示例"""

    # 1. 获取函数名
    function_name = tool_call.function.name

    # 2. 解析参数
    try:
        function_args = json.loads(tool_call.function.arguments)
    except json.JSONDecodeError as e:
        return f"参数解析错误: {e}"

    # 3. 执行函数
    if function_name == "get_weather":
        # 假设执行天气查询
        city = function_args.get("city")
        weather_data = get_weather_from_api(city)

        return f"{city}今天{weather_data['weather']}，{weather_data['temperature']}度"

    elif function_name == "calculate":
        # 假设执行计算
        expression = function_args.get("expression")
        result = eval(expression)  # 注意：实际使用中要避免 eval

        return f"计算结果: {result}"

    else:
        return f"未知的工具: {function_name}"
```

---

## 异常类型

LLM 模块可能抛出的异常类型。

### ConnectionError

**类型**：Python 内置异常 `ConnectionError`

**触发条件**：
- 无法连接到 LLM 服务
- 网络超时
- 服务不可用

**示例**：
```python
try:
    response = await client.generate([Message(role="user", content="你好")])
except ConnectionError as e:
    print(f"连接失败: {e}")
    print("请检查：")
    print("1. API 地址是否正确")
    print("2. LLM 服务是否已启动")
    print("3. 网络连接是否正常")
```

### AuthenticationError

**类型**：OpenAI 库异常（通常为 `openai.AuthenticationError`）

**触发条件**：
- API 密钥无效
- API 密钥已过期
- API 密钥权限不足

**示例**：
```python
try:
    response = await client.generate([Message(role="user", content="你好")])
except Exception as e:
    if "authentication" in str(e).lower():
        print("认证失败: API 密钥无效")
        print("请检查 API 密钥是否正确")
```

### ModelNotFoundError

**类型**：OpenAI 库异常（通常为 `openai.NotFoundError`）

**触发条件**：
- 模型名称不存在
- 模型未加载
- 模型名称拼写错误

**示例**：
```python
try:
    client = LLMClient(
        api_key="your-key",
        model="non-existent-model"  # 不存在的模型
    )
    response = await client.generate([Message(role="user", content="你好")])
except Exception as e:
    if "not found" in str(e).lower():
        print("模型不存在，请检查模型名称")
```

### ValueError

**类型**：Python 内置异常 `ValueError`

**触发条件**：
- 提供商类型不支持
- 工具格式错误
- 消息为空

**示例**：
```python
# 提供商不支持
try:
    client = LLMClient(
        api_key="your-key",
        provider="unsupported_provider"  # 类型错误
    )
except ValueError as e:
    print(f"参数错误: {e}")

# 消息为空
try:
    response = await client.generate([])  # 空消息列表
except ValueError as e:
    print(f"消息不能为空: {e}")
```

### 异常处理最佳实践

```python
import asyncio
from typing import Optional

async def robust_generate(
    client: LLMClient,
    messages: list[Message],
    tools: Optional[list] = None,
    max_retries: int = 3
) -> Optional[LLMResponse]:
    """带重试和错误处理的生成方法"""

    for attempt in range(max_retries):
        try:
            response = await client.generate(messages, tools)
            return response

        except ConnectionError as e:
            print(f"连接失败 (第 {attempt + 1} 次尝试): {e}")
            if attempt == max_retries - 1:
                print("达到最大重试次数，连接失败")
                return None

            # 指数退避
            await asyncio.sleep(2 ** attempt)
            continue

        except Exception as e:
            print(f"未知错误: {type(e).__name__}: {e}")
            return None

    return None
```

---

## 完整使用示例

### 示例 1：基本对话

```python
import asyncio
from src.llm import LLMClient, LLMProvider
from src.schema import Message

async def basic_chat_example():
    """基本对话示例"""

    # 创建客户端
    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    # 发送消息
    messages = [
        Message(role="user", content="你好，请介绍一下自己")
    ]

    # 生成响应
    response = await client.generate(messages)

    print("🤖 AI 回复:")
    print(response.content)
    print(f"\n完成原因: {response.finish_reason}")

asyncio.run(basic_chat_example())
```

### 示例 2：多轮对话

```python
async def multi_turn_example():
    """多轮对话示例"""

    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI
    )

    # 对话历史
    conversation = [
        Message(role="user", content="什么是 Python？"),
    ]

    # 第1轮
    response1 = await client.generate(conversation)
    print(f"AI: {response1.content}\n")

    # 添加 AI 回复到历史
    conversation.append(Message(
        role="assistant",
        content=response1.content
    ))

    # 第2轮
    conversation.append(Message(
        role="user",
        content="它有哪些优势？"
    ))
    response2 = await client.generate(conversation)
    print(f"AI: {response2.content}")

asyncio.run(multi_turn_example())
```

### 示例 3：工具调用

```python
async def tool_calling_example():
    """工具调用示例"""

    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI
    )

    # 定义工具
    tools = [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "获取指定城市的天气信息",
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

    # 发送消息
    messages = [
        Message(role="user", content="查询一下北京的天气")
    ]

    response = await client.generate(messages, tools=tools)

    # 检查工具调用
    if response.tool_calls:
        print("🔧 触发工具调用:")
        for tool_call in response.tool_calls:
            print(f"  工具: {tool_call.function.name}")
            print(f"  参数: {tool_call.function.arguments}")

            # 模拟执行工具
            import json
            args = json.loads(tool_call.function.arguments)
            # result = execute_tool(tool_call.function.name, args)
    else:
        print("🤖 直接回复:")
        print(response.content)

asyncio.run(tool_calling_example())
```

### 示例 4：思考内容

```python
async def thinking_example():
    """思考内容示例"""

    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI
    )

    messages = [
        Message(
            role="user",
            content="请逐步思考并计算：12 * 15 + 23 = ?"
        )
    ]

    response = await client.generate(messages)

    # 打印思考过程
    if response.thinking:
        print("🧠 思考过程:")
        print(response.thinking)
        print("\n" + "="*50 + "\n")

    # 打印最终答案
    print("💡 最终答案:")
    print(response.content)

asyncio.run(thinking_example())
```

### 示例 5：系统消息（设置角色）

```python
async def system_message_example():
    """系统消息设置 AI 角色"""

    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI
    )

    messages = [
        # 系统消息：设置 AI 的角色
        Message(
            role="system",
            content="你是一个专业的 Python 程序员，擅长编写优雅、高效的代码。"
        ),
        # 用户消息：提出请求
        Message(
            role="user",
            content="请写一个装饰器来计算函数执行时间"
        )
    ]

    response = await client.generate(messages)

    print("🤖 Python 程序员回答:")
    print(response.content)

asyncio.run(system_message_example())
```

---

**上一篇：** [架构设计](./架构设计.md)
**下一篇：** [开发者指南](./开发者指南.md)
