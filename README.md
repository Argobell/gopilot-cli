```
gopilot-cli/
├── cmd/
│   └── agent/
│       └── main.go              # 程序入口
├── internal/
│   ├── agent/
│   │   └── agent.go             # Agent 核心循环逻辑
│   ├── config/
│   │   └── config. go            # 配置加载与管理
│   ├── llm/
│   │   └── client. go            # LLM 客户端 (基于 openai-go)
│   ├── retry/
│   │   └── retry. go             # 重试机制
│   ├── schema/
│   │   └── schema. go            # 数据模型定义
│   ├── tools/
│   │   ├── base.go              # Tool 接口定义
│   │   ├── bash.go              # Bash 工具
│   │   └── file. go             # 文件工具
│   ├── logger/
│   │   └── logger. go            # 日志记录
│   └── utils/
│       └── utils.go             # 工具函数
├── configs/
│   ├── config.yaml              # 默认配置
│   └── system_prompt.md         # 系统提示词
├── examples/
│   └── simple_agent.go          # 示例代码
├── go.mod
├── go.sum
└── README.md
``` 

```
gopilot-cli/
├── main.go                      # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载与管理
│   ├── agent/
│   │   └── agent.go             # Agent 核心循环逻辑
│   ├── llm/
│   │   └── client.go            # LLM 客户端 (基于 openai-go)
│   ├── retry/
│   │   └── retry.go             # 重试机制
│   ├── schema/
│   │   └── schema.go            # 数据模型定义
│   ├── tools/
│   │   ├── base.go              # Tool 接口定义
│   │   ├── bash.go              # Bash 工具
│   │   └── file.go              # 文件工具
│   └── logger/
│       └── logger.go            # 日志记录
├── configs/
│   ├── config.yaml              # 默认配置
│   └── system_prompt.txt        # 系统提示词
├── tests/
│   ├── bash_test.go             # Bash 工具测试
│   ├── client_test.go           # LLM 客户端测试
│   └── utils_test.go            # 工具函数测试
├── api.md                       # API 文档
├── go.mod
├── go.sum
└── README.md
```