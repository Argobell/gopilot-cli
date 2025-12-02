# Gopilot-CLI

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

<a title="中文" href="./README_zh.md">
  <img src="https://img.shields.io/badge/-中文-F54A00?style=for-the-badge" alt="中文">
</a>
<a title="English" href="./README.md">
  <img src="https://img.shields.io/badge/-English-545759?style=for-the-badge" alt="English">
</a>

<div align="center">
  <img alt="Gopilot-CLI" src="./assets/gopilot.png" width="800">
</div>

Gopilot-CLI is a terminal-based, multi-turn AI coding agent written in Go.  
It talks to an OpenAI-compatible API and can run shell commands, read/write files in your workspace, and auto-summarize long sessions.

## Requirements

- Go `1.21+`
- An OpenAI-compatible chat completion API (e.g. OpenAI, local gateway, or OSS service)
- A valid API key for that service

## Quick Start

### Installation

```bash
# Clone and build
git clone <repository-url>
cd gopilot-cli
go build -o gopilot ./cmd/gopilot
```

### Configuration

Gopilot-CLI reads its main configuration from `configs/config.yaml`.  
You can set your API key either in this file or via environment variable:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

In `configs/config.yaml`:

```yaml
llm:
  api_key: "sk-xxx"                # optional if you use OPENAI_API_KEY
  api_base: "https://api.openai.com/v1"  # or your own compatible endpoint
  model: "gpt-4.1"                 # or any compatible model

agent:
  workspace_dir: "./workspace"     # default workspace folder
  max_steps: 50
  token_limit: 80000               # triggers history summarization
```

When both are set, the value in `configs/config.yaml` (`llm.api_key`) takes precedence over `OPENAI_API_KEY`.

### Usage

```bash
# Run with current directory as workspace
./gopilot

# Or specify workspace directory
./gopilot -w /path/to/workspace
```

Typical workflow:

- Run `gopilot` inside a project directory
- Describe the task you want it to perform (e.g. “add a REST handler”, “fix failing tests”)
- Let it use tools to edit files and run `go` commands for you

## Features

- 🔄 **Multi-turn Conversations** with context preservation
- 🛠️ **Tool Calling** for commands and file operations
- 📝 **Auto-summarization** when token limits exceeded
- 🎨 **Interactive Terminal** with command completion
- 🔁 **Retry Mechanism** with exponential backoff

## Tools

### Bash Tools
- `Bash` - Execute shell commands
- `BashOutput` - Monitor background processes
- `BashKill` - Terminate processes

### File Tools
- `Read` - Read files within workspace
- `Write` - Create/overwrite files
- `Edit` - Modify file contents

## Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/clear` | Clear session history |
| `/history` | Display message count |
| `/stats` | Show session statistics |
| `/exit` | Exit program |

Also supports: `exit`, `quit`, or `q`

## Development

```bash
# Build the binary
go build -o gopilot ./cmd/gopilot

# Run tests
go test ./tests/...

# Format code
go fmt ./...
```

## Dependencies

- [`openai/openai-go/v3`](https://github.com/openai/openai-go) - OpenAI API SDK
- [`c-bata/go-prompt`](https://github.com/c-bata/go-prompt) - Interactive terminal
- [`pkoukk/tiktoken-go`](https://github.com/pkoukk/tiktoken-go) - Token counting
- [`stretchr/testify`](https://github.com/stretchr/testify) - Testing assertions
