package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

const (
	// summaryLLMTemperature sets the fixed sampling temperature for summary generation.
	summaryLLMTemperature = 0.3
	// summaryLLMMaxTokens sets the fixed max token budget for summary generation.
	summaryLLMMaxTokens = 2048
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
	LLMResolver contract.LLMResolver
	LLMProvider string
	Model       string
}

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
	llm, err := u.LLMResolver.Resolve(u.LLMProvider, u.Model)
	if err != nil {
		return GenerateDebateSummaryOutput{}, err
	}

	agentPoints := make([][]string, 0, len(debateItem.Agents))
	for _, agent := range debateItem.Agents {
		transcript := strings.TrimSpace(formatAgentTranscript(debateItem.Rounds, agent.ID))
		if transcript == "" {
			agentPoints = append(agentPoints, []string{})
			continue
		}
		systemPrompt, err := renderPrompt("summarize_agent_points.md", map[string]any{
			"Topic": debateItem.Topic,
			"Agent": agent,
		})
		if err != nil {
			return GenerateDebateSummaryOutput{}, err
		}
		response, err := llm.GenerateJSON(ctx, contract.LLMRequest{
			SystemInstruction: systemPrompt,
			Messages:          []contract.LLMMessage{{Role: "user", Content: transcript}},
			Temperature:       summaryLLMTemperature,
			Model:             u.Model,
			MaxTokens:         summaryLLMMaxTokens,
		}, summaryAgentPointsResponseSchema())
		if err != nil {
			return GenerateDebateSummaryOutput{}, err
		}
		var parsed struct {
			Points []string `json:"points"`
		}
		if err := json.Unmarshal([]byte(response), &parsed); err != nil {
			return GenerateDebateSummaryOutput{}, fmt.Errorf("parse agent summary json: %w, %s", err, response)
		}
		agentPoints = append(agentPoints, normalizeSummaryPoints(parsed.Points))
	}

	conclusionPrompt, err := renderPrompt("summarize_conclusion.md", map[string]any{
		"Topic":  debateItem.Topic,
		"Agents": debateItem.Agents,
	})
	if err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	conclusionTranscript := strings.TrimSpace(formatConclusionTranscript(debateItem.Rounds, debateItem.Agents))
	conclusionResponse, err := llm.GenerateJSON(ctx, contract.LLMRequest{
		SystemInstruction: conclusionPrompt,
		Messages:          []contract.LLMMessage{{Role: "user", Content: conclusionTranscript}},
		Temperature:       summaryLLMTemperature,
		Model:             u.Model,
		MaxTokens:         summaryLLMMaxTokens,
	}, summaryConclusionResponseSchema())
	if err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	var conclusionParsed struct {
		FinalConclusion string `json:"final_conclusion"`
	}
	if err := json.Unmarshal([]byte(conclusionResponse), &conclusionParsed); err != nil {
		return GenerateDebateSummaryOutput{}, fmt.Errorf("parse conclusion json: %w, %s", err, conclusionResponse)
	}

	normalized := debate.DebateSummaryDetail{
		Agents:     normalizeSummaryAgents(agentPoints),
		Conclusion: strings.TrimSpace(conclusionParsed.FinalConclusion),
	}
	debateItem.Summary = &normalized
	if err := debateItem.SaveAs(input.Filename); err != nil {
		return GenerateDebateSummaryOutput{}, err
	}
	return GenerateDebateSummaryOutput{Summary: normalized}, nil
}

// formatAgentTranscript renders a transcript containing only messages from a single agent.
// Parameters: rounds is the full debate round list, agentID is the agent to extract messages for.
// Returns: a newline-delimited list of messages for the agent, or an empty string when no rounds match.
func formatAgentTranscript(rounds []debate.DebateRound, agentID string) string {
	var transcript strings.Builder
	for _, round := range rounds {
		if round.AgentID != agentID {
			continue
		}
		message := strings.TrimSpace(round.Message)
		if message == "" {
			continue
		}
		if transcript.Len() > 0 {
			transcript.WriteString("\n")
		}
		transcript.WriteString(message)
	}
	return transcript.String()
}

// formatConclusionTranscript renders the full debate transcript with explicit speaker labels.
// Parameters: rounds is the full debate round list, agents is the declared debate agent list for name resolution.
// Returns: a newline-delimited transcript with "Speaker: message" lines.
func formatConclusionTranscript(rounds []debate.DebateRound, agents []debate.DebateAgent) string {
	agentNamesByID := make(map[string]string, len(agents))
	for _, agent := range agents {
		if agent.ID == "" || agent.Name == "" {
			continue
		}
		agentNamesByID[agent.ID] = agent.Name
	}
	var transcript strings.Builder
	for _, round := range rounds {
		message := strings.TrimSpace(round.Message)
		if message == "" {
			continue
		}
		speaker := resolveSpeakerName(round.AgentID, agentNamesByID)
		if transcript.Len() > 0 {
			transcript.WriteString("\n")
		}
		fmt.Fprintf(&transcript, "%s: %s", speaker, message)
	}
	return transcript.String()
}

// resolveSpeakerName resolves a user-facing speaker name for a round.
// Parameters: agentID is the round agent identifier (empty indicates the user), agentNamesByID maps agent IDs to display names.
// Returns: "User" when agentID is empty, the agent name when available, otherwise the raw agentID.
func resolveSpeakerName(agentID string, agentNamesByID map[string]string) string {
	if agentID == "" {
		return "User"
	}
	if name, ok := agentNamesByID[agentID]; ok && name != "" {
		return name
	}
	return agentID
}

// normalizeSummaryAgents trims and filters empty summary points.
// Parameters: agents is the ordered list of summary points per agent.
// Returns: a cleaned list of points per agent, preserving the input ordering.
func normalizeSummaryAgents(agents [][]string) [][]string {
	if agents == nil {
		return [][]string{}
	}
	cleaned := make([][]string, 0, len(agents))
	for _, points := range agents {
		trimmed := make([]string, 0, len(points))
		for _, point := range points {
			value := strings.TrimSpace(point)
			if value == "" {
				continue
			}
			trimmed = append(trimmed, value)
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}

// normalizeSummaryPoints trims and filters empty summary points.
// Parameters: points is the list of summary points for a single agent.
// Returns: a cleaned list of points, preserving the input ordering.
func normalizeSummaryPoints(points []string) []string {
	if points == nil {
		return []string{}
	}
	cleaned := make([]string, 0, len(points))
	for _, point := range points {
		value := strings.TrimSpace(point)
		if value == "" {
			continue
		}
		cleaned = append(cleaned, value)
	}
	return cleaned
}

// summaryAgentPointsResponseSchema builds the JSON schema for the per-agent points response.
// Returns: JSON schema payload describing the expected response shape.
func summaryAgentPointsResponseSchema() *contract.LLMJSONSchema {
	return &contract.LLMJSONSchema{
		Name: "debate_summary_agent_points_response",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"points": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			"required":             []string{"points"},
			"additionalProperties": false,
		},
	}
}

// summaryConclusionResponseSchema builds the JSON schema for the debate conclusion response.
// Returns: JSON schema payload describing the expected response shape.
func summaryConclusionResponseSchema() *contract.LLMJSONSchema {
	return &contract.LLMJSONSchema{
		Name: "debate_summary_conclusion_response",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"final_conclusion": map[string]any{
					"type": "string",
				},
			},
			"required":             []string{"final_conclusion"},
			"additionalProperties": false,
		},
	}
}
