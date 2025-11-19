# API 参考文档

本文档提供 Tools 模块所有类、方法和数据模型的完整 API 参考。

## 📋 目录

- [基础类](#基础类)
  - [Tool 基类](#tool-基类)
  - [ToolResult 数据模型](#toolresult-数据模型)
- [Bash 工具](#bash-工具)
  - [BashTool](#bashtool)
  - [BashOutputTool](#bashoutputtool)
  - [BashKillTool](#bashkilltool)
  - [BashOutputResult](#bashoutputresult)
- [后台进程管理](#后台进程管理)
  - [BackgroundShell](#backgroundshell)
  - [BackgroundShellManager](#backgroundshellmanager)

---

## 基础类

### Tool 基类

`src/tools/base.py`

所有工具的抽象基类，定义统一的工具接口。

#### 类签名

```python
class Tool:
    """工具抽象基类"""
```

#### 属性

##### `name`

```python
@property
def name(self) -> str:
    """工具的唯一标识符"""
```

**返回值**：
- `str`：工具名称，用于标识和调用工具

**注意**：子类必须实现此属性

**示例**：
```python
class MyTool(Tool):
    @property
    def name(self) -> str:
        return "my_tool"
```

---

##### `description`

```python
@property
def description(self) -> str:
    """工具的功能描述"""
```

**返回值**：
- `str`：工具的详细功能说明，用于帮助和文档

**注意**：子类必须实现此属性

**示例**：
```python
@property
def description(self) -> str:
    return "这是一个示例工具，用于演示工具系统"
```

---

##### `parameters`

```python
@property
def parameters(self) -> dict[str, Any]:
    """工具参数的 JSON Schema 定义"""
```

**返回值**：
- `dict[str, Any]`：符合 JSON Schema 规范的参数定义

**JSON Schema 结构**：
```python
{
    "type": "object",
    "properties": {
        "param_name": {
            "type": "string",           # 参数类型
            "description": "参数说明",   # 参数描述
            "default": "默认值"          # 可选：默认值
        }
    },
    "required": ["param_name"]          # 必需参数列表
}
```

**示例**：
```python
@property
def parameters(self) -> dict[str, Any]:
    return {
        "type": "object",
        "properties": {
            "input_text": {
                "type": "string",
                "description": "要处理的输入文本"
            },
            "options": {
                "type": "object",
                "description": "可选配置项",
                "default": {}
            }
        },
        "required": ["input_text"]
    }
```

#### 方法

##### `execute()`

```python
async def execute(self, *args, **kwargs) -> ToolResult:
    """异步执行工具的核心逻辑"""
```

**参数**：
- `*args`：位置参数（根据具体工具而定）
- `**kwargs`：关键字参数（根据具体工具而定）

**返回值**：
- `ToolResult`：执行结果对象

**异常**：
- `NotImplementedError`：基类未实现，子类必须重写

**注意**：
- 此方法是 `async` 异步方法，必须使用 `await` 调用
- 子类必须实现具体的执行逻辑

**示例**：
```python
async def execute(self, input_text: str, options: dict = None) -> ToolResult:
    try:
        # 执行逻辑
        result = process_text(input_text, options or {})
        return ToolResult(success=True, content=result)
    except Exception as e:
        return ToolResult(success=False, error=str(e), content="")
```

---

##### `to_openai_schema()`

```python
def to_openai_schema(self) -> dict[str, Any]:
    """转换为 OpenAI 工具调用格式"""
```

**返回值**：
- `dict[str, Any]`：OpenAI 兼容的工具定义

**返回格式**：
```python
{
    "type": "function",
    "function": {
        "name": self.name,
        "description": self.description,
        "parameters": self.parameters
    }
}
```

**用途**：
用于将工具注册到 OpenAI API 进行函数调用

**示例**：
```python
tool = MyTool()
openai_schema = tool.to_openai_schema()

# 在 OpenAI API 中使用
response = openai.ChatCompletion.create(
    model="gpt-4",
    messages=[...],
    tools=[openai_schema]
)
```

---

### ToolResult 数据模型

`src/tools/base.py`

标准化的工具执行结果数据模型。

#### 类签名

```python
class ToolResult(BaseModel):
    """工具执行结果基类"""
```

#### 字段

##### `success`

```python
success: bool
```

**类型**：`bool`

**说明**：执行是否成功

**取值**：
- `True`：执行成功
- `False`：执行失败

---

##### `content`

```python
content: str
```

**类型**：`str`

**说明**：格式化后的输出内容

**用途**：
- 成功时包含主要输出
- 失败时可能包含错误信息

---

##### `error`

```python
error: str | None = None
```

**类型**：`str | None`

**默认值**：`None`

**说明**：错误信息（如果有）

**取值**：
- `None`：没有错误
- `str`：错误描述

#### 示例

```python
# 成功结果
result = ToolResult(
    success=True,
    content="操作完成",
    error=None
)

# 失败结果
result = ToolResult(
    success=False,
    content="",
    error="参数错误：input_text 不能为空"
)

# 访问字段
if result.success:
    print(result.content)
else:
    print(f"错误：{result.error}")
```

---

## Bash 工具

### BashTool

`src/tools/bash_tool.py`

执行 Bash 或 PowerShell 命令的工具，支持前台和后台执行。

#### 类签名

```python
class BashTool(Tool):
    """执行 shell 命令的工具（前台或后台）"""
```

#### 构造函数

```python
def __init__(self):
    """初始化 BashTool，自动检测操作系统"""
```

**说明**：
- 自动检测操作系统（Windows/Unix/Linux/macOS）
- 设置对应的 shell 名称（PowerShell 或 bash）

**示例**：
```python
bash_tool = BashTool()
print(bash_tool.shell_name)  # Windows: "PowerShell", 其他: "bash"
```

#### 属性

##### `name`

```python
@property
def name(self) -> str:
    return "bash"
```

**返回值**：`"bash"`

---

##### `description`

```python
@property
def description(self) -> str:
    """返回工具描述（根据操作系统）"""
```

**返回值**：详细的工具使用说明（Windows 或 Unix 版本）

---

##### `parameters`

```python
@property
def parameters(self) -> dict[str, Any]:
    """返回参数定义"""
```

**返回值**：
```python
{
    "type": "object",
    "properties": {
        "command": {
            "type": "string",
            "description": "要执行的命令"
        },
        "timeout": {
            "type": "integer",
            "description": "超时时间（秒），默认 120，最大 600",
            "default": 120
        },
        "run_in_background": {
            "type": "boolean",
            "description": "是否后台运行",
            "default": False
        }
    },
    "required": ["command"]
}
```

#### 方法

##### `execute()`

```python
async def execute(
    self,
    command: str,
    timeout: int = 120,
    run_in_background: bool = False,
) -> BashOutputResult:
    """执行 shell 命令"""
```

**参数**：

| 参数 | 类型 | 必需 | 默认值 | 说明 |
|------|------|------|--------|------|
| `command` | `str` | ✅ | - | 要执行的命令 |
| `timeout` | `int` | ❌ | `120` | 超时时间（秒），最大 600 |
| `run_in_background` | `bool` | ❌ | `False` | 是否后台运行 |

**返回值**：
- `BashOutputResult`：包含执行结果的对象

**超时行为**：
- 前台执行：超过 `timeout` 秒后自动终止进程
- 后台执行：`timeout` 参数被忽略

**示例 1：前台执行**

```python
bash_tool = BashTool()

# 执行简单命令
result = await bash_tool.execute("echo 'Hello'")

if result.success:
    print(result.stdout)        # Hello
    print(result.exit_code)     # 0

# 执行错误命令
result = await bash_tool.execute("exit 1")

if not result.success:
    print(result.error)         # Command failed with exit code 1
    print(result.exit_code)     # 1
```

**示例 2：后台执行**

```python
bash_tool = BashTool()

# 启动后台服务
result = await bash_tool.execute(
    command="python -m http.server 8080",
    run_in_background=True
)

if result.success:
    bash_id = result.bash_id
    print(f"服务已启动，进程 ID: {bash_id}")
    # 后续可用 bash_id 监控或终止
```

**示例 3：设置超时**

```python
bash_tool = BashTool()

# 10 秒超时
result = await bash_tool.execute(
    command="sleep 100",
    timeout=10
)

# 如果超时，success 为 False
if not result.success:
    print("命令超时")
```

**跨平台注意事项**：

```python
# Unix/Linux/macOS
result = await bash_tool.execute("ls -la")
result = await bash_tool.execute("echo 'text' > file.txt")

# Windows
result = await bash_tool.execute("dir")
result = await bash_tool.execute("echo 'text' | Out-File file.txt")
```

---

### BashOutputTool

`src/tools/bash_tool.py`

监控后台 Bash 进程的输出。

#### 类签名

```python
class BashOutputTool(Tool):
    """获取后台进程输出的工具"""
```

#### 属性

##### `name`

```python
@property
def name(self) -> str:
    return "bash_output"
```

##### `parameters`

```python
{
    "type": "object",
    "properties": {
        "bash_id": {
            "type": "string",
            "description": "后台 shell 的 ID"
        },
        "filter_str": {
            "type": "string",
            "description": "可选的正则表达式过滤器"
        }
    },
    "required": ["bash_id"]
}
```

#### 方法

##### `execute()`

```python
async def execute(
    self,
    bash_id: str,
    filter_str: str | None = None,
) -> BashOutputResult:
    """获取后台进程的输出"""
```

**参数**：

| 参数 | 类型 | 必需 | 默认值 | 说明 |
|------|------|------|--------|------|
| `bash_id` | `str` | ✅ | - | 后台进程的唯一标识符 |
| `filter_str` | `str \| None` | ❌ | `None` | 正则表达式过滤器 |

**返回值**：
- `BashOutputResult`：包含新输出的结果对象

**重要特性**：
- ✅ **增量输出**：仅返回自上次调用后的新输出
- ✅ **正则过滤**：可选择性地只返回匹配的行
- ✅ **非阻塞**：立即返回，不等待新输出

**示例 1：获取所有新输出**

```python
output_tool = BashOutputTool()

# 获取新输出
result = await output_tool.execute(bash_id="abc12345")

if result.success:
    print("新输出:")
    print(result.stdout)
else:
    print(f"错误：{result.error}")
```

**示例 2：过滤特定内容**

```python
output_tool = BashOutputTool()

# 只获取包含 "ERROR" 的行
result = await output_tool.execute(
    bash_id="abc12345",
    filter_str="ERROR"
)

if result.success and result.stdout:
    print("错误日志:")
    print(result.stdout)
```

**示例 3：循环监控**

```python
output_tool = BashOutputTool()
bash_id = "abc12345"

# 持续监控 10 次
for i in range(10):
    await asyncio.sleep(1)  # 每秒检查一次

    result = await output_tool.execute(bash_id=bash_id)

    if result.success:
        if result.stdout:
            print(f"[{i+1}] 新输出:")
            print(result.stdout)
        else:
            print(f"[{i+1}] 暂无新输出")
```

**错误处理**：

```python
result = await output_tool.execute(bash_id="nonexistent")

if not result.success:
    print(result.error)  # "Shell not found: nonexistent. Available: [...]"
```

---

### BashKillTool

`src/tools/bash_tool.py`

终止后台 Bash 进程。

#### 类签名

```python
class BashKillTool(Tool):
    """终止后台进程的工具"""
```

#### 属性

##### `name`

```python
@property
def name(self) -> str:
    return "bash_kill"
```

##### `parameters`

```python
{
    "type": "object",
    "properties": {
        "bash_id": {
            "type": "string",
            "description": "要终止的后台 shell ID"
        }
    },
    "required": ["bash_id"]
}
```

#### 方法

##### `execute()`

```python
async def execute(self, bash_id: str) -> BashOutputResult:
    """终止后台进程"""
```

**参数**：

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `bash_id` | `str` | ✅ | 要终止的后台进程 ID |

**返回值**：
- `BashOutputResult`：包含终止状态和剩余输出的结果对象

**终止过程**：
1. 获取进程的剩余输出
2. 发送 `SIGTERM` 信号（优雅终止）
3. 等待最多 5 秒
4. 如果未终止，发送 `SIGKILL` 信号（强制终止）
5. 清理所有相关资源

**示例 1：基本使用**

```python
kill_tool = BashKillTool()

# 终止后台进程
result = await kill_tool.execute(bash_id="abc12345")

if result.success:
    print("进程已终止")

    # 查看终止前的最后输出
    if result.stdout:
        print("最后输出:")
        print(result.stdout)
```

**示例 2：完整流程**

```python
bash_tool = BashTool()
kill_tool = BashKillTool()

# 1. 启动后台进程
start_result = await bash_tool.execute(
    command="python long_running_task.py",
    run_in_background=True
)

bash_id = start_result.bash_id

# 2. 等待一段时间
await asyncio.sleep(10)

# 3. 终止进程
kill_result = await kill_tool.execute(bash_id=bash_id)

if kill_result.success:
    print("任务已终止")
    print(f"退出码: {kill_result.exit_code}")
```

**错误处理**：

```python
result = await kill_tool.execute(bash_id="nonexistent")

if not result.success:
    print(result.error)  # "Shell not found: nonexistent"
```

**注意事项**：
- ⚠️ 终止后，进程 ID 将从管理器中移除，无法再次访问
- ⚠️ 终止是永久性的，无法恢复
- ✅ 建议在终止前使用 `BashOutputTool` 保存重要输出

---

### BashOutputResult

`src/tools/bash_tool.py`

Bash 工具专用的执行结果数据模型。

#### 类签名

```python
class BashOutputResult(ToolResult):
    """Bash 命令执行结果（扩展 ToolResult）"""
```

#### 字段

##### 继承字段

- `success: bool`：执行是否成功
- `content: str`：自动格式化的输出（由 `format_content()` 生成）
- `error: str | None`：错误信息

##### 扩展字段

###### `stdout`

```python
stdout: str
```

**类型**：`str`

**说明**：命令的标准输出

---

###### `stderr`

```python
stderr: str
```

**类型**：`str`

**说明**：命令的标准错误输出

---

###### `exit_code`

```python
exit_code: int
```

**类型**：`int`

**说明**：进程退出码

**常见取值**：
- `0`：成功
- `1`：一般错误
- `127`：命令未找到
- `-1`：内部错误或超时

---

###### `bash_id`

```python
bash_id: str | None = None
```

**类型**：`str | None`

**默认值**：`None`

**说明**：后台进程的唯一标识符（仅后台执行时）

#### 方法

##### `format_content()`

```python
@model_validator(mode="after")
def format_content(self) -> "BashOutputResult":
    """自动从 stdout 和 stderr 生成 content 字段"""
```

**说明**：
- Pydantic 模型验证器，自动调用
- 合并 `stdout`、`stderr` 和 `exit_code` 到 `content`

**格式**：
```
{stdout}
[stderr]
{stderr}
[exit_code]: {exit_code}
```

**示例**：

```python
result = BashOutputResult(
    success=False,
    stdout="Some output",
    stderr="Error message",
    exit_code=1
)

print(result.content)
# 输出:
# Some output
# [stderr]
# Error message
# [exit_code]: 1
```

#### 使用示例

```python
# 检查执行结果
result = await bash_tool.execute("ls")

print(f"成功: {result.success}")
print(f"标准输出: {result.stdout}")
print(f"标准错误: {result.stderr}")
print(f"退出码: {result.exit_code}")
print(f"格式化内容: {result.content}")

# 后台执行结果
result = await bash_tool.execute("sleep 10", run_in_background=True)

print(f"后台 ID: {result.bash_id}")  # abc12345
```

---

## 后台进程管理

### BackgroundShell

`src/tools/bash_tool.py`

封装单个后台 shell 进程的数据容器。

#### 类签名

```python
class BackgroundShell:
    """后台 shell 数据容器（纯数据类）"""
```

#### 构造函数

```python
def __init__(
    self,
    bash_id: str,
    command: str,
    process: asyncio.subprocess.Process,
    start_time: float
):
    """初始化后台 shell"""
```

**参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `bash_id` | `str` | 唯一标识符 |
| `command` | `str` | 执行的命令 |
| `process` | `asyncio.subprocess.Process` | 进程对象 |
| `start_time` | `float` | 启动时间戳 |

#### 属性

| 属性 | 类型 | 说明 |
|------|------|------|
| `bash_id` | `str` | 唯一标识符 |
| `command` | `str` | 执行的命令 |
| `process` | `Process` | asyncio 进程对象 |
| `start_time` | `float` | 启动时间（`time.time()`） |
| `output_lines` | `list[str]` | 输出行缓冲 |
| `last_read_index` | `int` | 最后读取位置 |
| `status` | `str` | 进程状态 |
| `exit_code` | `int \| None` | 退出码 |

**`status` 取值**：
- `"running"`：正在运行
- `"completed"`：成功完成（exit_code == 0）
- `"failed"`：失败完成（exit_code != 0）
- `"terminated"`：被终止
- `"error"`：发生错误

#### 方法

##### `add_output()`

```python
def add_output(self, line: str) -> None:
    """添加新的输出行"""
```

**参数**：
- `line: str`：输出行内容

**示例**：
```python
shell.add_output("Processing item 1...")
shell.add_output("Processing item 2...")
```

---

##### `get_new_output()`

```python
def get_new_output(self, filter_pattern: str | None = None) -> list[str]:
    """获取自上次读取后的新输出"""
```

**参数**：
- `filter_pattern: str | None`：可选的正则表达式过滤器

**返回值**：
- `list[str]`：新输出行列表

**特性**：
- 仅返回未读的输出行
- 自动更新 `last_read_index`
- 支持正则表达式过滤

**示例**：
```python
# 获取所有新输出
new_lines = shell.get_new_output()
print("\n".join(new_lines))

# 只获取包含 "ERROR" 的行
error_lines = shell.get_new_output(filter_pattern="ERROR")
print("\n".join(error_lines))

# 第二次调用返回空（没有新输出）
more_lines = shell.get_new_output()  # []
```

---

##### `update_status()`

```python
def update_status(self, is_alive: bool, exit_code: int | None = None) -> None:
    """更新进程状态"""
```

**参数**：
- `is_alive: bool`：进程是否存活
- `exit_code: int | None`：退出码（如果进程结束）

**行为**：
- `is_alive=True`：设置状态为 `"running"`
- `is_alive=False` + `exit_code=0`：设置状态为 `"completed"`
- `is_alive=False` + `exit_code≠0`：设置状态为 `"failed"`

**示例**：
```python
# 进程运行中
shell.update_status(is_alive=True)

# 进程成功结束
shell.update_status(is_alive=False, exit_code=0)
assert shell.status == "completed"

# 进程失败结束
shell.update_status(is_alive=False, exit_code=1)
assert shell.status == "failed"
```

---

##### `terminate()`

```python
async def terminate(self) -> None:
    """异步终止进程"""
```

**行为**：
1. 发送 `SIGTERM` 信号
2. 等待最多 5 秒
3. 超时则发送 `SIGKILL` 信号
4. 更新状态为 `"terminated"`

**示例**：
```python
# 终止进程
await shell.terminate()

print(shell.status)      # "terminated"
print(shell.exit_code)   # 进程退出码
```

---

### BackgroundShellManager

`src/tools/bash_tool.py`

全局后台进程管理器（单例模式）。

#### 类签名

```python
class BackgroundShellManager:
    """全局后台 shell 管理器（单例）"""
```

#### 类变量

```python
_shells: dict[str, BackgroundShell] = {}
_monitor_tasks: dict[str, asyncio.Task] = {}
```

**说明**：
- `_shells`：存储所有后台进程（bash_id → BackgroundShell）
- `_monitor_tasks`：存储所有监控任务（bash_id → Task）

#### 类方法

##### `add()`

```python
@classmethod
def add(cls, shell: BackgroundShell) -> None:
    """添加后台 shell 到管理器"""
```

**参数**：
- `shell: BackgroundShell`：要添加的 shell 对象

**示例**：
```python
shell = BackgroundShell("id-1", "sleep 10", process, time.time())
BackgroundShellManager.add(shell)
```

---

##### `get()`

```python
@classmethod
def get(cls, bash_id: str) -> BackgroundShell | None:
    """获取指定 ID 的后台 shell"""
```

**参数**：
- `bash_id: str`：shell 的唯一标识符

**返回值**：
- `BackgroundShell | None`：找到则返回 shell 对象，否则返回 None

**示例**：
```python
shell = BackgroundShellManager.get("id-1")

if shell:
    print(f"找到进程: {shell.command}")
else:
    print("进程不存在")
```

---

##### `get_available_ids()`

```python
@classmethod
def get_available_ids(cls) -> list[str]:
    """获取所有可用的 bash ID"""
```

**返回值**：
- `list[str]`：所有后台进程的 ID 列表

**示例**：
```python
ids = BackgroundShellManager.get_available_ids()
print(f"当前有 {len(ids)} 个后台进程")
print(f"进程 ID: {', '.join(ids)}")
```

---

##### `start_monitor()`

```python
@classmethod
async def start_monitor(cls, bash_id: str) -> None:
    """启动对指定 shell 的输出监控"""
```

**参数**：
- `bash_id: str`：要监控的 shell ID

**行为**：
- 创建异步监控任务
- 持续读取进程输出
- 自动更新进程状态
- 进程结束时自动停止

**监控流程**：
1. 持续读取 `process.stdout`
2. 每读到一行，调用 `shell.add_output()`
3. 进程结束时，调用 `shell.update_status()`
4. 清理监控任务

**示例**：
```python
# 添加并启动监控
BackgroundShellManager.add(shell)
await BackgroundShellManager.start_monitor(shell.bash_id)

# 监控任务在后台持续运行
```

---

##### `terminate()`

```python
@classmethod
async def terminate(cls, bash_id: str) -> BackgroundShell:
    """终止指定的后台 shell 并清理所有资源"""
```

**参数**：
- `bash_id: str`：要终止的 shell ID

**返回值**：
- `BackgroundShell`：被终止的 shell 对象

**异常**：
- `ValueError`：如果 shell 不存在

**清理步骤**：
1. 调用 `shell.terminate()` 终止进程
2. 取消监控任务
3. 从管理器中移除 shell

**示例**：
```python
try:
    shell = await BackgroundShellManager.terminate("id-1")
    print(f"已终止进程: {shell.command}")
    print(f"退出码: {shell.exit_code}")
except ValueError as e:
    print(f"错误: {e}")
```

---

## 完整使用示例

### 示例 1：前台命令执行

```python
from src.tools.bash_tool import BashTool
import asyncio

async def main():
    bash_tool = BashTool()

    # 执行 git 命令
    result = await bash_tool.execute("git status")

    if result.success:
        print("Git 状态:")
        print(result.stdout)
    else:
        print(f"执行失败: {result.error}")

asyncio.run(main())
```

### 示例 2：后台任务管理

```python
from src.tools.bash_tool import BashTool, BashOutputTool, BashKillTool
import asyncio

async def background_task_demo():
    bash_tool = BashTool()
    output_tool = BashOutputTool()
    kill_tool = BashKillTool()

    # 1. 启动后台任务
    result = await bash_tool.execute(
        command="python -m http.server 8000",
        run_in_background=True
    )

    bash_id = result.bash_id
    print(f"服务已启动: {bash_id}")

    # 2. 监控输出
    for i in range(5):
        await asyncio.sleep(2)
        output = await output_tool.execute(bash_id=bash_id)

        if output.stdout:
            print(f"[{i+1}] 输出:")
            print(output.stdout)

    # 3. 终止服务
    kill_result = await kill_tool.execute(bash_id=bash_id)
    print(f"服务已终止，退出码: {kill_result.exit_code}")

asyncio.run(background_task_demo())
```

### 示例 3：错误处理

```python
async def error_handling_demo():
    bash_tool = BashTool()

    # 尝试执行可能失败的命令
    result = await bash_tool.execute("nonexistent_command")

    if not result.success:
        print("命令执行失败:")
        print(f"  退出码: {result.exit_code}")
        print(f"  错误: {result.error}")

        if result.stderr:
            print(f"  stderr: {result.stderr}")

asyncio.run(error_handling_demo())
```

---

**上一篇：** [架构设计](./架构设计.md)
**下一篇：** [开发者指南](./开发者指南.md)
