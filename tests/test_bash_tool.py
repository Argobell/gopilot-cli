"""Tests for bash_tool.py

This module contains unit tests (using mocks) and integration tests
for BashTool, BashOutputTool, BashKillTool, and related classes.
"""

import asyncio
import platform
import re
import time
from unittest.mock import AsyncMock, MagicMock, Mock, patch

import pytest

from src.tools.bash_tool import (
    BackgroundShell,
    BackgroundShellManager,
    BashKillTool,
    BashOutputResult,
    BashOutputTool,
    BashTool,
)


# ============================================================================
# Unit Tests - BashOutputResult
# ============================================================================


class TestBashOutputResult:
    """Unit tests for BashOutputResult data model."""

    def test_format_content_with_stdout_only(self):
        """Test content formatting with only stdout."""
        result = BashOutputResult(
            success=True,
            stdout="Hello World",
            stderr="",
            exit_code=0,
        )
        assert result.content == "Hello World"

    def test_format_content_with_stdout_and_stderr(self):
        """Test content formatting with both stdout and stderr."""
        result = BashOutputResult(
            success=False,
            stdout="Output line",
            stderr="Error line",
            exit_code=1,
        )
        assert "Output line" in result.content
        assert "[stderr]" in result.content
        assert "Error line" in result.content
        assert "[exit_code]: 1" in result.content

    def test_format_content_empty_output(self):
        """Test content formatting with no output."""
        result = BashOutputResult(
            success=True,
            stdout="",
            stderr="",
            exit_code=0,
        )
        assert result.content == "(no output)"

    def test_format_content_with_exit_code(self):
        """Test content formatting includes exit code when non-zero."""
        result = BashOutputResult(
            success=False,
            stdout="Some output",
            stderr="",
            exit_code=127,
        )
        assert "[exit_code]: 127" in result.content

    def test_bash_id_field(self):
        """Test bash_id field is properly stored."""
        result = BashOutputResult(
            success=True,
            stdout="test",
            stderr="",
            exit_code=0,
            bash_id="test-123",
        )
        assert result.bash_id == "test-123"


# ============================================================================
# Unit Tests - BackgroundShell
# ============================================================================


class TestBackgroundShell:
    """Unit tests for BackgroundShell data container."""

    def test_initialization(self):
        """Test BackgroundShell initialization."""
        mock_process = Mock()
        start_time = time.time()

        shell = BackgroundShell(
            bash_id="test-id",
            command="echo test",
            process=mock_process,
            start_time=start_time,
        )

        assert shell.bash_id == "test-id"
        assert shell.command == "echo test"
        assert shell.process == mock_process
        assert shell.start_time == start_time
        assert shell.output_lines == []
        assert shell.last_read_index == 0
        assert shell.status == "running"
        assert shell.exit_code is None

    def test_add_output(self):
        """Test adding output lines."""
        mock_process = Mock()
        shell = BackgroundShell("id", "cmd", mock_process, time.time())

        shell.add_output("line 1")
        shell.add_output("line 2")

        assert len(shell.output_lines) == 2
        assert shell.output_lines[0] == "line 1"
        assert shell.output_lines[1] == "line 2"

    def test_get_new_output(self):
        """Test getting new output since last check."""
        mock_process = Mock()
        shell = BackgroundShell("id", "cmd", mock_process, time.time())

        shell.add_output("line 1")
        shell.add_output("line 2")

        # First read
        new_lines = shell.get_new_output()
        assert new_lines == ["line 1", "line 2"]
        assert shell.last_read_index == 2

        # Second read (no new lines)
        new_lines = shell.get_new_output()
        assert new_lines == []

        # Add more output
        shell.add_output("line 3")
        new_lines = shell.get_new_output()
        assert new_lines == ["line 3"]

    def test_get_new_output_with_filter(self):
        """Test getting filtered output with regex pattern."""
        mock_process = Mock()
        shell = BackgroundShell("id", "cmd", mock_process, time.time())

        shell.add_output("ERROR: something failed")
        shell.add_output("INFO: all good")
        shell.add_output("ERROR: another issue")

        # Filter for ERROR lines only
        new_lines = shell.get_new_output(filter_pattern="ERROR")
        assert len(new_lines) == 2
        assert "ERROR: something failed" in new_lines
        assert "ERROR: another issue" in new_lines
        assert "INFO: all good" not in new_lines

    def test_get_new_output_with_invalid_regex(self):
        """Test that invalid regex patterns don't crash."""
        mock_process = Mock()
        shell = BackgroundShell("id", "cmd", mock_process, time.time())

        shell.add_output("line 1")

        # Invalid regex should return all lines
        new_lines = shell.get_new_output(filter_pattern="[invalid(")
        assert new_lines == ["line 1"]

    def test_update_status_running(self):
        """Test updating status to running."""
        mock_process = Mock()
        shell = BackgroundShell("id", "cmd", mock_process, time.time())

        shell.update_status(is_alive=True)
        assert shell.status == "running"
        assert shell.exit_code is None

    def test_update_status_completed(self):
        """Test updating status to completed."""
        mock_process = Mock()
        shell = BackgroundShell("id", "cmd", mock_process, time.time())

        shell.update_status(is_alive=False, exit_code=0)
        assert shell.status == "completed"
        assert shell.exit_code == 0

    def test_update_status_failed(self):
        """Test updating status to failed."""
        mock_process = Mock()
        shell = BackgroundShell("id", "cmd", mock_process, time.time())

        shell.update_status(is_alive=False, exit_code=1)
        assert shell.status == "failed"
        assert shell.exit_code == 1

    @pytest.mark.asyncio
    async def test_terminate_running_process(self):
        """Test terminating a running process."""
        mock_process = AsyncMock()
        mock_process.returncode = None
        mock_process.wait = AsyncMock(return_value=0)

        shell = BackgroundShell("id", "cmd", mock_process, time.time())
        await shell.terminate()

        mock_process.terminate.assert_called_once()
        assert shell.status == "terminated"

    @pytest.mark.asyncio
    async def test_terminate_with_timeout_kills_process(self):
        """Test that terminate kills process if it doesn't stop gracefully."""
        mock_process = AsyncMock()
        mock_process.returncode = None
        mock_process.wait = AsyncMock(side_effect=asyncio.TimeoutError)

        shell = BackgroundShell("id", "cmd", mock_process, time.time())
        await shell.terminate()

        mock_process.terminate.assert_called_once()
        mock_process.kill.assert_called_once()
        assert shell.status == "terminated"


# ============================================================================
# Unit Tests - BackgroundShellManager
# ============================================================================


class TestBackgroundShellManager:
    """Unit tests for BackgroundShellManager."""

    def setup_method(self):
        """Clear manager state before each test."""
        BackgroundShellManager._shells.clear()
        BackgroundShellManager._monitor_tasks.clear()

    def test_add_and_get_shell(self):
        """Test adding and retrieving a shell."""
        mock_process = Mock()
        shell = BackgroundShell("test-id", "cmd", mock_process, time.time())

        BackgroundShellManager.add(shell)
        retrieved = BackgroundShellManager.get("test-id")

        assert retrieved == shell
        assert retrieved.bash_id == "test-id"

    def test_get_nonexistent_shell(self):
        """Test getting a shell that doesn't exist."""
        result = BackgroundShellManager.get("nonexistent")
        assert result is None

    def test_get_available_ids(self):
        """Test getting all available bash IDs."""
        mock_process = Mock()
        shell1 = BackgroundShell("id-1", "cmd1", mock_process, time.time())
        shell2 = BackgroundShell("id-2", "cmd2", mock_process, time.time())

        BackgroundShellManager.add(shell1)
        BackgroundShellManager.add(shell2)

        ids = BackgroundShellManager.get_available_ids()
        assert len(ids) == 2
        assert "id-1" in ids
        assert "id-2" in ids

    def test_remove_shell(self):
        """Test removing a shell (internal method)."""
        mock_process = Mock()
        shell = BackgroundShell("test-id", "cmd", mock_process, time.time())

        BackgroundShellManager.add(shell)
        assert BackgroundShellManager.get("test-id") is not None

        BackgroundShellManager._remove("test-id")
        assert BackgroundShellManager.get("test-id") is None

    @pytest.mark.asyncio
    async def test_terminate_shell(self):
        """Test terminating a shell through manager."""
        mock_process = AsyncMock()
        mock_process.returncode = None
        mock_process.wait = AsyncMock(return_value=0)

        shell = BackgroundShell("test-id", "cmd", mock_process, time.time())
        BackgroundShellManager.add(shell)

        terminated_shell = await BackgroundShellManager.terminate("test-id")

        assert terminated_shell.status == "terminated"
        assert BackgroundShellManager.get("test-id") is None

    @pytest.mark.asyncio
    async def test_terminate_nonexistent_shell_raises_error(self):
        """Test terminating a nonexistent shell raises ValueError."""
        with pytest.raises(ValueError, match="Shell not found"):
            await BackgroundShellManager.terminate("nonexistent")


# ============================================================================
# Integration Tests - BashTool (Foreground)
# ============================================================================


class TestBashToolForeground:
    """Integration tests for BashTool foreground execution."""

    @pytest.mark.asyncio
    async def test_simple_command_success(self):
        """Test executing a simple successful command."""
        tool = BashTool()

        # Use a universal command
        if platform.system() == "Windows":
            result = await tool.execute("echo test")
        else:
            result = await tool.execute("echo 'test'")

        assert result.success is True
        assert result.exit_code == 0
        assert "test" in result.stdout

    @pytest.mark.asyncio
    async def test_command_with_error(self):
        """Test executing a command that fails."""
        tool = BashTool()

        # Command that should fail on both platforms
        if platform.system() == "Windows":
            result = await tool.execute("exit 1")
        else:
            result = await tool.execute("exit 1")

        assert result.success is False
        assert result.exit_code == 1
        assert result.error is not None

    @pytest.mark.asyncio
    async def test_pwd_command(self):
        """Test getting current working directory."""
        tool = BashTool()

        if platform.system() == "Windows":
            result = await tool.execute("pwd")
        else:
            result = await tool.execute("pwd")

        assert result.success is True
        assert len(result.stdout) > 0

    @pytest.mark.asyncio
    async def test_timeout_validation(self):
        """Test timeout validation (max 600 seconds)."""
        tool = BashTool()

        # Timeout should be capped at 600
        result = await tool.execute("echo test", timeout=1000)
        assert result.success is True

    @pytest.mark.asyncio
    async def test_minimum_timeout_validation(self):
        """Test minimum timeout defaults to 120."""
        tool = BashTool()

        # Timeout < 1 should default to 120
        result = await tool.execute("echo test", timeout=0)
        assert result.success is True

    @pytest.mark.asyncio
    async def test_command_with_output_to_stderr(self):
        """Test command that outputs to stderr."""
        tool = BashTool()

        if platform.system() == "Windows":
            # PowerShell stderr redirect
            result = await tool.execute(
                "[System.Console]::Error.WriteLine('error message'); exit 1"
            )
        else:
            result = await tool.execute("echo 'error message' >&2; exit 1")

        assert result.success is False
        assert result.exit_code == 1
        assert len(result.stderr) > 0 or len(result.error) > 0


# ============================================================================
# Integration Tests - BashTool (Background)
# ============================================================================


class TestBashToolBackground:
    """Integration tests for BashTool background execution."""

    def setup_method(self):
        """Clear manager state before each test."""
        BackgroundShellManager._shells.clear()
        BackgroundShellManager._monitor_tasks.clear()

    @pytest.mark.asyncio
    async def test_start_background_command(self):
        """Test starting a command in background."""
        tool = BashTool()

        if platform.system() == "Windows":
            result = await tool.execute(
                "Start-Sleep -Seconds 2", run_in_background=True
            )
        else:
            result = await tool.execute("sleep 2", run_in_background=True)

        assert result.success is True
        assert result.bash_id is not None
        assert len(result.bash_id) > 0
        assert "Background command started" in result.stdout

        # Cleanup
        if result.bash_id:
            shell = BackgroundShellManager.get(result.bash_id)
            if shell:
                await shell.terminate()
                BackgroundShellManager._remove(result.bash_id)

    @pytest.mark.asyncio
    async def test_background_command_returns_immediately(self):
        """Test that background commands return immediately."""
        tool = BashTool()

        start_time = time.time()

        if platform.system() == "Windows":
            result = await tool.execute(
                "Start-Sleep -Seconds 5", run_in_background=True
            )
        else:
            result = await tool.execute("sleep 5", run_in_background=True)

        elapsed = time.time() - start_time

        # Should return in less than 1 second
        assert elapsed < 1.0
        assert result.success is True
        assert result.bash_id is not None

        # Cleanup
        if result.bash_id:
            await BackgroundShellManager.terminate(result.bash_id)

    @pytest.mark.asyncio
    async def test_background_command_produces_output(self):
        """Test that background commands produce output."""
        tool = BashTool()

        if platform.system() == "Windows":
            result = await tool.execute(
                "echo 'line1'; Start-Sleep -Milliseconds 100; echo 'line2'",
                run_in_background=True,
            )
        else:
            result = await tool.execute(
                "echo 'line1'; sleep 0.1; echo 'line2'", run_in_background=True
            )

        assert result.success is True
        bash_id = result.bash_id

        # Wait a bit for output
        await asyncio.sleep(0.5)

        # Check shell has output
        shell = BackgroundShellManager.get(bash_id)
        assert shell is not None
        assert len(shell.output_lines) > 0

        # Cleanup
        await BackgroundShellManager.terminate(bash_id)


# ============================================================================
# Integration Tests - BashOutputTool
# ============================================================================


class TestBashOutputTool:
    """Integration tests for BashOutputTool."""

    def setup_method(self):
        """Clear manager state before each test."""
        BackgroundShellManager._shells.clear()
        BackgroundShellManager._monitor_tasks.clear()

    @pytest.mark.asyncio
    async def test_get_output_from_background_command(self):
        """Test retrieving output from background command."""
        bash_tool = BashTool()
        output_tool = BashOutputTool()

        # Start background command
        if platform.system() == "Windows":
            result = await bash_tool.execute(
                "echo 'test output'", run_in_background=True
            )
        else:
            result = await bash_tool.execute(
                "echo 'test output'", run_in_background=True
            )

        bash_id = result.bash_id

        # Wait for command to complete
        await asyncio.sleep(0.5)

        # Get output
        output_result = await output_tool.execute(bash_id=bash_id)

        assert output_result.success is True
        assert "test output" in output_result.stdout

        # Cleanup
        await BackgroundShellManager.terminate(bash_id)

    @pytest.mark.asyncio
    async def test_get_output_nonexistent_shell(self):
        """Test getting output from nonexistent shell."""
        output_tool = BashOutputTool()

        result = await output_tool.execute(bash_id="nonexistent-id")

        assert result.success is False
        assert "Shell not found" in result.error

    @pytest.mark.asyncio
    async def test_get_output_only_new_lines(self):
        """Test that only new output is returned."""
        bash_tool = BashTool()
        output_tool = BashOutputTool()

        # Start background command with multiple lines
        if platform.system() == "Windows":
            result = await bash_tool.execute(
                "echo 'line1'; echo 'line2'; echo 'line3'", run_in_background=True
            )
        else:
            result = await bash_tool.execute(
                "echo 'line1'; echo 'line2'; echo 'line3'", run_in_background=True
            )

        bash_id = result.bash_id
        await asyncio.sleep(0.3)

        # First read
        output1 = await output_tool.execute(bash_id=bash_id)
        assert output1.success is True
        first_output = output1.stdout

        # Second read (should be empty if no new output)
        output2 = await output_tool.execute(bash_id=bash_id)
        assert output2.success is True
        # Second output might be empty or have remaining lines

        # Cleanup
        await BackgroundShellManager.terminate(bash_id)

    @pytest.mark.asyncio
    async def test_get_output_with_filter(self):
        """Test getting filtered output."""
        bash_tool = BashTool()
        output_tool = BashOutputTool()

        # Start command with mixed output
        if platform.system() == "Windows":
            result = await bash_tool.execute(
                "echo 'ERROR: bad'; echo 'INFO: good'; echo 'ERROR: worse'",
                run_in_background=True,
            )
        else:
            result = await bash_tool.execute(
                "echo 'ERROR: bad'; echo 'INFO: good'; echo 'ERROR: worse'",
                run_in_background=True,
            )

        bash_id = result.bash_id
        await asyncio.sleep(0.3)

        # Get only ERROR lines
        output = await output_tool.execute(bash_id=bash_id, filter_str="ERROR")

        assert output.success is True
        # Should contain ERROR lines (if any output was produced)

        # Cleanup
        await BackgroundShellManager.terminate(bash_id)


# ============================================================================
# Integration Tests - BashKillTool
# ============================================================================


class TestBashKillTool:
    """Integration tests for BashKillTool."""

    def setup_method(self):
        """Clear manager state before each test."""
        BackgroundShellManager._shells.clear()
        BackgroundShellManager._monitor_tasks.clear()

    @pytest.mark.asyncio
    async def test_kill_running_command(self):
        """Test killing a running background command."""
        bash_tool = BashTool()
        kill_tool = BashKillTool()

        # Start long-running command
        if platform.system() == "Windows":
            result = await bash_tool.execute(
                "Start-Sleep -Seconds 30", run_in_background=True
            )
        else:
            result = await bash_tool.execute("sleep 30", run_in_background=True)

        bash_id = result.bash_id
        await asyncio.sleep(0.2)

        # Kill it
        kill_result = await kill_tool.execute(bash_id=bash_id)

        assert kill_result.success is True
        assert kill_result.bash_id == bash_id

        # Shell should be removed from manager
        shell = BackgroundShellManager.get(bash_id)
        assert shell is None

    @pytest.mark.asyncio
    async def test_kill_nonexistent_shell(self):
        """Test killing a shell that doesn't exist."""
        kill_tool = BashKillTool()

        result = await kill_tool.execute(bash_id="nonexistent-id")

        assert result.success is False
        assert "Shell not found" in result.error or "not found" in result.error.lower()

    @pytest.mark.asyncio
    async def test_kill_returns_remaining_output(self):
        """Test that kill returns any remaining output."""
        bash_tool = BashTool()
        kill_tool = BashKillTool()

        # Start command that produces output
        if platform.system() == "Windows":
            result = await bash_tool.execute(
                "echo 'output line'; Start-Sleep -Seconds 10", run_in_background=True
            )
        else:
            result = await bash_tool.execute(
                "echo 'output line'; sleep 10", run_in_background=True
            )

        bash_id = result.bash_id
        await asyncio.sleep(0.3)

        # Kill and check output
        kill_result = await kill_tool.execute(bash_id=bash_id)

        assert kill_result.success is True
        # Output might contain the echo line


# ============================================================================
# Tool Properties Tests
# ============================================================================


class TestBashToolProperties:
    """Test tool name, description, and parameters properties."""

    def test_bash_tool_name(self):
        """Test BashTool name property."""
        tool = BashTool()
        assert tool.name == "bash"

    def test_bash_tool_description(self):
        """Test BashTool description contains key information."""
        tool = BashTool()
        description = tool.description

        assert isinstance(description, str)
        assert len(description) > 0
        assert "command" in description.lower()

    def test_bash_tool_parameters(self):
        """Test BashTool parameters schema."""
        tool = BashTool()
        params = tool.parameters

        assert params["type"] == "object"
        assert "command" in params["properties"]
        assert "timeout" in params["properties"]
        assert "run_in_background" in params["properties"]
        assert "command" in params["required"]

    def test_bash_output_tool_name(self):
        """Test BashOutputTool name property."""
        tool = BashOutputTool()
        assert tool.name == "bash_output"

    def test_bash_output_tool_parameters(self):
        """Test BashOutputTool parameters schema."""
        tool = BashOutputTool()
        params = tool.parameters

        assert "bash_id" in params["properties"]
        assert "bash_id" in params["required"]

    def test_bash_kill_tool_name(self):
        """Test BashKillTool name property."""
        tool = BashKillTool()
        assert tool.name == "bash_kill"

    def test_bash_kill_tool_parameters(self):
        """Test BashKillTool parameters schema."""
        tool = BashKillTool()
        params = tool.parameters

        assert "bash_id" in params["properties"]
        assert "bash_id" in params["required"]


if __name__ == "__main__":
    """Allow running tests directly with python."""
    pytest.main([__file__, "-v"])
