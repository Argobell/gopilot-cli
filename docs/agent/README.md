# Agent 模块开发文档

欢迎使用 gopilot-cli Agent 模块开发文档！

## 📚 文档概览

本文档集提供了 `src/agent.py` 模块的完整开发指南，包括架构设计、API 参考、使用教程和扩展指南。

### 什么是 Agent 模块？

Agent 模块是 gopilot-cli 的核心组件，实现了一个智能代理（Agent）系统，具备自主任务执行、工具调用、消息管理和上下文管理能力。模块采用事件驱动架构，支持多轮对话、自动工具调用、消息历史管理和智能消息摘要，特别适合构建复杂的 AI 助手和自动化任务系统。

### 核心特性

- ✅ **自主执行**：智能代理循环，自动执行任务直到完成
- ✅ **工具调用**：完整支持 Function Calling，可调用多种工具
- ✅ **消息管理**：自动管理对话历史，支持 system/user/assistant/tool 四种消息类型
- ✅ **上下文管理**：智能消息摘要，防止上下文溢出（基于 token 限制）
- ✅ **彩色输出**：丰富的 ANSI 彩色终端输出，提升可读性
- ✅ **思考模式**：支持链式思考（reasoning），适合复杂问题求解
- ✅ **异步执行**：基于 asyncio 的高性能异步执行
- ✅ **强类型**：基于类型注解的类型安全设计
- ✅ **错误处理**：完善的异常处理和错误恢复机制
- ✅ **工作空间**：内置工作空间目录管理，支持相对路径解析

## 📖 文档导航

### [🚀 快速入门](./快速入门.md)
适合刚开始使用 Agent 模块的开发者。涵盖：
- 环境准备和初始化
- 基本 Agent 创建和运行
- 工具集成示例
- 多轮对话场景
- 5 分钟上手教程

**推荐阅读顺序：第一篇**

---

### [🏗️ 架构设计](./架构设计.md)
深入了解 Agent 模块的内部设计。包括：
- 模块架构图（事件驱动架构）
- 设计模式和设计原则
- 核心组件说明（Agent 类、消息管理、工具调用）
- 模块依赖关系
- 执行流程详解

**推荐阅读顺序：第二篇**

---

### [📋 API 参考](./API参考.md)
完整的 API 文档，包含所有类和方法的详细说明。涵盖：
- Agent 类 API
- Colors 类 API
- 初始化参数详解
- run() 方法详解
- 消息管理 API
- 参数、返回值、异常说明
- 丰富的代码示例

**推荐阅读顺序：第三篇（作为参考手册）**

---

### [👨‍💻 开发者指南](./开发者指南.md)
学习如何扩展和自定义 Agent 模块。包括：
- 创建自定义 Agent
- 添加新工具
- 自定义消息处理
- 最佳实践和设计模式
- 错误处理策略
- 性能优化建议
- 贡献代码规范

**推荐阅读顺序：第四篇（进阶内容）**

---

## 🎯 快速链接

### 常见任务

| 任务 | 文档链接 |
|------|---------|
| 创建 Agent 实例 | [快速入门 - 基本使用](./快速入门.md#基本使用) |
| 运行 Agent | [快速入门 - 运行 Agent](./快速入门.md#运行-agent) |
| 添加用户消息 | [快速入门 - 添加消息](./快速入门.md#添加消息) |
| 使用工具调用 | [快速入门 - 工具调用](./快速入门.md#工具调用) |
| 管理消息历史 | [快速入门 - 消息管理](./快速入门.md#消息管理) |
| 配置工作空间 | [快速入门 - 工作空间](./快速入门.md#工作空间) |
| 自定义工具 | [开发者指南 - 自定义工具](./开发者指南.md#自定义工具) |

### API 速查

| 类/方法 | 说明 | 文档链接 |
|---------|------|---------|
| `Agent` | 智能代理主类 | [API 参考 - Agent](./API参考.md#agent-类) |
| `Colors` | 终端颜色定义类 | [API 参考 - Colors](./API参考.md#colors-类) |
| `Agent.__init__()` | 初始化 Agent | [API 参考 - __init__()](./API参考.md#__init__) |
| `Agent.run()` | 执行代理循环 | [API 参考 - run()](./API参考.md#run) |
| `Agent.add_user_message()` | 添加用户消息 | [API 参考 - add_user_message()](./API参考.md#add_user_message) |
| `Agent._summarize_messages()` | 消息摘要 | [API 参考 - _summarize_messages()](./API参考.md#_summarize_messages) |

## 🔧 模块结构

```
src/
└── agent.py                   # 智能代理核心实现
    ├── Colors                 # ANSI 颜色定义类
    │   ├── RESET, BOLD, DIM   # 样式常量
    │   ├── RED, GREEN, ...    # 基础颜色
    │   └── BRIGHT_*           # 高亮颜色
    │
    └── Agent                  # 智能代理主类
        ├── __init__()         # 初始化代理
        │   ├── llm            # LLM 客户端
        │   ├── tools          # 工具字典
        │   ├── max_steps      # 最大步数
        │   ├── token_limit    # Token 限制
        │   ├── workspace_dir  # 工作空间
        │   ├── system_prompt  # 系统提示
        │   └── messages       # 消息历史
        │
        ├── add_user_message() # 添加用户消息
        ├── _estimate_tokens() # Token 估算
        ├── _summarize_messages() # 消息摘要
        ├── _create_summary()  # 创建摘要
        │
        └── run()              # 执行代理循环
            ├── _summarize_messages() # 检查并摘要
            ├── 显示步骤信息
            ├── LLM 生成响应
            ├── 处理工具调用
            ├── 执行工具
            ├── 添加消息历史
            └── 检查完成条件
```

## 💡 使用示例

### 快速示例：基本 Agent

```python
import asyncio
from src.agent import Agent
from src.llm import LLMClient, LLMProvider
from src.schema import Message

async def basic_agent():
    # 创建 LLM 客户端
    llm_client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    # 定义系统提示
    system_prompt = "你是一个专业的 Python 编程助手"

    # 定义工具（可选）
    tools = [
        # 工具列表...
    ]

    # 创建 Agent
    agent = Agent(
        llm_client=llm_client,
        system_prompt=system_prompt,
        tools=tools,
        max_steps=10,
        workspace_dir="./workspace"
    )

    # 添加用户消息
    agent.add_user_message("请帮我写一个快速排序算法")

    # 运行 Agent
    result = await agent.run()

    print("最终结果:")
    print(result)

# 运行
# asyncio.run(basic_agent())
```

### 快速示例：多步任务执行

```python
async def multi_step_agent():
    llm_client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    agent = Agent(
        llm_client=llm_client,
        system_prompt="你是一个数据分析师，擅长处理和解释数据",
        tools=[],
        max_steps=20
    )

    # 复杂任务
    agent.add_user_message(
        "分析当前目录下所有 CSV 文件的趋势，"
        "并生成一份综合报告"
    )

    result = await agent.run()

    print("任务完成:", result)
```

### 快速示例：带思考模式的 Agent

```python
async def thinking_agent():
    llm_client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    agent = Agent(
        llm_client=llm_client,
        system_prompt=(
            "你是数学专家。请逐步思考并解决数学问题。"
            "使用思考模式来展示你的推理过程。"
        ),
        tools=[],
        max_steps=15
    )

    agent.add_user_message("计算 15 * 23 * 37，并展示思考过程")

    result = await agent.run()

    # 思考内容会在运行时显示
```

## 📦 依赖

Agent 模块依赖以下组件：

- **Python 3.12+**：异步 I/O 支持
- **asyncio**：异步执行框架
- **tiktoken**：OpenAI 的 Tokenizer，用于精确 Token 计数
- **src.llm**：LLM 客户端模块
- **src.schema**：消息数据模型
- **src.tools**：工具基类和结果类
- **src.utils**：工具函数（calculate_display_width）

其他标准库依赖：
- `json`：JSON 数据处理
- `pathlib`：路径操作
- `typing`：类型系统

## 🎨 支持的功能

当前 Agent 模块支持：

| 功能 | 状态 | 说明 |
|------|------|------|
| 自主执行 | ✅ 支持 | 自动循环执行直到任务完成或达到最大步数 |
| 工具调用 | ✅ 支持 | 完整的 Function Calling 支持 |
| 思考模式 | ✅ 支持 | 显示 AI 的思考过程（reasoning） |
| 消息摘要 | ✅ 支持 | 自动防止上下文溢出 |
| 彩色输出 | ✅ 支持 | ANSI 彩色终端输出 |
| 工作空间 | ✅ 支持 | 工作空间目录管理 |
| 异步执行 | ✅ 支持 | 基于 asyncio 的高性能执行 |
| 错误处理 | ✅ 支持 | 完善的异常处理和恢复 |

未来计划：
- 流式响应支持
- 并发工具执行
- 插件化架构
- 可视化界面

## 🧪 测试

Agent 模块可以配合测试框架使用：

```bash
# 运行测试
uv run pytest tests/ -v -k agent

# 查看测试覆盖率
uv run pytest tests/ --cov=src.agent --cov-report=term-missing
```

详见：[项目测试文档](../../tests/)

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
- [LLM 模块文档](../llm/README.md)
- [Tools 模块文档](../tools/README.md)
- [测试文档](../../tests/)
- [OpenAI Function Calling](https://platform.openai.com/docs/guides/function-calling)
- [asyncio 文档](https://docs.python.org/3/library/asyncio.html)

---

**最后更新：** 2025-11-28
**文档版本：** 0.0.1