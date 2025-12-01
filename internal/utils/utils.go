package terminalwidth

import (
	"regexp"
	"unicode"

	"golang.org/x/text/width"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func runeWidth(r rune) int {
	switch {
	case unicode.Is(unicode.Mn, r):
		return 0
	case isEmoji(r):
		return 2
	}

	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	}
	return 1
}

func isEmoji(r rune) bool {
	return r >= 0x1F300 && r <= 0x1FAFF
}

func CalculateDisplayWidth(s string) int {
	s = ansiEscape.ReplaceAllString(s, "")
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func TruncateWithEllipsis(text string, maxWidth int, ellipsis ...string) string {
	if maxWidth <= 0 {
		return ""
	}

	e := "…"
	if len(ellipsis) > 0 && ellipsis[0] != "" {
		e = ellipsis[0]
	}

	plain := ansiEscape.ReplaceAllString(text, "")
	if CalculateDisplayWidth(plain) <= maxWidth {
		return text
	}

	eWidth := CalculateDisplayWidth(e)
	if maxWidth <= eWidth {
		return truncateWidth(plain, maxWidth)
	}

	available := maxWidth - eWidth
	return truncateWidth(plain, available) + e
}

func truncateWidth(s string, max int) string {
	w := 0
	out := make([]rune, 0, len(s))
	for _, r := range s {
		rw := runeWidth(r)
		if w+rw > max {
			break
		}
		out = append(out, r)
		w += rw
	}
	return string(out)
}

func PadToWidth(text string, targetWidth int, align string, fillChar rune) string {
	current := CalculateDisplayWidth(text)
	if current >= targetWidth {
		return text
	}

	pad := targetWidth - current
	left := pad / 2
	right := pad - left

	switch align {
	case "left":
		return text + repeat(fillChar, pad)
	case "right":
		return repeat(fillChar, pad) + text
	case "center":
		return repeat(fillChar, left) + text + repeat(fillChar, right)
	default:
		panic("invalid align (must be left, right, center)")
	}
}

func repeat(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = r
	}
	return string(b)
}
