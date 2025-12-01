package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"gopilot-cli/internal/tools"
)

// helper small sleep for background output
func wait() { time.Sleep(600 * time.Millisecond) }

func isWindows() bool { return runtime.GOOS == "windows" }

// =======================================
// Foreground command
// =======================================

func TestForegroundCommand(t *testing.T) {
	bash := tools.NewBashTool()

	res, err := bash.Execute(context.Background(), map[string]any{
		"command": "echo 'Hello from foreground'",
	})
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if !res.Success {
		t.Fatalf("Expected success, got error: %s", res.Error)
	}
	if res.ExitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", res.ExitCode)
	}
	if res.Stdout == "" {
		t.Fatalf("Unexpected empty stdout")
	}
}

// =======================================
// Stdout + Stderr
// =======================================

func TestForegroundCommandWithStderr(t *testing.T) {
	bash := tools.NewBashTool()

	var command string
	if isWindows() {
		// PowerShell: Write-Error 会导致退出码=1，所以 Success = false 是正常行为
		command = `Write-Output "stdout message"; Write-Error "stderr message"`
	} else {
		command = `echo "stdout message" && echo "stderr message" >&2`
	}

	res, err := bash.Execute(context.Background(), map[string]any{
		"command": command,
	})
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	// ---------- Windows 特殊逻辑 ----------
	if isWindows() {
		// PowerShell Write-Error 会让 Success=false → 这是预期
		if res.Stdout == "" || !contains(res.Stdout, "stdout message") {
			t.Fatalf("stdout missing: %q", res.Stdout)
		}
		if res.Stderr == "" || !contains(res.Stderr, "stderr message") {
			t.Fatalf("stderr missing: %q", res.Stderr)
		}
		return
	}

	// ---------- Linux/macOS 逻辑 ----------
	if !res.Success {
		t.Fatalf("Expected success but failed: %v", res.Error)
	}
	if !contains(res.Stdout, "stdout message") {
		t.Fatalf("stdout missing: %q", res.Stdout)
	}
	if !contains(res.Stderr, "stderr message") {
		t.Fatalf("stderr missing: %q", res.Stderr)
	}
}

func contains(s, sub string) bool { return len(s) > 0 && stringContains(s, sub) }
func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s[:len(sub)] == sub || (len(s) > len(sub) && stringContains(s[1:], sub)))
}

// =======================================
// Failure
// =======================================

func TestCommandFailure(t *testing.T) {
	bash := tools.NewBashTool()

	res, err := bash.Execute(context.Background(), map[string]any{
		"command": "ls /nonexistent_12345",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Success {
		t.Fatalf("Expected failure, got success")
	}
	if res.ExitCode == 0 {
		t.Fatalf("Expected non-zero exit code")
	}
	if res.Error == "" {
		t.Fatalf("Expected error message")
	}
}

// =======================================
// Timeout
// =======================================

func TestCommandTimeout(t *testing.T) {
	bash := tools.NewBashTool()

	res, err := bash.Execute(context.Background(), map[string]any{
		"command": "sleep 5",
		"timeout": 1,
	})

	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if res.Success {
		t.Fatalf("Expected timeout failure")
	}

	// Windows: exit code 1 is normal
	if isWindows() {
		if res.ExitCode == 0 {
			t.Fatalf("Expected non-zero exit code for Windows timeout")
		}
	} else {
		if res.ExitCode != -1 {
			t.Fatalf("Expected exit_code -1, got %d", res.ExitCode)
		}
	}
}

// =======================================
// Background run
// =======================================

func TestBackgroundCommand(t *testing.T) {
	bash := tools.NewBashTool()

	res, err := bash.Execute(context.Background(), map[string]any{
		"command":           "for i in 1 2 3; do echo Line-$i; sleep 0.2; done",
		"run_in_background": true,
	})
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}
	if !res.Success {
		t.Fatalf("Unexpected failure: %v", res.Error)
	}
	if res.BashID == "" {
		t.Fatalf("Missing bash_id")
	}

	bashID := res.BashID
	output := tools.NewBashOutputTool()

	wait()

	outRes, _ := output.Execute(context.Background(), map[string]any{
		"bash_id": bashID,
	})

	if !outRes.Success {
		t.Fatalf("Output read failed: %v", outRes.Error)
	}
	if outRes.Stdout == "" {
		t.Fatalf("Expected some output")
	}

	tools.NewBashKillTool().Execute(context.Background(), map[string]any{
		"bash_id": bashID,
	})
}

// =======================================
// Incremental output
// =======================================

func TestBashOutputMonitoring(t *testing.T) {
	bash := tools.NewBashTool()

	res, _ := bash.Execute(context.Background(), map[string]any{
		"command":           "for i in 1 2 3 4; do echo Line-$i; sleep 0.2; done",
		"run_in_background": true,
	})

	bashID := res.BashID
	out := tools.NewBashOutputTool()

	for i := 0; i < 3; i++ {
		wait()
		r, _ := out.Execute(context.Background(), map[string]any{
			"bash_id": bashID,
		})
		if !r.Success {
			t.Fatalf("Failed to read output")
		}
	}

	tools.NewBashKillTool().Execute(context.Background(), map[string]any{
		"bash_id": bashID,
	})
}

// =======================================
// Filter
// =======================================

func TestBashOutputFilter(t *testing.T) {
	bash := tools.NewBashTool()

	var command string
	if isWindows() {
		command = `for ($i=1; $i -le 5; $i++) { Write-Output "Line $i"; Start-Sleep -Milliseconds 300 }`
	} else {
		command = `for i in 1 2 3 4 5; do echo "Line $i"; sleep 0.3; done`
	}

	res, _ := bash.Execute(context.Background(), map[string]any{
		"command":           command,
		"run_in_background": true,
	})

	bashID := res.BashID
	wait()

	out := tools.NewBashOutputTool()

	r, _ := out.Execute(context.Background(), map[string]any{
		"bash_id":    bashID,
		"filter_str": `Line (2|4)`,
	})

	if !r.Success {
		t.Fatalf("Filter failed: %v", r.Error)
	}

	if !contains(r.Stdout, "Line 2") && !contains(r.Stdout, "Line 4") {
		t.Fatalf("Filtered output incorrect: %q", r.Stdout)
	}

	tools.NewBashKillTool().Execute(context.Background(), map[string]any{
		"bash_id": bashID,
	})
}

// =======================================
// Kill background task
// =======================================

func TestBashKill(t *testing.T) {
	bash := tools.NewBashTool()

	res, _ := bash.Execute(context.Background(), map[string]any{
		"command":           "sleep 99",
		"run_in_background": true,
	})
	bashID := res.BashID

	kill := tools.NewBashKillTool()
	k, _ := kill.Execute(context.Background(), map[string]any{
		"bash_id": bashID,
	})

	if !k.Success {
		t.Fatalf("Kill failed: %v", k.Error)
	}
}

// =======================================
// Kill nonexistent
// =======================================

func TestBashKillNonexistent(t *testing.T) {
	kill := tools.NewBashKillTool()

	res, _ := kill.Execute(context.Background(), map[string]any{
		"bash_id": "invalid123",
	})

	if res.Success {
		t.Fatalf("Expected failure when killing non-existent shell")
	}
}

// =======================================
// Output nonexistent
// =======================================

func TestBashOutputNonexistent(t *testing.T) {
	out := tools.NewBashOutputTool()

	res, _ := out.Execute(context.Background(), map[string]any{
		"bash_id": "invalid123",
	})

	if res.Success {
		t.Fatalf("Expected failure for nonexistent output")
	}
}

// =======================================
// Multiple background
// =======================================

func TestMultipleBackgroundCommands(t *testing.T) {
	bash := tools.NewBashTool()
	ids := []string{}

	for i := 0; i < 3; i++ {
		r, _ := bash.Execute(context.Background(), map[string]any{
			"command":           "for j in 1 2; do echo CMD-" + string(rune('A'+i)) + "-$j; sleep 0.1; done",
			"run_in_background": true,
		})
		ids = append(ids, r.BashID)
	}

	wait()
	out := tools.NewBashOutputTool()

	for _, id := range ids {
		r, _ := out.Execute(context.Background(), map[string]any{"bash_id": id})
		if !r.Success {
			t.Fatalf("Output read failed for %s", id)
		}
	}

	for _, id := range ids {
		tools.NewBashKillTool().Execute(context.Background(), map[string]any{"bash_id": id})
	}
}

// =======================================
// Timeout validation (clamp)
// =======================================

func TestTimeoutValidation(t *testing.T) {
	bash := tools.NewBashTool()

	r, _ := bash.Execute(context.Background(), map[string]any{
		"command": "echo test",
		"timeout": 1000,
	})
	if !r.Success {
		t.Fatalf("Unexpected failure for timeout>600")
	}

	r2, _ := bash.Execute(context.Background(), map[string]any{
		"command": "echo test",
		"timeout": 0,
	})
	if !r2.Success {
		t.Fatalf("Unexpected failure for timeout<1")
	}
}
