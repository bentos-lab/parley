package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// CreateRoundInput defines the inputs for creating a debate round.
type CreateRoundInput struct {
	Filename    string
	AgentID     string
	Content     string
	LLMProvider string
	LLMModel    string
}

// CreateRoundOutput is the result of creating a debate round.
type CreateRoundOutput struct {
	Round debate.DebateRound
}

// CreateRoundUsecase creates rounds using the configured LLM.
type CreateRoundUsecase struct {
	LLMResolver contract.LLMResolver
	Defaults    LLMDefaults
}

// Execute creates a new round and persists the debate.
func (u *CreateRoundUsecase) Execute(ctx context.Context, input CreateRoundInput) (CreateRoundOutput, error) {
	debateItem, err := debate.LoadDebate(input.Filename)
	if err != nil {
		return CreateRoundOutput{}, err
	}
	round, err := u.createRound(ctx, debateItem, input)
	if err != nil {
		return CreateRoundOutput{}, err
	}
	if err := debateItem.SaveAs(input.Filename); err != nil {
		return CreateRoundOutput{}, err
	}
	return CreateRoundOutput{Round: round}, nil
}

const (
	// roundLLMTemperature sets the fixed sampling temperature for round generation.
	roundLLMTemperature = 0.7
	// roundLLMMaxTokens sets the fixed max token budget for round generation.
	roundLLMMaxTokens = 10000
)

func (u *CreateRoundUsecase) createRound(ctx context.Context, debateItem *debate.Debate, input CreateRoundInput) (debate.DebateRound, error) {
	if debateItem == nil {
		return debate.DebateRound{}, fmt.Errorf("debate is required")
	}
	if input.Content != "" {
		return debateItem.AppendRound(input.AgentID, input.Content), nil
	}
	if u.LLMResolver == nil {
		return debate.DebateRound{}, fmt.Errorf("llm resolver is required")
	}
	provider, model, err := ResolveLLMSelection(input.LLMProvider, input.LLMModel, debateItem.LLMProvider, debateItem.LLMModel, u.Defaults)
	if err != nil {
		return debate.DebateRound{}, err
	}
	llm, err := u.LLMResolver.Resolve(provider, model)
	if err != nil {
		return debate.DebateRound{}, err
	}
	selectedID := input.AgentID
	if selectedID == "" {
		selected, err := debateItem.SelectAgentID()
		if err != nil {
			return debate.DebateRound{}, err
		}
		selectedID = selected
	}
	agent, err := debateItem.FindAgent(selectedID)
	if err != nil {
		return debate.DebateRound{}, err
	}
	systemPrompt, err := renderPrompt("create_round.md", map[string]any{
		"Topic":       debateItem.Topic,
		"Agent":       agent,
		"OtherAgents": debateItem.OtherAgentNames(agent.ID),
	})
	if err != nil {
		return debate.DebateRound{}, err
	}
	historyPrompt := debateItem.FormatHistoryPrompt()
	messages := []contract.LLMMessage{
		{Role: "user", Content: historyPrompt},
	}
	response, err := llm.GenerateJSON(ctx, contract.LLMRequest{
		SystemInstruction: systemPrompt,
		Messages:          messages,
		Temperature:       roundLLMTemperature,
		MaxTokens:         roundLLMMaxTokens,
	}, roundResponseSchema())
	if err != nil {
		return debate.DebateRound{}, err
	}
	var parsed struct {
		Weakness   string `json:"weakness"`
		NewPoint   string `json:"new_point"`
		Rebuttal   string `json:"rebuttal"`
		FinalSpeak string `json:"final_speak"`
		Summary    string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return debate.DebateRound{}, fmt.Errorf("parse round json: %w, %s", err, response)
	}
	message := strings.TrimSpace(parsed.FinalSpeak)
	round := debateItem.AppendRoundDetailed(
		selectedID,
		message,
		strings.TrimSpace(parsed.Weakness),
		strings.TrimSpace(parsed.NewPoint),
		strings.TrimSpace(parsed.Rebuttal),
		strings.TrimSpace(parsed.Summary),
	)
	return round, nil
}

// roundResponseSchema builds the JSON schema for a debate round response.
// Returns: JSON schema payload describing the expected response shape.
func roundResponseSchema() *contract.LLMJSONSchema {
	return &contract.LLMJSONSchema{
		Name: "debate_round_response",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"weakness": map[string]any{
					"type": "string",
				},
				"new_point": map[string]any{
					"type": "string",
				},
				"rebuttal": map[string]any{
					"type": "string",
				},
				"final_speak": map[string]any{
					"type": "string",
				},
				"summary": map[string]any{
					"type": "string",
				},
			},
			"required":             []string{"weakness", "new_point", "rebuttal", "final_speak", "summary"},
			"additionalProperties": false,
		},
	}
}
