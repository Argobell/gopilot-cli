package tests

import (
	"strings"
	"testing"

	tw "gopilot-cli/internal/utils/terminal"
)

// ------------------------
// TestCalculateDisplayWidth
// ------------------------

func TestCalculateDisplayWidth_ASCII(t *testing.T) {
	if tw.CalculateDisplayWidth("Hello") != 5 {
		t.Errorf("expected 5")
	}
	if tw.CalculateDisplayWidth("World") != 5 {
		t.Errorf("expected 5")
	}
	if tw.CalculateDisplayWidth("Test 123") != 8 {
		t.Errorf("expected 8")
	}
}

func TestCalculateDisplayWidth_Empty(t *testing.T) {
	if tw.CalculateDisplayWidth("") != 0 {
		t.Errorf("expected 0")
	}
}

func TestCalculateDisplayWidth_Emoji(t *testing.T) {
	if tw.CalculateDisplayWidth("🤖") != 2 {
		t.Errorf("expected emoji = 2")
	}
	if tw.CalculateDisplayWidth("💭") != 2 {
		t.Errorf("expected emoji = 2")
	}
	if tw.CalculateDisplayWidth("🤖 Agent") != 8 { // 2 + 1 + 5
		t.Errorf("expected 8")
	}
}

func TestCalculateDisplayWidth_Chinese(t *testing.T) {
	if tw.CalculateDisplayWidth("你好") != 4 {
		t.Errorf("expected 4")
	}
	if tw.CalculateDisplayWidth("你好世界") != 8 {
		t.Errorf("expected 8")
	}
	if tw.CalculateDisplayWidth("中文") != 4 {
		t.Errorf("expected 4")
	}
}

func TestCalculateDisplayWidth_Japanese(t *testing.T) {
	if tw.CalculateDisplayWidth("日本語") != 6 { // 3 chars * 2
		t.Errorf("expected 6")
	}
}

func TestCalculateDisplayWidth_Mixed(t *testing.T) {
	if tw.CalculateDisplayWidth("Hello 你好") != 10 { // 5 +1 +4
		t.Errorf("expected 10")
	}
	if tw.CalculateDisplayWidth("Test 🤖") != 7 { // 4 +1 +2
		t.Errorf("expected 7")
	}
}

func TestCalculateDisplayWidth_ANSI(t *testing.T) {
	if tw.CalculateDisplayWidth("\033[31mRed\033[0m") != 3 {
		t.Errorf("expected 3")
	}
	if tw.CalculateDisplayWidth("\033[31m🤖\033[0m") != 2 {
		t.Errorf("expected 2")
	}
}

func TestCalculateDisplayWidth_Combining(t *testing.T) {
	e := "e\u0301" // e + combining accent
	if tw.CalculateDisplayWidth(e) != 1 {
		t.Errorf("expected 1")
	}
}

func TestCalculateDisplayWidth_ComplexANSI(t *testing.T) {
	text := "\033[1m\033[36mBold Cyan\033[0m"
	if tw.CalculateDisplayWidth(text) != 9 {
		t.Errorf("expected 9")
	}
}

// ------------------------
// Test TruncateWithEllipsis
// ------------------------

func TestTruncate_NoNeed(t *testing.T) {
	if tw.TruncateWithEllipsis("Hello", 10) != "Hello" {
		t.Errorf("no truncation expected")
	}
	if tw.TruncateWithEllipsis("Test", 5) != "Test" {
		t.Errorf("no truncation expected")
	}
}

func TestTruncate_ExactFit(t *testing.T) {
	if tw.TruncateWithEllipsis("Hello", 5) != "Hello" {
		t.Errorf("exact-width expected")
	}
}

func TestTruncate_ASCII(t *testing.T) {
	if tw.TruncateWithEllipsis("Hello World", 8) != "Hello W…" {
		t.Errorf("expected 'Hello W…'")
	}
	if tw.TruncateWithEllipsis("Testing", 4) != "Tes…" {
		t.Errorf("expected 'Tes…'")
	}
}

func TestTruncate_Chinese(t *testing.T) {
	r := tw.TruncateWithEllipsis("你好世界", 5)
	if tw.CalculateDisplayWidth(r) > 5 {
		t.Errorf("width overflow")
	}
	if !strings.Contains(r, "…") {
		t.Errorf("ellipsis missing")
	}
}

func TestTruncate_Emoji(t *testing.T) {
	r := tw.TruncateWithEllipsis("🤖🤖🤖", 3)
	if tw.CalculateDisplayWidth(r) > 3 {
		t.Errorf("should fit width <= 3")
	}
}

func TestTruncate_Zero(t *testing.T) {
	if tw.TruncateWithEllipsis("Hello", 0) != "" {
		t.Errorf("expected empty")
	}
}

func TestTruncate_WidthOne(t *testing.T) {
	r := tw.TruncateWithEllipsis("Hello", 1)
	if len([]rune(r)) > 1 {
		t.Errorf("result too long")
	}
}

func TestTruncate_ANSI(t *testing.T) {
	colored := "\033[31mHello World\033[0m"
	r := tw.TruncateWithEllipsis(colored, 8)
	if strings.Contains(r, "\033[") {
		t.Errorf("ANSI codes should be stripped")
	}
	if !strings.Contains(r, "…") {
		t.Errorf("ellipsis missing")
	}
}

// ------------------------
// Test PadToWidth
// ------------------------

func TestPad_LeftAlign(t *testing.T) {
	r := tw.PadToWidth("Hello", 10, "left", ' ')
	if r != "Hello     " {
		t.Errorf("expected left padded result")
	}
}

func TestPad_RightAlign(t *testing.T) {
	r := tw.PadToWidth("Hello", 10, "right", ' ')
	if r != "     Hello" {
		t.Errorf("expected right padding")
	}
}

func TestPad_Center(t *testing.T) {
	r := tw.PadToWidth("Test", 10, "center", ' ')
	if r != "   Test   " {
		t.Errorf("expected center alignment")
	}
}

func TestPad_CenterOdd(t *testing.T) {
	r := tw.PadToWidth("Hi", 7, "center", ' ')
	if len(r) != 7 || !strings.Contains(r, "Hi") {
		t.Errorf("unexpected center alignment")
	}
}

func TestPad_Chinese(t *testing.T) {
	r := tw.PadToWidth("你好", 10, "left", ' ')
	if tw.CalculateDisplayWidth(r) != 10 {
		t.Errorf("expected width 10")
	}
}

func TestPad_Emoji(t *testing.T) {
	r := tw.PadToWidth("🤖", 10, "left", ' ')
	if tw.CalculateDisplayWidth(r) != 10 {
		t.Errorf("expected width 10")
	}
}

func TestPad_NoNeed(t *testing.T) {
	r := tw.PadToWidth("Hello", 5, "left", ' ')
	if r != "Hello" {
		t.Errorf("expected unchanged")
	}
}

func TestPad_TextExceeds(t *testing.T) {
	r := tw.PadToWidth("Hello World", 5, "left", ' ')
	if r != "Hello World" {
		t.Errorf("over max → return as-is")
	}
}

func TestPad_InvalidAlign(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("expected panic for invalid align")
		}
	}()
	tw.PadToWidth("Test", 10, "invalid", ' ')
}

func TestPad_CustomFill(t *testing.T) {
	r := tw.PadToWidth("Test", 10, "left", '-')
	if r != "Test------" {
		t.Errorf("expected custom fill")
	}
}

// ------------------------
// Real-world scenario tests
// ------------------------

func TestReal_StepHeader(t *testing.T) {
	text := "💭 Step 1/50"
	if tw.CalculateDisplayWidth(text) != 12 {
		t.Errorf("expected width=12")
	}
}

func TestReal_ModelLine(t *testing.T) {
	line := "Model: minimax-01"
	if tw.CalculateDisplayWidth(line) <= 0 {
		t.Errorf("invalid width")
	}
}

func TestReal_ChineseModel(t *testing.T) {
	line := "Model: 模型-01"
	if tw.CalculateDisplayWidth(line) != 14 {
		t.Errorf("expected 14")
	}
}

func TestReal_Banner(t *testing.T) {
	banner := "🤖 Mini Agent - Multi-turn Interactive Session"
	if tw.CalculateDisplayWidth(banner) != 46 {
		t.Errorf("expected 46")
	}
}
