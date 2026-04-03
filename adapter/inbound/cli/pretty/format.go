package pretty

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	roundPlainStyle      = lipgloss.NewStyle()
	roundBoldStyle       = lipgloss.NewStyle().Bold(true)
	roundBoldItalicStyle = lipgloss.NewStyle().Bold(true).Italic(true)
)

func formatRoundMessage(message string) string {
	if message == "" {
		return ""
	}

	var out strings.Builder
	var buf strings.Builder
	bold := false
	italic := false

	flush := func() {
		if buf.Len() == 0 {
			return
		}
		out.WriteString(renderRoundWithStyle(buf.String(), bold, italic))
		buf.Reset()
	}

	for i := 0; i < len(message); {
		if tag := pauseTagAt(message, i); tag != "" {
			flush()
			out.WriteString(roundBoldStyle.Render("[pause]"))
			i += len(tag)
			continue
		}

		if message[i] == '[' {
			end := strings.IndexByte(message[i+1:], ']')
			if end >= 0 {
				flush()
				token := message[i : i+1+end+1]
				out.WriteString(roundBoldStyle.Render(token))
				i += 1 + end + 1
				continue
			}
		}

		if strings.HasPrefix(message[i:], "**") {
			if italic || hasDoubleAsteriskMatch(message, i+2) {
				flush()
				italic = !italic
				i += 2
				continue
			}
			buf.WriteString("**")
			i += 2
			continue
		}

		if message[i] == '*' {
			if i+1 < len(message) && message[i+1] == '*' {
				buf.WriteString("**")
				i += 2
				continue
			}
			if bold || hasSingleAsteriskMatch(message, i+1) {
				flush()
				bold = !bold
				i++
				continue
			}
		}

		buf.WriteByte(message[i])
		i++
	}

	flush()
	return out.String()
}

func renderRoundWithStyle(text string, bold bool, italic bool) string {
	if italic {
		return roundBoldItalicStyle.Render(text)
	}
	if bold {
		return roundBoldStyle.Render(text)
	}
	return roundPlainStyle.Render(text)
}

func pauseTagAt(message string, index int) string {
	if strings.HasPrefix(message[index:], "<pause300>") {
		return "<pause300>"
	}
	if strings.HasPrefix(message[index:], "<pause500>") {
		return "<pause500>"
	}
	if strings.HasPrefix(message[index:], "<pause1000>") {
		return "<pause1000>"
	}
	return ""
}

func hasDoubleAsteriskMatch(message string, start int) bool {
	return strings.Contains(message[start:], "**")
}

func hasSingleAsteriskMatch(message string, start int) bool {
	for i := start; i < len(message); i++ {
		if message[i] != '*' {
			continue
		}
		if i+1 < len(message) && message[i+1] == '*' {
			i++
			continue
		}
		return true
	}
	return false
}
