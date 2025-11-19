# 更新日志

本文件记录 Gopilot-CLI 项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

## [0.0.2] - 2025-11-19

### 变更
- **代码格式化改进**：统一代码风格，提升可读性和一致性
  - 优化代码缩进和空行使用
  - 改进文档字符串格式
  - 统一代码结构布局

### 测试
- **测试文件优化**：改进测试用例质量
  - 更新 `test_basic_chat.py`：统一代码格式，优化测试用例
  - 更新 `test_thinking.py`：改进思考内容测试逻辑

### 文档
- **文档结构重组**：优化项目文档组织
  - 删除 `中文开发文档.md`（文档内容迁移至独立目录）
  - 文档文件移动到 `docs/` 目录进行统一管理

### 开发
- **新增功能**：
  - 新增 `tests/test_bash_tool.py` 测试文件
  - 新增 `src/tools/` 模块目录
  - 新增 `docs/` 文档目录

## [0.0.1] - 2025-11-18

### 新增
- 初始版本发布，提供统一的 LLM 客户端接口库
- **统一接口架构**：
  - 新增 `LLMClientBase` 抽象基类，定义标准客户端接口
  - 新增 `LLMClient` 统一包装器，支持根据提供商自动选择客户端实现
  - 采用策略模式和工厂模式设计，易于扩展和维护

- **OpenAI 提供商支持**：
  - 新增 `OpenAIClient` 实现类，完整支持 OpenAI 兼容 API
  - 自动处理 API 端点后缀（自动添加 `/v1`）
  - 支持自定义 API 基础地址和模型名称

- **对话消息系统**：
  - 新增 `Message` 数据模型，支持多角色消息（system、user、assistant、tool）
  - 支持多轮对话和对话历史管理
  - 完整的消息类型验证和序列化

- **工具调用功能**：
  - 新增 `ToolCall` 和 `FunctionCall` 数据模型
  - 支持 function/tool calling 能力
  - 支持工具调用的完整生命周期管理
  - 灵活的工具定义方式（字典格式和对象格式）

- **推理内容支持**：
  - 新增 `reasoning_details` 提取功能
  - 支持模型的思考过程展示
  - 独立的 `thinking` 字段与响应内容分离

- **数据验证与类型安全**：
  - 基于 Pydantic v2.12.4+ 的强类型数据模型
  - 完整的输入输出验证
  - 优秀的 IDE 类型提示支持

- **核心数据模型**：
  - 新增 `LLMResponse` 响应模型
  - 新增 `LLMProvider` 枚举类型
  - 所有模型均支持完整的类型注解

### 技术规格
- Python 3.12+ 要求
- OpenAI SDK >=2.8.1
- Pydantic >=2.12.4
- 异步编程支持（async/await）
- 线程安全设计

### 文档
- 新增详细的中文开发文档（`中文开发文档.md`）
- 包含完整的快速开始指南
- 提供多个实际使用示例：
  - 基本对话示例
  - 多轮对话示例
  - 工具调用完整流程示例
  - 推理内容提取示例
- 详细的 API 参考文档
- 扩展新提供商的开发指南
- 最佳实践建议

### 项目结构
```
gopilot-cli/
├── src/
│   ├── __init__.py              # 包初始化，导出主要 API
│   ├── llm/                     # LLM 客户端模块
│   │   ├── __init__.py
│   │   ├── base.py              # 抽象基类定义
│   │   ├── openai_client.py     # OpenAI 提供商实现
│   │   └── llm_wrapper.py       # 统一包装器
│   └── schema/                  # 数据模型模块
│       ├── __init__.py
│       └── schema.py            # Pydantic 数据模型
├── pyproject.toml               # 项目配置
└── README.md                    # 项目说明
```

### 示例代码

#### 基本使用示例
```python
from src import LLMClient, Message

client = LLMClient(
    api_key="your-api-key",
    provider=LLMProvider.OPENAI,
    api_base="https://api.openai.com",
    model="gpt-4"
)

messages = [
    Message(role="system", content="你是一个有用的助手。"),
    Message(role="user", content="你好，请介绍一下自己。")
]

response = await client.generate(messages)
print(response.content)
```

#### 工具调用示例
```python
tools = [
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "获取城市天气信息",
            "parameters": {
                "type": "object",
                "properties": {
                    "city": {"type": "string"}
                }
            }
        }
    }
]

response = await client.generate(messages, tools=tools)
if response.tool_calls:
    # 处理工具调用
    pass
```

### 已知限制
- 当前版本仅支持 OpenAI 兼容的 API 提供商
- 其他提供商（如 Anthropic、Google 等）将在后续版本中支持
- 流式响应功能尚未实现（后续版本规划中）

### 致谢
感谢所有为项目初期版本贡献想法和反馈的开发者。

---

## 版本说明

### 版本号规则
- **主版本号**：不兼容的 API 修改
- **次版本号**：向下兼容的功能性新增
- **修订号**：向下兼容的问题修正

### 更新类型
- `新增`：新功能
- `变更`：对现有功能的修改
- `弃用`：即将删除的功能
- `移除`：已删除的功能
- `修复`：问题修复
- `安全`：安全相关修复

### 获取帮助
- 提交 Issue：[项目 Issues 页面]
- 贡献代码：欢迎提交 Pull Request
