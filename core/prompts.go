package core

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed prompts/*.md
var promptFS embed.FS

// renderPrompt loads and renders a prompt template with data.
func renderPrompt(name string, data any) (string, error) {
	content, err := promptFS.ReadFile("prompts/" + name)
	if err != nil {
		return "", fmt.Errorf("read prompt %s: %w", name, err)
	}
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parse prompt %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render prompt %s: %w", name, err)
	}
	return buf.String(), nil
}
