package native

import (
	"regexp"
	"strings"
)

// Client implements the TTS contract using native OS tools.
type Client struct{}

var (
	angleMarkupRegex  = regexp.MustCompile(`<[^>]*>`)
	squareMarkupRegex = regexp.MustCompile(`\[[^\]]*\]`)
	asteriskRegex     = regexp.MustCompile(`\*+`)
	whitespaceRegex   = regexp.MustCompile(`\s+`)
)

// NewClient creates a new native TTS client.
// Parameters: none.
// Returns: the initialized native client.
func NewClient() *Client {
	return &Client{}
}

// NormalizeText removes audio markup and normalizes whitespace for synthesis.
// Parameters: text is the raw synthesized content that may include markup tags.
// Returns: the cleaned text that native TTS can safely synthesize.
func NormalizeText(text string) string {
	stripped := angleMarkupRegex.ReplaceAllString(text, " ")
	stripped = squareMarkupRegex.ReplaceAllString(stripped, " ")
	stripped = asteriskRegex.ReplaceAllString(stripped, "")
	return normalizeWhitespace(stripped)
}

// AgentVoices returns the voice catalog for native TTS.
// Parameters: none.
// Returns: an empty map because native TTS does not expose a voice catalog here.
func (c *Client) AgentVoices() map[string]string {
	return map[string]string{}
}

// normalizeWhitespace collapses runs of whitespace and trims leading/trailing space.
// Parameters: text is the string to normalize.
// Returns: a space-normalized string.
func normalizeWhitespace(text string) string {
	normalized := whitespaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(normalized)
}
