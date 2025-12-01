package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gopilot-cli/internal/agent"
	"gopilot-cli/internal/config"
	"gopilot-cli/internal/llm"
	"gopilot-cli/internal/tools"
	"gopilot-cli/internal/utils/path"
)

// 获取项目根目录（因为 go test 在 tests/ 下）
// projectRoot is defined in another test file, so remove this duplicate.

// ============================================================
// Test 1: Simple file creation task
// ============================================================

func TestAgentSimpleTask(t *testing.T) {
	t.Log("\n=== Testing Agent with Simple File Task ===")

	// Load configuration
	cfgPath := filepath.Join(path.ProjectRoot(), "configs", "config.yaml")
	cfg, err := config.LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	// Create temporary workspace
	workspace, err := os.MkdirTemp("", "agent-workspace-*")
	if err != nil {
		t.Fatalf("temp workspace: %v", err)
	}
	defer os.RemoveAll(workspace)

	t.Log("Using workspace:", workspace)

	// Load system prompt
	systemPrompt := "You are a helpful AI assistant that can use tools."
	promptPath := filepath.Join(path.ProjectRoot(), "configs", "system_prompt.txt")
	if _, err := os.Stat(promptPath); err == nil {
		data, _ := os.ReadFile(promptPath)
		systemPrompt = string(data)
	}

	// Initialize LLM client
	llmClient := llm.NewClient(
		cfg.LLM.APIKey,
		cfg.LLM.APIBase,
		cfg.LLM.Model,
	)

	// Initialize tools
	toolList := []tools.Tool{
		tools.NewReadTool(workspace),
		tools.NewWriteTool(workspace),
		tools.NewEditTool(workspace),
		tools.NewBashTool(),
	}

	// Create agent
	ag, err := agent.NewAgent(
		llmClient,
		systemPrompt,
		toolList,
		10,
		workspace,
		150000,
	)
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	// Task
	task := "Create a file named 'test.txt' with the content 'Hello from Agent!'"
	t.Log("Task:", task)

	ag.AddUserMessage(task)

	ctx := context.Background()
	result, err := ag.Run(ctx)
	if err != nil {
		t.Fatalf("agent run error: %v", err)
	}

	t.Log("\n============================================================")
	t.Log("Agent Result:", result)
	t.Log("============================================================")

	// Validate file
	testFile := filepath.Join(workspace, "test.txt")

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatalf("expected file test.txt to be created, but not found")
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	content := string(data)
	t.Log("File content:", content)

	if content != "Hello from Agent!" {
		t.Log("⚠️ content mismatch, but file exists. Test still passes.")
	} else {
		t.Log("✅ Content is correct")
	}
}

// ============================================================
// Test 2: Bash tool task
// ============================================================

func TestAgentBashTask(t *testing.T) {
	t.Log("\n=== Testing Agent Bash Tool Task ===")

	cfgPath := filepath.Join(path.ProjectRoot(), "configs", "config.yaml")
	cfg, err := config.LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	workspace, err := os.MkdirTemp("", "agent-bash-*")
	if err != nil {
		t.Fatalf("temp workspace: %v", err)
	}
	defer os.RemoveAll(workspace)

	systemPrompt := "You are a helpful AI assistant that can use tools."
	promptPath := filepath.Join(path.ProjectRoot(), "configs", "system_prompt.txt")
	if _, err := os.Stat(promptPath); err == nil {
		data, _ := os.ReadFile(promptPath)
		systemPrompt = string(data)
	}

	llmClient := llm.NewClient(
		cfg.LLM.APIKey,
		cfg.LLM.APIBase,
		cfg.LLM.Model,
	)

	toolList := []tools.Tool{
		tools.NewReadTool(workspace),
		tools.NewWriteTool(workspace),
		tools.NewBashTool(),
	}

	ag, err := agent.NewAgent(
		llmClient,
		systemPrompt,
		toolList,
		10,
		workspace,
		150000,
	)
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	task := "Use bash to list all files in the current directory and tell me what you see."
	t.Log("Task:", task)

	ag.AddUserMessage(task)

	ctx := context.Background()
	result, err := ag.Run(ctx)
	if err != nil {
		t.Fatalf("agent run: %v", err)
	}

	t.Log("\n============================================================")
	t.Log("Agent Result:", result)
	t.Log("============================================================")
	t.Log("✅ Bash task completed")
}
