package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// GenerateDebateSummaryInput defines inputs for generating a debate summary.
type GenerateDebateSummaryInput struct {
	Filename string
	ForceNew bool
}

// GenerateDebateSummaryOutput is the result of generating a debate summary.
type GenerateDebateSummaryOutput struct {
	Summary debate.DebateSummaryDetail
}

// GenerateDebateSummaryUsecase generates and stores a debate summary using an LLM.
type GenerateDebateSummaryUsecase struct {
	LLMResolver contract.Resolver[contract.LLM]
	LLMProvider string
	Model       string
}

const (
	// summaryLLMTemperature sets the fixed sampling temperature for summary generation.
	summaryLLMTemperature = 0.3
	// summaryLLMMaxTokens sets the fixed max token budget for summary generation.
	summaryLLMMaxTokens = 2048
)

// Execute generates a debate summary and persists it to storage.
// Parameters: ctx is the request context, input holds the filename and force flag.
// Returns: the generated summary or an error when generation fails.
func (u *GenerateDebateSummaryUsecase) Execute(ctx context.Context, input GenerateDebateSummaryInput) (GenerateDebateSummaryOutput, error) {
	if input.Filename == "" {
		return GenerateDebateSummaryOutput{}, fmt.Errorf("filename is required")
	}
	debateItem, err := debate.LoadDebate(input.Filename)
	if err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	if len(debateItem.Rounds) == 0 {
		return GenerateDebateSummaryOutput{}, fmt.Errorf("no rounds to summarize")
	}
	if debateItem.Summary != nil && !input.ForceNew {
		return GenerateDebateSummaryOutput{Summary: *debateItem.Summary}, nil
	}
	if u.LLMResolver == nil {
		return GenerateDebateSummaryOutput{}, fmt.Errorf("llm resolver is required")
	}
	if u.LLMProvider == "" {
		return GenerateDebateSummaryOutput{}, fmt.Errorf("llm_provider is required")
	}
	llm, err := u.LLMResolver.Resolve(u.LLMProvider)
	if err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	systemPrompt, err := renderPrompt("summarize_debate.md", map[string]any{
		"Topic":  debateItem.Topic,
		"Agents": debateItem.Agents,
	})
	if err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	transcript := strings.TrimSpace(debateItem.FormatTranscript())
	messages := []contract.LLMMessage{
		{Role: "user", Content: transcript},
	}
	response, err := llm.GenerateJSON(ctx, contract.LLMRequest{
		SystemInstruction: systemPrompt,
		Messages:          messages,
		Temperature:       summaryLLMTemperature,
		Model:             u.Model,
		MaxTokens:         summaryLLMMaxTokens,
	}, summaryResponseSchema())
	if err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	var parsed struct {
		Agents     map[string][]string `json:"agents"`
		Conclusion string              `json:"conclusion"`
	}
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return GenerateDebateSummaryOutput{}, fmt.Errorf("parse summary json: %w, %s", err, response)
	}
	normalized := debate.DebateSummaryDetail{
		Agents:     normalizeSummaryAgents(parsed.Agents),
		Conclusion: strings.TrimSpace(parsed.Conclusion),
	}
	debateItem.Summary = &normalized
	if err := debateItem.SaveAs(input.Filename); err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	return GenerateDebateSummaryOutput{Summary: normalized}, nil
}

// normalizeSummaryAgents trims and filters empty summary points.
// Parameters: agents maps agent IDs to summary points.
// Returns: a cleaned map of agent IDs to summary points.
func normalizeSummaryAgents(agents map[string][]string) map[string][]string {
	if agents == nil {
		return map[string][]string{}
	}
	cleaned := make(map[string][]string, len(agents))
	for agentID, points := range agents {
		trimmed := make([]string, 0, len(points))
		for _, point := range points {
			value := strings.TrimSpace(point)
			if value == "" {
				continue
			}
			trimmed = append(trimmed, value)
		}
		cleaned[agentID] = trimmed
	}
	return cleaned
}

// summaryResponseSchema builds the JSON schema for the debate summary response.
// Returns: JSON schema payload describing the expected response shape.
func summaryResponseSchema() *contract.LLMJSONSchema {
	return &contract.LLMJSONSchema{
		Name: "debate_summary_response",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agents": map[string]any{
					"type": "object",
					"additionalProperties": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"conclusion": map[string]any{
					"type": "string",
				},
			},
			"required":             []string{"agents", "conclusion"},
			"additionalProperties": false,
		},
	}
}
