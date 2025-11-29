# API 参考文档

本文档提供 Agent 模块的完整 API 参考，包括所有类、方法和属性的详细说明。

## 📋 目录

- [类概览](#类概览)
- [Agent 类](#agent-类)
- [Colors 类](#colors-类)
- [方法详解](#方法详解)
- [使用示例](#使用示例)
- [常见问题](#常见问题)

---

## 类概览

### 核心类

| 类 | 说明 | 文件位置 |
|----|------|----------|
| `Agent` | 智能代理主类，实现自主任务执行 | src/agent.py:41 |
| `Colors` | ANSI 颜色定义类，用于彩色终端输出 | src/agent.py:15 |

### 相关类

| 类 | 说明 | 来源 |
|----|------|------|
| `LLMClient` | LLM 客户端基类 | src/llm/llm_wrapper.py |
| `Message` | 消息数据模型 | src/schema.py |
| `Tool` | 工具基类 | src/tools/base.py |
| `ToolResult` | 工具执行结果 | src/tools/base.py |

---

## Agent 类

**完整路径**：`src.agent.Agent`

**继承**：无，直接继承自 object

**说明**：智能代理核心类，负责协调 LLM 交互、工具调用、消息管理和任务执行。

### 类的功能特性

- ✅ 自主执行任务（循环执行模式）
- ✅ 工具调用支持（Function Calling）
- ✅ 消息历史管理
- ✅ 智能消息摘要
- ✅ Token 估算和管理
- ✅ 工作空间目录管理
- ✅ 彩色终端输出
- ✅ 异步执行

### 属性

#### `llm: LLMClient`

- **类型**：`LLMClientBase`
- **说明**：LLM 客户端实例，负责与 LLM 通信
- **示例**：

```python
agent = Agent(llm_client=my_llm_client, ...)
print(agent.llm)  # <src.llm.openai_client.OpenAIClient object>
```

#### `tools: dict[str, Tool]`

- **类型**：`dict[str, Tool]`
- **说明**：工具字典，以工具名称为键，工具实例为值
- **创建方式**：`{tool.name: tool for tool in tools}`
- **示例**：

```python
class MyTool(Tool):
    name = "my_tool"
    # ...

agent = Agent(llm_client=..., tools=[MyTool()])
print(list(agent.tools.keys()))  # ['my_tool']
```

#### `max_steps: int`

- **类型**：`int`
- **默认值**：`30`
- **说明**：最大执行步数，防止无限循环
- **示例**：

```python
agent = Agent(..., max_steps=10)
print(agent.max_steps)  # 10
```

#### `token_limit: int`

- **类型**：`int`
- **默认值**：`80000`
- **说明**：Token 限制，超过时触发消息摘要
- **示例**：

```python
agent = Agent(..., token_limit=60000)
print(agent.token_limit)  # 60000
```

#### `workspace_dir: Path`

- **类型**：`pathlib.Path`
- **说明**：工作空间目录 Path 对象
- **示例**：

```python
agent = Agent(..., workspace_dir="./workspace")
print(agent.workspace_dir)  # PosixPath('./workspace')
print(agent.workspace_dir.absolute())  # /absolute/path/to/workspace
```

#### `system_prompt: str`

- **类型**：`str`
- **说明**：系统提示，如果不含 "Current Workspace" 会自动添加工作空间信息
- **示例**：

```python
agent = Agent(
    llm_client=...,
    system_prompt="你是一个助手",
    ...
)
print(agent.system_prompt)  # "你是一个助手\n\n## Current Workspace\n..."
```

#### `messages: list[Message]`

- **类型**：`list[Message]`
- **说明**：消息历史列表，初始包含系统消息
- **示例**：

```python
agent = Agent(...)
print(len(agent.messages))  # 1 (只有系统消息)

agent.add_user_message("你好")
print(len(agent.messages))  # 2
```

---

## Colors 类

**完整路径**：`src.agent.Colors`

**继承**：无，直接继承自 object

**说明**：ANSI 颜色码定义类，用于终端彩色输出。

### 类变量

#### 样式常量

| 常量 | 值 | 用途 |
|------|----|----|
| `RESET` | `"\033[0m"` | 重置所有样式 |
| `BOLD` | `"\033[1m"` | 粗体 |
| `DIM` | `"\033[2m"` | 暗色 |

#### 基础颜色

| 常量 | 值 | 用途 |
|------|----|----|
| `RED` | `"\033[31m"` | 红色 |
| `GREEN` | `"\033[32m"` | 绿色 |
| `YELLOW` | `"\033[33m"` | 黄色 |
| `BLUE` | `"\033[34m"` | 蓝色 |
| `MAGENTA` | `"\033[35m"` | 紫色 |
| `CYAN` | `"\033[36m"` | 青色 |

#### 高亮颜色

| 常量 | 值 | 用途 |
|------|----|----|
| `BRIGHT_BLACK` | `"\033[90m"` | 亮黑色 |
| `BRIGHT_RED` | `"\033[91m"` | 亮红色 |
| `BRIGHT_GREEN` | `"\033[92m"` | 亮绿色 |
| `BRIGHT_YELLOW` | `"\033[93m"` | 亮黄色 |
| `BRIGHT_BLUE` | `"\033[94m"` | 亮蓝色 |
| `BRIGHT_MAGENTA` | `"\033[95m"` | 亮紫色 |
| `BRIGHT_CYAN` | `"\033[96m"` | 亮青色 |
| `BRIGHT_WHITE` | `"\033[97m"` | 亮白色 |

### 使用示例

```python
print(f"{Colors.BRIGHT_GREEN}✓ 成功！{Colors.RESET}")
print(f"{Colors.BRIGHT_RED}✗ 失败！{Colors.RESET}")
print(f"{Colors.BOLD}{Colors.BRIGHT_BLUE}重要信息{Colors.RESET}")
print(f"{Colors.DIM}这是一行暗色文字{Colors.RESET}")
```

---

## 方法详解

## `Agent.__init__()`

**签名**：
```python
def __init__(
    self,
    llm_client: LLMClient,
    system_prompt: str,
    tools: list[Tool],
    max_steps: int = 30,
    workspace_dir: str = "./workspace",
    token_limit: int = 80000,
):
```

**功能**：初始化 Agent 实例

**参数**：

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `llm_client` | `LLMClient` | - | **必需**。LLM 客户端实例 |
| `system_prompt` | `str` | - | **必需**。系统提示词 |
| `tools` | `list[Tool]` | - | **必需**。工具列表 |
| `max_steps` | `int` | `30` | 可选。最大执行步数 |
| `workspace_dir` | `str` | `"./workspace"` | 可选。工作空间目录 |
| `token_limit` | `int` | `80000` | 可选。Token 限制 |

**功能说明**：
1. 初始化所有属性
2. 创建工作空间目录
3. 自动添加工作空间信息到系统提示
4. 初始化消息历史（包含系统消息）

**返回值**：无

**异常**：
- 无显式异常，但可能在创建目录时抛出 `OSError`

**示例**：

```python
from src.agent import Agent
from src.llm import LLMClient, LLMProvider

# 基础初始化
agent = Agent(
    llm_client=llm_client,
    system_prompt="你是一个专业助手",
    tools=[]
)

# 完整配置
agent = Agent(
    llm_client=llm_client,
    system_prompt="你是一个 Python 编程助手",
    tools=[my_tool],
    max_steps=20,
    workspace_dir="/tmp/workspace",
    token_limit=60000
)

print(agent.workspace_dir)  # PosixPath('/tmp/workspace')
print(agent.max_steps)  # 20
```

## `Agent.add_user_message()`

**签名**：
```python
def add_user_message(self, content: str) -> None:
```

**功能**：添加用户消息到消息历史

**参数**：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| `content` | `str` | 消息内容 |

**功能说明**：
1. 创建 `Message(role="user", content=content)` 对象
2. 将消息追加到 `messages` 列表

**返回值**：无

**示例**：

```python
agent = Agent(...)

# 方式 1：使用 add_user_message
agent.add_user_message("你好，请介绍一下自己")

# 方式 2：直接操作 messages
from src.schema import Message
agent.messages.append(
    Message(role="user", content="Python 有什么特性？")
)

print(len(agent.messages))  # 3 (system + user1 + user2)
```

## `Agent._estimate_tokens()`

**签名**：
```python
def _estimate_tokens(self) -> int:
```

**功能**：估算当前消息历史的 Token 数量

**功能说明**：
1. 尝试使用 tiktoken 的 `cl100k_base` 编码器
2. 计算所有消息内容的 Token 数
3. 计算思考内容的 Token 数
4. 计算工具调用的 Token 数
5. 添加每条消息的元数据开销（约 4 tokens）
6. 如果 tiktoken 不可用，回退到 `_estimate_tokens_fallback()`

**返回值**：`int` - 估算的 Token 数量

**异常**：
- 无显式异常，可能抛出 `Exception`（内部捕获）

**内部实现**：
```python
def _estimate_tokens(self) -> int:
    try:
        encoding = tiktoken.get_encoding("cl100k_base")
    except Exception:
        return self._estimate_tokens_fallback()

    total_tokens = 0

    for msg in self.messages:
        # 计算内容
        if isinstance(msg.content, str):
            total_tokens += len(encoding.encode(msg.content))

        # 计算思考内容
        if msg.thinking:
            total_tokens += len(encoding.encode(msg.thinking))

        # 计算工具调用
        if msg.tool_calls:
            total_tokens += len(encoding.encode(str(msg.tool_calls)))

        # 元数据开销
        total_tokens += 4

    return total_tokens
```

**示例**：

```python
agent = Agent(...)

# 添加消息
agent.add_user_message("这是一个很长的消息" * 100)

# 估算 Token
token_count = agent._estimate_tokens()
print(f"当前消息约含 {token_count} 个 Token")
```

## `Agent._estimate_tokens_fallback()`

**签名**：
```python
def _estimate_tokens_fallback(self) -> int:
```

**功能**：回退的 Token 估算方法

**功能说明**：
1. 当 tiktoken 不可用时使用
2. 基于字符数进行粗略估算
3. 平均 2.5 字符 per token

**返回值**：`int` - 估算的 Token 数量

**示例**：

```python
# 内部自动调用，无需手动调用
token_count = agent._estimate_tokens_fallback()
```

## `Agent._summarize_messages()`

**签名**：
```python
async def _summarize_messages(self) -> None:
```

**功能**：检查并执行消息摘要

**功能说明**：
1. 估算当前 Token 数
2. 如果未超过限制，直接返回
3. 如果超过限制，执行摘要：
   - 保留系统消息
   - 保留所有用户消息
   - 对每个执行轮次进行摘要
   - 替换为摘要消息
4. 计算摘要后的 Token 数
5. 输出摘要统计信息

**返回值**：无

**输出**：
- 触发摘要时的提示信息
- 摘要完成后的统计信息

**消息结构变化**：
```
摘要前: [System] [U1] [A1] [T1] [A2] [T2] [A3] [U2]
摘要后: [System] [U1] [Summary1] [U2]
```

**示例**：

```python
agent = Agent(..., token_limit=40000)  # 较低限制

# 长时间对话
for i in range(20):
    agent.add_user_message(f"问题 {i}")
    await agent.run()

# Agent 会自动执行消息摘要
# 可通过查看 agent.messages 验证
```

## `Agent._create_summary()`

**签名**：
```python
async def _create_summary(
    self,
    messages: list[Message],
    round_num: int
) -> str:
```

**功能**：为单个执行轮次创建摘要

**参数**：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| `messages` | `list[Message]` | 要摘要的消息列表 |
| `round_num` | `int` | 轮次编号（从 1 开始） |

**功能说明**：
1. 构建摘要内容，格式为：
   ```
   Round {round_num} 执行过程:

   Assistant: ...
   -> 调用工具: ...
   <- 工具返回: ...
   ```
2. 调用 LLM 生成简洁摘要
3. 返回摘要文本

**返回值**：`str` - 摘要文本

**异常**：
- 可能在调用 LLM 时抛出异常，捕获后返回原始内容

**示例**：

```python
# 内部方法，通常不需要手动调用
# 但可以通过以下方式测试
round_messages = agent.messages[1:5]  # 一个轮次的消息
summary = await agent._create_summary(round_messages, 1)
print(summary)
```

## `Agent.run()`

**签名**：
```python
async def run(self) -> str:
```

**功能**：执行 Agent 循环

**功能说明**：
1. 进入执行循环（最多 `max_steps` 次）
2. 每次循环：
   - 检查并摘要消息历史
   - 显示步骤信息
   - 调用 LLM 生成响应
   - 显示思考内容（如果有）
   - 显示响应内容
   - 检查是否完成（无工具调用）
   - 执行工具调用（如果有）
   - 显示工具调用和结果
   - 添加消息到历史
   - 更新步骤数
3. 如果达到最大步数，返回超时错误

**返回值**：
- `str` - 最终结果或错误信息

**循环流程**：
```
while step < max_steps:
    1. 消息摘要
    2. LLM 生成
    3. 显示响应
    4. 工具调用
    5. 检查完成
    6. 更新步骤
```

**输出格式**：
```
步骤头:
╭────────────────────────────────────────────────────────────╮│
 Step X/max_steps                                           │
╰────────────────────────────────────────────────────────────╯

思考内容（如果有）：
🧠 Thinking:
思考过程

Assistant 响应：
🤖 Assistant:
响应内容

工具调用（如果有）：
🔧 Tool Call: 工具名
   Arguments:
   {参数}

✓ Result: 结果
✗ Error: 错误
```

**示例**：

```python
import asyncio
from src.agent import Agent

async def basic_run():
    agent = Agent(
        llm_client=llm_client,
        system_prompt="你是一个编程助手",
        tools=[],
        max_steps=10
    )

    agent.add_user_message("写一个快速排序算法")

    result = await agent.run()

    print("="*50)
    print("最终结果:")
    print(result)

# 运行
asyncio.run(basic_run())
```

**完整示例**：

```python
import asyncio
from src.agent import Agent
from src.llm import LLMClient, LLMProvider
from src.tools.base import Tool, ToolResult

class CalculatorTool(Tool):
    name = "calculator"
    description = "执行数学计算"

    async def execute(self, expression: str) -> ToolResult:
        try:
            result = eval(expression)
            return ToolResult(success=True, content=str(result))
        except Exception as e:
            return ToolResult(success=False, content="", error=str(e))

async def full_example():
    # 1. 创建 LLM 客户端
    llm_client = LLMClient(
        api_key="your-api-key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    # 2. 创建工具
    calculator = CalculatorTool()

    # 3. 创建 Agent
    agent = Agent(
        llm_client=llm_client,
        system_prompt=(
            "你是一个数学助手，可以使用计算器工具。"
        ),
        tools=[calculator],
        max_steps=15,
        workspace_dir="./workspace",
        token_limit=60000
    )

    # 4. 添加任务
    agent.add_user_message("计算 (15 + 23) * (37 - 12) / 5")

    # 5. 运行
    result = await agent.run()

    print("\n" + "="*60)
    print("任务完成！")
    print(f"结果: {result}")

# 运行
asyncio.run(full_example())
```

---

## 使用示例

### 示例 1：最小化使用

```python
import asyncio
from src.agent import Agent
from src.llm import LLMClient, LLMProvider

async def minimal_example():
    llm_client = LLMClient(
        api_key="key",
        provider=LLMProvider.OPENAI,
        api_base="http://localhost:8080",
        model="gpt-oss"
    )

    agent = Agent(
        llm_client=llm_client,
        system_prompt="你是一个助手",
        tools=[]
    )

    agent.add_user_message("你好")
    result = await agent.run()
    print(result)

asyncio.run(minimal_example())
```

### 示例 2：多轮对话

```python
import asyncio
from src.agent import Agent

async def conversation_example():
    agent = Agent(...)

    # 第 1 轮
    agent.add_user_message("什么是机器学习？")
    answer1 = await agent.run()

    # 第 2 轮（Agent 保留历史）
    agent.add_user_message("能详细说说监督学习吗？")
    answer2 = await agent.run()

    # 第 3 轮
    agent.add_user_message("推荐一些学习资源")
    answer3 = await agent.run()

    print(f"第 1 轮: {answer1}")
    print(f"第 2 轮: {answer2}")
    print(f"第 3 轮: {answer3}")

asyncio.run(conversation_example())
```

### 示例 3：工具调用

```python
import asyncio
from src.agent import Agent
from src.tools.base import Tool, ToolResult

class FileReaderTool(Tool):
    name = "read_file"
    description = "读取文件"

    async def execute(self, file_path: str) -> ToolResult:
        try:
            with open(file_path, 'r') as f:
                content = f.read()
            return ToolResult(success=True, content=content)
        except Exception as e:
            return ToolResult(success=False, content="", error=str(e))

async def tool_example():
    agent = Agent(
        llm_client=llm_client,
        system_prompt="你是一个文件助手",
        tools=[FileReaderTool()],
        max_steps=10
    )

    agent.add_user_message("读取 README.md 文件")
    result = await agent.run()
    print(result)

asyncio.run(tool_example())
```

### 示例 4：消息历史访问

```python
async def history_example():
    agent = Agent(...)

    agent.add_user_message("问题 1")
    await agent.run()

    agent.add_user_message("问题 2")
    await agent.run()

    # 查看消息历史
    print(f"消息数量: {len(agent.messages)}")

    for i, msg in enumerate(agent.messages):
        print(f"\n消息 {i}:")
        print(f"  角色: {msg.role}")
        print(f"  内容: {msg.content[:100]}...")
```

### 示例 5：自定义工作空间

```python
import os
from pathlib import Path

async def workspace_example():
    workspace = "/tmp/custom_workspace"
    os.makedirs(workspace, exist_ok=True)

    agent = Agent(
        llm_client=llm_client,
        system_prompt="你是一个文件助手",
        tools=[],
        workspace_dir=workspace
    )

    # 工作空间自动创建
    print(f"工作空间: {agent.workspace_dir}")
    print(f"存在: {agent.workspace_dir.exists()}")

    # 可以读取环境变量获取路径
    env_path = os.getenv("AGENT_WORKSPACE", "./default_workspace")
    agent2 = Agent(..., workspace_dir=env_path)

asyncio.run(workspace_example())
```

### 示例 6：Token 监控

```python
async def token_monitoring():
    agent = Agent(..., token_limit=20000)  # 较低限制

    for i in range(10):
        agent.add_user_message(f"问题 {i}")
        await agent.run()

        # 手动检查 Token
        tokens = agent._estimate_tokens()
        print(f"当前 Token: {tokens}/{agent.token_limit}")

        if tokens > agent.token_limit * 0.8:
            print("⚠️ Token 使用接近限制")

asyncio.run(token_monitoring())
```

### 示例 7：错误处理

```python
async def error_handling():
    agent = Agent(
        llm_client=None,  # 故意传入无效值
        system_prompt="测试",
        tools=[],
        max_steps=5
    )

    agent.add_user_message("测试")

    try:
        result = await agent.run()
        print("结果:", result)
    except Exception as e:
        print(f"捕获异常: {type(e).__name__}: {e}")

asyncio.run(error_handling())
```

### 示例 8：并发执行

```python
async def concurrent_agents():
    # 创建多个 Agent
    agents = [
        Agent(llm_client, f"你是一个助手 {i}", [], max_steps=5)
        for i in range(3)
    ]

    # 添加任务
    tasks = [
        agent.add_user_message(f"任务 {i}")
        for i, agent in enumerate(agents)
    ]

    # 并发运行
    results = await asyncio.gather(
        *[agent.run() for agent in agents]
    )

    for i, result in enumerate(results):
        print(f"Agent {i}: {result}")

asyncio.run(concurrent_agents())
```

---

## 常见问题

### Q1：如何查看所有可用的颜色？

**A：**

```python
from src.agent import Colors

# 打印所有颜色常量
print("基础颜色:")
for name in ['RED', 'GREEN', 'YELLOW', 'BLUE', 'MAGENTA', 'CYAN']:
    print(f"{Colors.__dict__[name]}{name}{Colors.RESET}")

print("\n高亮颜色:")
for name in ['BRIGHT_RED', 'BRIGHT_GREEN', 'BRIGHT_YELLOW',
             'BRIGHT_BLUE', 'BRIGHT_MAGENTA', 'BRIGHT_CYAN']:
    print(f"{Colors.__dict__[name]}{name}{Colors.RESET}")
```

### Q2：如何自定义系统提示？

**A：**

```python
# 方式 1：直接在初始化时设置
agent = Agent(
    llm_client=llm_client,
    system_prompt="""
    你是一个专业的 Python 编程助手。
    你的特点：
    - 擅长编写优雅、高效的代码
    - 遵循 PEP 8 编码规范
    - 总是提供代码示例和解释
    """,
    tools=[]
)

# 方式 2：修改现有 Agent
agent.system_prompt = "新的系统提示"
# 需要重新初始化 messages
from src.schema import Message
agent.messages = [Message(role="system", content=agent.system_prompt)]
```

### Q3：如何访问消息历史？

**A：**

```python
# 查看所有消息
for i, msg in enumerate(agent.messages):
    print(f"{i}: {msg.role} - {msg.content[:50]}...")

# 查看特定角色消息
user_messages = [m for m in agent.messages if m.role == "user"]
assistant_messages = [m for m in agent.messages if m.role == "assistant"]

# 查看最近 N 条消息
recent_messages = agent.messages[-5:]
```

### Q4：如何手动触发消息摘要？

**A：**

```python
# 直接调用私有方法（不推荐）
await agent._summarize_messages()

# 或者通过降低 token_limit 强制触发
agent.token_limit = 1  # 设置为极小值
await agent.run()  # 会立即触发摘要
```

### Q5：如何查看工具调用历史？

**A：**

```python
# 查找所有工具调用
tool_calls = []
for msg in agent.messages:
    if msg.tool_calls:
        tool_calls.extend(msg.tool_calls)

# 查看工具调用详情
for tc in tool_calls:
    print(f"工具: {tc.function.name}")
    print(f"参数: {tc.function.arguments}")
```

### Q6：如何获取执行统计？

**A：**

```python
# 统计消息数量
print(f"总消息数: {len(agent.messages)}")
print(f"用户消息: {sum(1 for m in agent.messages if m.role == 'user')}")
print(f"助手消息: {sum(1 for m in agent.messages if m.role == 'assistant')}")
print(f"工具消息: {sum(1 for m in agent.messages if m.role == 'tool')}")

# 统计工具调用
tool_call_count = sum(
    len(msg.tool_calls) if msg.tool_calls else 0
    for msg in agent.messages
)
print(f"工具调用次数: {tool_call_count}")

# 估算 Token 使用
print(f"Token 使用: {agent._estimate_tokens()}/{agent.token_limit}")
```

### Q7：如何暂停和恢复执行？

**A：**

```python
# Agent 不直接支持暂停
# 可以通过以下方式实现

class PausableAgent(Agent):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.paused = False
        self.pause_event = asyncio.Event()

    async def run(self):
        # 在适当位置检查暂停状态
        if self.paused:
            await self.pause_event.wait()
        # ... 其他逻辑

# 使用
agent = PausableAgent(...)
# 暂停
agent.paused = True
# 恢复
agent.paused = False
agent.pause_event.set()
```

### Q8：如何清理消息历史？

**A：**

```python
# 保留系统消息，清空其他
agent.messages = [agent.messages[0]]

# 保留最近 N 条消息
keep_count = 20
system_msg = agent.messages[0] if agent.messages[0].role == "system" else None
recent = agent.messages[-keep_count:]
agent.messages = ([system_msg] if system_msg else []) + recent

# 手动摘要
await agent._summarize_messages()
```

---

## 相关文档

- [快速入门](./快速入门.md)
- [架构设计](./架构设计.md)
- [开发者指南](./开发者指南.md)
- [README](./README.md)

---

**上一篇：** [架构设计](./架构设计.md)
**下一篇：** [开发者指南](./开发者指南.md)