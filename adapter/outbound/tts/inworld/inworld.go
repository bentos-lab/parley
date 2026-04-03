package inworld

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var voiceCatalog = map[string]string{
	"Callum":   "Casual and friendly Australian male voice, ideal for informal instructional content.",
	"Carter":   "Energetic, mature radio announcer-style male voice, great for storytelling, pep talks, and voiceovers.",
	"Chloe":    "Thoughtful, introspective youthful female voice, perfect for coming-of-age narratives, personal growth stories, and emotional teen dramas.",
	"Claire":   "Warm, gentle Eastern European female voice, ideal for bedtime stories, relaxation podcasts",
	"Veronica": "Intimidating, commanding female voice, perfect for ruthless antagonists, high-stakes negotiations, and chilling monologues.",
	"Rupert":   "Resonant, commanding British male voice, ideal for motivational speeches, epic film trailers, and dynamic corporate presentations.",
	"Marlene":  "Friendly, relaxed Southern female voice, ideal for home-style cooking tutorials, community event promotions, and downhome commercials.",
	"Elliot":   "A calm, steady male voice, suitable for nature documentaries, general informational content, and relaxed narrations.",
	"Edward":   "American male with a emphatic, confident and streetwise tone",
	"Tessa":    "Upbeat, conversational Australian female voice, perfect for lifestyle vlogs, playful advertisements, and engaging social media content.",
}

// Config holds Inworld TTS configuration.
type Config struct {
	APIKey     string
	ModelID    string
	SampleRate int
	Temp       float64
}

// Client implements the TTS contract using Inworld TTS.
type Client struct {
	config Config
	http   *http.Client
}

var (
	pauseTagRegex   = regexp.MustCompile(`<pause[^>]*>`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// NewClient creates a new Inworld TTS client.
// Parameters: config holds the Inworld TTS configuration settings.
// Returns: the initialized Inworld client.
func NewClient(config Config) *Client {
	return &Client{config: config, http: &http.Client{}}
}

// NormalizeText replaces pause tags with SSML breaks and normalizes whitespace.
// Parameters: text is the raw synthesized content that may include pause tags.
// Returns: the normalized text with SSML breaks for Inworld.
func NormalizeText(text string) string {
	normalizedNewlines := strings.ReplaceAll(text, "\r\n", " ")
	normalizedNewlines = strings.ReplaceAll(normalizedNewlines, "\n", " ")
	replaced := pauseTagRegex.ReplaceAllStringFunc(normalizedNewlines, func(tag string) string {
		switch strings.ToLower(tag) {
		case "<pause300>":
			return `<break time="300ms" />`
		case "<pause500>":
			return `<break time="500ms" />`
		case "<pause1000>":
			return `<break time="1s" />`
		default:
			return `<break time="500ms" />`
		}
	})
	return normalizeWhitespace(replaced)
}

// Synthesize generates WAV audio bytes for the given text using Inworld.
// Parameters: ctx controls cancellation, text is the content to synthesize, voiceName selects a required voice.
// Returns: the synthesized WAV bytes or an error if synthesis fails.
func (c *Client) Synthesize(ctx context.Context, text string, voiceName string) ([]byte, error) {
	if voiceName == "" {
		return nil, fmt.Errorf("voice name is required for Inworld TTS")
	}
	normalized := NormalizeText(text)
	payload := synthesizeRequest{
		Text:      normalized,
		VoiceID:   voiceName,
		ModelID:   c.config.ModelID,
		Temp:      c.config.Temp,
		AudioCfg:  audioConfig{AudioEncoding: "LINEAR16", SampleRateHertz: c.config.SampleRate},
		Normalize: "ON",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.inworld.ai/tts/v1/voice", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+c.config.APIKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	payloadBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("inworld error %d: %s", resp.StatusCode, string(payloadBytes))
	}
	var parsed synthesizeResponse
	if err := json.Unmarshal(payloadBytes, &parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(parsed.AudioContent)
	if err != nil {
		return nil, fmt.Errorf("decode audio content: %w", err)
	}
	return decoded, nil
}

// AgentVoices returns the voice catalog for Inworld.
// Parameters: none.
// Returns: a map of voice name to voice description.
func (c *Client) AgentVoices() map[string]string {
	return voiceCatalog
}

// normalizeWhitespace collapses runs of whitespace and trims leading/trailing space.
// Parameters: text is the string to normalize.
// Returns: a space-normalized string.
func normalizeWhitespace(text string) string {
	normalized := whitespaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(normalized)
}

type synthesizeRequest struct {
	Text      string      `json:"text"`
	VoiceID   string      `json:"voiceId"`
	ModelID   string      `json:"modelId"`
	AudioCfg  audioConfig `json:"audioConfig"`
	Temp      float64     `json:"temperature,omitempty"`
	Normalize string      `json:"applyTextNormalization,omitempty"`
}

type audioConfig struct {
	AudioEncoding   string `json:"audioEncoding"`
	SampleRateHertz int    `json:"sampleRateHertz"`
}

type synthesizeResponse struct {
	AudioContent string `json:"audioContent"`
}
