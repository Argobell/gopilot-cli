```
gopilot-cli/
├── cmd/
│   └── cli/
│       └── main.go              # 程序入口
├── internal/
│   ├── agent/
│   │   └── agent.go             # Agent 核心循环逻辑
│   ├── config/
│   │   └── config.go            # 配置加载与管理
│   ├── llm/
│   │   └── client.go            # LLM 客户端 (基于 openai-go)
│   ├── logger/
│   │   └── logger.go            # 日志记录
│   ├── retry/
│   │   └── retry.go             # 重试机制
│   ├── schema/
│   │   └── schema.go            # 数据模型定义
│   ├── tools/
│   │   ├── base.go              # Tool 接口定义
│   │   ├── bash.go              # Bash 工具
│   │   └── file.go              # 文件工具
│   └── utils/
│       ├── path/
│       │   └── path.go          # 路径处理
│       └── terminal/
│           └── terminal.go      # 终端交互
├── configs/
│   ├── config.yaml              # 默认配置
│   └── system_prompt.txt        # 系统提示词
├── tests/
│   ├── agent_test.go            # Agent 测试
│   ├── bash_test.go             # Bash 工具测试
│   ├── client_test.go           # LLM 客户端测试
│   └── utils_test.go            # 工具函数测试
├── go.mod
├── go.sum
└── README.md
```