# LLM 模块开发文档

欢迎使用 gopilot-cli LLM 模块开发文档！

## 📚 文档概览

本文档集提供了 `src/llm` 模块的完整开发指南，包括架构设计、API 参考、使用教程和扩展指南。

### 什么是 LLM 模块？

LLM 模块是 gopilot-cli 的核心组件之一，提供了一套统一的接口用于与大语言模型（LLM）进行交互。模块采用三层架构设计，支持多种 LLM 提供商，兼容 OpenAI 协议，特别针对 llama.cpp 等本地部署场景进行了优化。

### 核心特性

- ✅ **统一接口**：抽象基类设计，提供一致的 API 调用体验
- ✅ **多提供商支持**：工厂模式轻松切换不同 LLM 提供商
- ✅ **OpenAI 兼容**：完全兼容 OpenAI API 协议，可用于 llama.cpp
- ✅ **思考内容支持**：内置 reasoning_content 支持，实现链式思考
- ✅ **工具调用**：完整的 Function Calling 功能实现
- ✅ **异步执行**：基于 asyncio 的高性能异步调用
- ✅ **多轮对话**：自动管理对话历史和消息状态
- ✅ **强类型**：基于 Pydantic 的数据模型，确保类型安全
- ✅ **易于扩展**：清晰的抽象设计，方便添加新提供商

## 📖 文档导航

### [🚀 快速入门](./快速入门.md)
适合刚开始使用 LLM 模块的开发者。涵盖：
- 环境准备和安装
- 基本使用示例（初始化、基本对话）
- 高级功能（工具调用、思考内容）
- 常见使用场景（本地部署、云端 API）
- 5 分钟上手教程

**推荐阅读顺序：第一篇**

---

### [🏗️ 架构设计](./架构设计.md)
深入了解 LLM 模块的内部设计。包括：
- 模块架构图（三层架构设计）
- 设计模式和设计原则
- 核心组件说明（LLMClientBase、OpenAIClient、LLMClient）
- 模块依赖关系
- 思考内容和工具调用机制

**推荐阅读顺序：第二篇**

---

### [📋 API 参考](./API参考.md)
完整的 API 文档，包含所有类和方法的详细说明。涵盖：
- LLMClientBase 抽象基类 API
- OpenAIClient 完整 API
- LLMClient 工厂类 API
- Message 对象详解
- LLMResponse 对象详解
- 参数、返回值、异常说明
- 丰富的代码示例

**推荐阅读顺序：第三篇（作为参考手册）**

---

### [👨‍💻 开发者指南](./开发者指南.md)
学习如何扩展和自定义 LLM 模块。包括：
- 添加新的 LLM 提供商
- 自定义消息转换器
- 自定义工具格式
- 最佳实践和设计模式
- 错误处理策略
- 贡献代码规范

**推荐阅读顺序：第四篇（进阶内容）**

---

## 🎯 快速链接

### 常见任务

| 任务 | 文档链接 |
|------|---------|
| 初始化客户端 | [快速入门 - 基本使用](./快速入门.md#基本使用) |
| 发送对话消息 | [快速入门 - 基本对话](./快速入门.md#基本对话) |
| 启用思考模式 | [快速入门 - 思考内容](./快速入门.md#思考内容) |
| 使用工具调用 | [快速入门 - 工具调用](./快速入门.md#工具调用) |
| 管理多轮对话 | [快速入门 - 多轮对话](./快速入门.md#多轮对话) |
| 添加新提供商 | [开发者指南 - 扩展 LLM 提供商](./开发者指南.md#扩展-llm-提供商) |

### API 速查

| 类/方法 | 说明 | 文档链接 |
|---------|------|---------|
| `LLMClient` | 统一的 LLM 客户端入口（推荐） | [API 参考 - LLMClient](./API参考.md#llmclient-工厂类) |
| `OpenAIClient` | OpenAI 兼容客户端 | [API 参考 - OpenAIClient](./API参考.md#openaiclient) |
| `LLMClientBase` | 抽象基类 | [API 参考 - LLMClientBase](./API参考.md#llmclientbase-抽象基类) |
| `LLMClient.generate()` | 生成 LLM 响应 | [API 参考 - generate()](./API参考.md#generatemessages-tools) |
| `Message` | 消息数据模型 | [API 参考 - Message](./API参考.md#message-对象) |
| `LLMResponse` | 响应数据模型 | [API 参考 - LLMResponse](./API参考.md#llmresponse-对象) |

## 🔧 模块结构

```
src/llm/
├── __init__.py              # 模块初始化和导出
│   ├── LLMClientBase       # 抽象基类
│   ├── OpenAIClient        # OpenAI 兼容客户端
│   ├── LLMClient           # 工厂类（推荐使用）
│   └── LLMProvider         # 提供商枚举
│
├── base.py                  # 抽象基类定义
│   └── LLMClientBase       # 定义标准接口
│       ├── __init__()      # 初始化
│       ├── generate()      # 生成响应（抽象）
│       ├── _prepare_request()  # 准备请求（抽象）
│       └── _convert_messages() # 转换消息（抽象）
│
├── openai_client.py         # OpenAI 协议实现
│   └── OpenAIClient         # 继承自 LLMClientBase
│       ├── __init__()      # 初始化 AsyncOpenAI 客户端
│       ├── _make_api_request()  # 执行 API 请求
│       ├── _convert_messages()  # 消息格式转换
│       ├── _convert_tools()     # 工具格式转换
│       ├── _parse_response()    # 响应解析
│       └── generate()           # 生成响应
│
└── llm_wrapper.py           # 工厂类和包装器
    └── LLMClient            # 工厂模式实现
        ├── __init__()      # 根据 provider 自动实例化
        └── generate()      # 委托给具体客户端
```

## 💡 使用示例

### 快速示例：基本对话

```python
from src.llm import LLMClient
from src.schema import Message

async def basic_chat():
    # 创建客户端（使用 llama.cpp 默认配置）
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

    print(response.content)  # AI 的回复

# 运行
# asyncio.run(basic_chat())
```

### 快速示例：多轮对话

```python
async def multi_turn_chat():
    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    # 多轮对话历史
    messages = [
        Message(role="user", content="什么是机器学习？"),
        Message(role="assistant", content="机器学习是..."),
        Message(role="user", content="能详细说说监督学习吗？")
    ]

    response = await client.generate(messages)

    print(response.content)  # 关于监督学习的详细解释
```

### 快速示例：工具调用

```python
async def tool_calling_example():
    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    # 定义工具
    tools = [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "获取天气信息",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "city": {"type": "string"}
                    }
                }
            }
        }
    ]

    messages = [
        Message(role="user", content="北京今天天气怎么样？")
    ]

    response = await client.generate(messages, tools=tools)

    # 检查工具调用
    if response.tool_calls:
        for tool_call in response.tool_calls:
            print(f"调用工具: {tool_call.function.name}")
            print(f"参数: {tool_call.function.arguments}")
```

### 快速示例：思考内容

```python
async def thinking_example():
    client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    messages = [
        Message(
            role="user",
            content="请逐步思考并计算: 15 * 23 = ?"
        )
    ]

    response = await client.generate(messages)

    print(f"思考过程: {response.thinking}")
    print(f"最终答案: {response.content}")

    # 保留思考内容到对话历史
    messages.append(
        Message(
            role="assistant",
            content=response.content,
            thinking=response.thinking
        )
    )
```

## 📦 依赖

LLM 模块依赖以下 Python 库：

- **Python 3.12+**
- **asyncio**：异步 I/O 支持（标准库）
- **typing**：类型注解（标准库）
- **openai**：OpenAI 兼容客户端（第三方库）
- **pydantic**：数据验证和模型定义
- **src.schema**：自定义数据模型（项目内）

其他标准库依赖：
- `json`：JSON 数据处理
- `typing`：类型系统

## 🎨 支持的提供商

当前支持：

| 提供商 | 状态 | 说明 |
|--------|------|------|
| OpenAI | ✅ 支持 | 完全兼容 OpenAI API，适用于 llama.cpp 等本地部署 |

未来计划支持：
- Anthropic Claude
- Google Gemini
- 本地模型（Hugging Face Transformers）

## 🧪 测试

LLM 模块拥有完整的测试覆盖：

```bash
# 运行所有测试
uv run pytest tests/test_*.py -v

# 运行特定测试
uv run pytest tests/test_basic_chat.py -v
uv run pytest tests/test_thinking.py -v

# 查看测试覆盖率
uv run pytest tests/ --cov=src.llm --cov-report=term-missing
```

详见：[开发者指南 - 测试](./开发者指南.md#测试指南)

## 🤝 贡献

欢迎贡献代码！请参阅 [开发者指南](./开发者指南.md) 了解：
- 代码规范
- 提交流程
- 测试要求
- 文档更新

## 📄 许可证

本项目遵循项目根目录的许可证协议。

---

## 🔗 相关资源

- [项目主 README](../../README.md)
- [中文开发文档](../../中文开发文档.md)
- [Tools 模块文档](../tools/README.md)
- [测试文档](../../tests/)
- [llama.cpp 项目](https://github.com/ggerganov/llama.cpp)
- [OpenAI API 文档](https://platform.openai.com/docs/)

---

**最后更新：** 2025-11-19
**文档版本：** 0.0.1
