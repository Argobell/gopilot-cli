# Tools 模块开发文档

欢迎使用 gopilot-cli Tools 模块开发文档！

## 📚 文档概览

本文档集提供了 `src/tools` 模块的完整开发指南，包括架构设计、API 参考、使用教程和扩展指南。

### 什么是 Tools 模块？

Tools 模块是 gopilot-cli 的核心组件之一，提供了一套可扩展的工具系统，用于执行各种系统操作。目前主要实现了 Bash 命令执行功能，支持前台和后台执行模式。

### 核心特性

- ✅ **异步执行**：所有工具基于 asyncio 实现，支持高性能并发
- ✅ **跨平台兼容**：自动适配 Windows (PowerShell) 和 Unix/Linux/macOS (bash)
- ✅ **后台进程管理**：完整的生命周期管理，支持输出监控和进程终止
- ✅ **文件操作工具**：读取、写入、编辑文件，支持分块读取和智能截断
- ✅ **OpenAI 兼容**：自动生成 OpenAI 工具调用格式，便于 AI 集成
- ✅ **强类型验证**：基于 Pydantic 的数据模型，确保类型安全
- ✅ **易于扩展**：清晰的抽象基类，方便添加新工具

## 📖 文档导航

### [🚀 快速入门](./快速入门.md)
适合刚开始使用 Tools 模块的开发者。涵盖：
- 环境准备和安装
- 基本使用示例
- 常见使用场景
- 5 分钟上手教程

**推荐阅读顺序：第一篇**

---

### [🏗️ 架构设计](./架构设计.md)
深入了解 Tools 模块的内部设计。包括：
- 模块架构图（Mermaid 类图和序列图）
- 设计模式和设计原则
- 核心组件说明
- 模块依赖关系

**推荐阅读顺序：第二篇**

---

### [📋 API 参考](./API参考.md)
完整的 API 文档，包含所有类和方法的详细说明。涵盖：
- Tool 基类 API
- ToolResult 数据模型
- BashTool、BashOutputTool、BashKillTool 详细 API
- BackgroundShell 和 BackgroundShellManager
- 参数、返回值、异常说明
- 丰富的代码示例

**推荐阅读顺序：第三篇（作为参考手册）**

---

### [👨‍💻 开发者指南](./开发者指南.md)
学习如何扩展和自定义 Tools 模块。包括：
- 创建自定义工具的步骤
- 最佳实践和设计模式
- 错误处理策略
- 测试指南
- 贡献代码规范

**推荐阅读顺序：第四篇（进阶内容）**

---

## 🎯 快速链接

### 常见任务

| 任务 | 文档链接 |
|------|---------|
| 执行简单命令 | [快速入门 - 基本使用](./快速入门.md#基本使用) |
| 启动后台进程 | [快速入门 - 后台执行](./快速入门.md#后台执行) |
| 监控进程输出 | [API 参考 - BashOutputTool](./API参考.md#bashoutputtool) |
| 读取文件 | [快速入门 - 文件操作示例](./快速入门.md#文件操作示例) |
| 编辑文件 | [快速入门 - 文件操作示例](./快速入门.md#文件操作示例) |
| 创建新工具 | [开发者指南 - 创建自定义工具](./开发者指南.md#创建自定义工具) |
| 了解架构 | [架构设计 - 模块架构](./架构设计.md#模块架构) |

### API 速查

| 类/方法 | 说明 | 文档链接 |
|---------|------|---------|
| `Tool` | 工具基类 | [API 参考 - Tool](./API参考.md#tool-基类) |
| `BashTool.execute()` | 执行 Bash 命令 | [API 参考 - BashTool](./API参考.md#bashtool) |
| `BashOutputTool.execute()` | 获取后台输出 | [API 参考 - BashOutputTool](./API参考.md#bashoutputtool) |
| `BashKillTool.execute()` | 终止后台进程 | [API 参考 - BashKillTool](./API参考.md#bashkilltool) |
| `ReadTool.execute()` | 读取文件内容 | [API 参考 - ReadTool](./API参考.md#readtool) |
| `WriteTool.execute()` | 写入文件内容 | [API 参考 - WriteTool](./API参考.md#writetool) |
| `EditTool.execute()` | 编辑文件内容 | [API 参考 - EditTool](./API参考.md#edittool) |
| `BackgroundShellManager` | 后台进程管理器 | [API 参考 - BackgroundShellManager](./API参考.md#backgroundshellmanager) |

## 🔧 模块结构

```
src/tools/
├── __init__.py              # 模块初始化
├── base.py                  # 基类定义
│   ├── Tool                 # 工具抽象基类
│   └── ToolResult           # 工具结果数据模型
├── bash_tool.py             # Bash 工具实现
│   ├── BashTool             # Bash 命令执行工具
│   ├── BashOutputTool       # 后台输出监控工具
│   ├── BashKillTool         # 后台进程终止工具
│   ├── BashOutputResult     # Bash 结果数据模型
│   ├── BackgroundShell      # 后台进程数据容器
│   └── BackgroundShellManager  # 后台进程管理器
└── file_tools.py            # 文件操作工具实现
    ├── ReadTool             # 文件读取工具
    ├── WriteTool            # 文件写入工具
    └── EditTool             # 文件编辑工具
    └── truncate_text_by_tokens()  # 文本截断工具函数
```

## 💡 使用示例

### 快速示例：执行命令

```python
from src.tools.bash_tool import BashTool

# 创建工具实例
bash_tool = BashTool()

# 执行简单命令
result = await bash_tool.execute("echo 'Hello, World!'")

print(result.success)   # True
print(result.stdout)    # Hello, World!
print(result.exit_code) # 0
```

### 快速示例：后台执行

```python
from src.tools.bash_tool import BashTool, BashOutputTool, BashKillTool

# 启动后台服务
bash_tool = BashTool()
result = await bash_tool.execute(
    command="python -m http.server 8080",
    run_in_background=True
)

bash_id = result.bash_id  # 获取进程 ID

# 监控输出
output_tool = BashOutputTool()
output = await output_tool.execute(bash_id=bash_id)
print(output.stdout)

# 终止进程
kill_tool = BashKillTool()
await kill_tool.execute(bash_id=bash_id)
```

### 快速示例：文件操作

```python
from src.tools.file_tools import ReadTool, WriteTool, EditTool

# 读取文件
read_tool = ReadTool()
content = await read_tool.execute(path="example.txt")

# 写入文件
write_tool = WriteTool()
await write_tool.execute(
    path="output.txt",
    content="Hello, World!"
)

# 编辑文件
edit_tool = EditTool()
await edit_tool.execute(
    path="example.txt",
    old_str="旧文本",
    new_str="新文本"
)
```

## 📦 依赖

Tools 模块依赖以下 Python 库：

- **Python 3.12+**
- **asyncio**：异步 I/O 支持（标准库）
- **pydantic**：数据验证和模型定义
- **typing**：类型注解（标准库）
- **tiktoken**：文本分词和 token 计数（用于文件内容截断）

其他标准库依赖：
- `platform`：平台检测
- `re`：正则表达式
- `time`：时间戳
- `uuid`：唯一 ID 生成

## 🧪 测试

Tools 模块拥有完整的测试覆盖（90%+）：

```bash
# 运行所有测试
uv run pytest tests/test_bash_tool.py -v

# 查看测试覆盖率
uv run pytest tests/test_bash_tool.py --cov=src.tools.bash_tool --cov-report=term-missing
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

**最后更新：** 2025-11-19
**文档版本：** 0.0.1
