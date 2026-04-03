package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bentos-lab/parley/core/contract"
)

// ParleyCommandType defines WhatsApp /parley command names.
type ParleyCommandType string

const (
	ParleyCommandCreate  ParleyCommandType = "create"
	ParleyCommandList    ParleyCommandType = "list"
	ParleyCommandDelete  ParleyCommandType = "delete"
	ParleyCommandResume  ParleyCommandType = "resume"
	ParleyCommandAudio   ParleyCommandType = "audio"
	ParleyCommandUnknown ParleyCommandType = "unknown"
)

// ParleyCommandAgent represents an agent specification parsed from user text.
type ParleyCommandAgent struct {
	Name   string `json:"name"`
	Stance string `json:"stance"`
}

// ParleyCommand is the parsed command payload.
type ParleyCommand struct {
	Command   ParleyCommandType
	DebateID  string
	Topic     string
	NumAgents int
	NumRounds int
	Agents    []ParleyCommandAgent
}

// ParleyCommandHistoryMessage records the role + text to feed into the prompt.
type ParleyCommandHistoryMessage struct {
	Role    string
	Content string
}

// ParseParleyCommandInput configures the parser usecase.
type ParseParleyCommandInput struct {
	Message          string
	History          []ParleyCommandHistoryMessage
	DefaultNumAgents int
	DefaultNumRounds int
}

// ParseParleyCommandOutput returns the parsed command.
type ParseParleyCommandOutput struct {
	Command ParleyCommand
}

// ParseParleyCommandUsecase parses a /parley command via the LLM.
type ParseParleyCommandUsecase struct {
	LLMResolver contract.Resolver[contract.LLM]
	LLMProvider string
	Model       string
}

const (
	parseParleyLLMTemperature = 0.3
	parseParleyLLMMaxTokens   = 1024
)

// Execute converts the WhatsApp message + history into a structured command.
func (u *ParseParleyCommandUsecase) Execute(ctx context.Context, input ParseParleyCommandInput) (ParseParleyCommandOutput, error) {
	if input.Message == "" {
		return ParseParleyCommandOutput{Command: ParleyCommand{Command: ParleyCommandUnknown}}, fmt.Errorf("message is required")
	}
	if u.LLMResolver == nil {
		return ParseParleyCommandOutput{}, fmt.Errorf("llm resolver is required")
	}
	if u.LLMProvider == "" {
		return ParseParleyCommandOutput{}, fmt.Errorf("llm provider is required")
	}
	llm, err := u.LLMResolver.Resolve(u.LLMProvider)
	if err != nil {
		return ParseParleyCommandOutput{}, err
	}
	systemPrompt, err := renderPrompt("parse_parley_command.md", map[string]any{})
	if err != nil {
		return ParseParleyCommandOutput{}, err
	}
	historyContent := buildHistoryContent(input.History, input.Message)
	messages := []contract.LLMMessage{{Role: "user", Content: historyContent}}
	response, err := llm.GenerateJSON(ctx, contract.LLMRequest{
		SystemInstruction: systemPrompt,
		Messages:          messages,
		Temperature:       parseParleyLLMTemperature,
		Model:             u.Model,
		MaxTokens:         parseParleyLLMMaxTokens,
	}, parseParleySchema())
	if err != nil {
		return ParseParleyCommandOutput{}, err
	}
	parsed, err := parseParleyResponse(response)
	if err != nil {
		return ParseParleyCommandOutput{}, err
	}
	applyParleyDefaults(&parsed.Command, input.DefaultNumAgents, input.DefaultNumRounds)
	return ParseParleyCommandOutput{Command: parsed.Command}, nil
}

// buildHistoryContent stitches the provided history entries with the current message into a single string.
func buildHistoryContent(history []ParleyCommandHistoryMessage, current string) string {
	var builder strings.Builder
	if len(history) > 0 {
		builder.WriteString("History:\n")
		for _, entry := range history {
			content := strings.TrimSpace(entry.Content)
			if content == "" {
				continue
			}
			builder.WriteString(entry.Role)
			builder.WriteString(": ")
			builder.WriteString(content)
			builder.WriteString("\n")
		}
	}
	builder.WriteString("Current message:\n")
	builder.WriteString(current)
	return builder.String()
}

// parseParleySchema provides the JSON schema used to validate the command parser output.
func parseParleySchema() *contract.LLMJSONSchema {
	return &contract.LLMJSONSchema{
		Name: "parse_parley_command",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"translation": map[string]any{
					"type":        "string",
					"description": "the translation of original user request",
				},
				"reason": map[string]any{
					"type":        "string",
					"description": "the brief reason in extracting each parameters",
				},
				"command": map[string]any{
					"type": "string",
					"enum": []string{"create", "list", "delete", "resume", "audio", "unknown"},
				},
				"debate_id":  map[string]any{"type": "string"},
				"topic":      map[string]any{"type": "string"},
				"num_agents": map[string]any{"type": "integer"},
				"num_rounds": map[string]any{"type": "integer"},
				"agents": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name":   map[string]any{"type": "string"},
							"stance": map[string]any{"type": "string"},
						},
						"required":             []string{"name", "stance"},
						"additionalProperties": false,
					},
				},
			},
			"required":             []string{"reason", "command"},
			"additionalProperties": false,
		},
	}
}

type parleyResponseEnvelope struct {
	Command   string               `json:"command"`
	DebateID  string               `json:"debate_id"`
	Topic     string               `json:"topic"`
	NumAgents int                  `json:"num_agents"`
	NumRounds int                  `json:"num_rounds"`
	Agents    []ParleyCommandAgent `json:"agents"`
}

// parseParleyResponse converts the JSON produced by the LLM into the concrete command structure.
func parseParleyResponse(resp string) (ParseParleyCommandOutput, error) {
	var parsed parleyResponseEnvelope
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		return ParseParleyCommandOutput{}, fmt.Errorf("parse parley response: %w", err)
	}
	cmd := ParleyCommand{Command: ParleyCommandType(parsed.Command)}
	if cmd.Command == "" {
		cmd.Command = ParleyCommandUnknown
	}
	cmd.DebateID = strings.TrimSpace(parsed.DebateID)
	cmd.Topic = strings.TrimSpace(parsed.Topic)
	cmd.NumAgents = parsed.NumAgents
	cmd.NumRounds = parsed.NumRounds
	cmd.Agents = make([]ParleyCommandAgent, 0, len(parsed.Agents))
	for _, agent := range parsed.Agents {
		name := strings.TrimSpace(agent.Name)
		stance := strings.TrimSpace(agent.Stance)
		if name == "" || stance == "" {
			continue
		}
		cmd.Agents = append(cmd.Agents, ParleyCommandAgent{Name: name, Stance: stance})
	}
	return ParseParleyCommandOutput{Command: cmd}, nil
}

// applyParleyDefaults ensures optional command fields fall back to sane defaults.
func applyParleyDefaults(cmd *ParleyCommand, defaultAgents, defaultRounds int) {
	if cmd.NumAgents <= 0 {
		cmd.NumAgents = defaultAgents
	}
	if cmd.NumRounds <= 0 {
		cmd.NumRounds = defaultRounds
	}
	if cmd.Command == "" {
		cmd.Command = ParleyCommandUnknown
	}
}
