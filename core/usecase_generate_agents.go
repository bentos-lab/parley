package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// GenerateAgentsInput defines inputs for generating debate agents.
type GenerateAgentsInput struct {
	Topic       string
	Count       int
	LLMProvider string
	LLMModel    string
}

// GenerateAgentsOutput is the result of a debate agent generation.
type GenerateAgentsOutput struct {
	Agents []debate.DebateAgent
}

// GenerateAgentsUsecase generates debate agents using an LLM.
type GenerateAgentsUsecase struct {
	LLMResolver contract.LLMResolver
	Defaults    LLMDefaults
}

type agentsResponse struct {
	Agents []agentResponse `json:"agents"`
}

type agentResponse struct {
	Name   string `json:"name"`
	Stance string `json:"stance"`
}

const (
	// agentsLLMTemperature sets the fixed sampling temperature for agent generation.
	agentsLLMTemperature = 0.7
	// agentsLLMMaxTokens sets the fixed max token budget for agent generation.
	agentsLLMMaxTokens = 4096
)

// Execute generates a list of debate agents for the provided topic.
func (u *GenerateAgentsUsecase) Execute(ctx context.Context, input GenerateAgentsInput) (GenerateAgentsOutput, error) {
	if input.Topic == "" {
		return GenerateAgentsOutput{}, fmt.Errorf("topic is required")
	}
	if input.Count <= 0 {
		return GenerateAgentsOutput{}, fmt.Errorf("count must be positive")
	}
	if u.LLMResolver == nil {
		return GenerateAgentsOutput{}, fmt.Errorf("llm resolver is required")
	}
	provider, model, err := ResolveLLMSelection(input.LLMProvider, input.LLMModel, "", "", u.Defaults)
	if err != nil {
		return GenerateAgentsOutput{}, err
	}
	llm, err := u.LLMResolver.Resolve(provider, model)
	if err != nil {
		return GenerateAgentsOutput{}, err
	}
	systemPrompt, err := renderPrompt("generate_agents.md", map[string]any{
		"Topic": input.Topic,
		"Count": input.Count,
	})
	if err != nil {
		return GenerateAgentsOutput{}, err
	}
	resp, err := llm.GenerateJSON(ctx, contract.LLMRequest{
		SystemInstruction: systemPrompt,
		Messages: []contract.LLMMessage{
			{Role: "user", Content: input.Topic},
		},
		Temperature: agentsLLMTemperature,
		MaxTokens:   agentsLLMMaxTokens,
	}, agentsResponseSchema())
	if err != nil {
		return GenerateAgentsOutput{}, err
	}
	var parsed agentsResponse
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		return GenerateAgentsOutput{}, fmt.Errorf("parse agents json: %w, %s", err, resp)
	}
	agents := make([]debate.DebateAgent, 0, len(parsed.Agents))
	for _, agent := range parsed.Agents {
		agents = append(agents, debate.DebateAgent{
			Name:   agent.Name,
			Stance: agent.Stance,
		})
	}
	return GenerateAgentsOutput{Agents: agents}, nil
}

// agentsResponseSchema builds the JSON schema for the agents generation response.
// Returns: JSON schema payload describing the expected response shape.
func agentsResponseSchema() *contract.LLMJSONSchema {
	return &contract.LLMJSONSchema{
		Name: "agents_response",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agents": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"type": "string",
							},
							"stance": map[string]any{
								"type": "string",
							},
						},
						"required":             []string{"name", "stance"},
						"additionalProperties": false,
					},
				},
			},
			"required":             []string{"agents"},
			"additionalProperties": false,
		},
	}
}
