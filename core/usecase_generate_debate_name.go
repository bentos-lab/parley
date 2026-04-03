package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/bentos-lab/parley/core/contract"
)

// GenerateDebateNameInput defines inputs for generating a debate name.
type GenerateDebateNameInput struct {
	Topic string
}

// GenerateDebateNameOutput is the result of a debate name generation.
type GenerateDebateNameOutput struct {
	Name string
}

// GenerateDebateNameUsecase generates debate names using an LLM.
type GenerateDebateNameUsecase struct {
	LLMResolver contract.Resolver[contract.LLM]
	LLMProvider string
	Model       string
}

const (
	// debateNameLLMTemperature sets the fixed sampling temperature for debate name generation.
	debateNameLLMTemperature = 0.7
	// debateNameLLMMaxTokens sets the fixed max token budget for debate name generation.
	debateNameLLMMaxTokens = 4096
)

// Execute generates a debate name for the provided topic.
func (u *GenerateDebateNameUsecase) Execute(ctx context.Context, input GenerateDebateNameInput) (GenerateDebateNameOutput, error) {
	if input.Topic == "" {
		return GenerateDebateNameOutput{}, fmt.Errorf("topic is required")
	}
	if u.LLMResolver == nil {
		return GenerateDebateNameOutput{}, fmt.Errorf("llm resolver is required")
	}
	if u.LLMProvider == "" {
		return GenerateDebateNameOutput{}, fmt.Errorf("llm_provider is required")
	}
	llm, err := u.LLMResolver.Resolve(u.LLMProvider)
	if err != nil {
		return GenerateDebateNameOutput{}, err
	}
	systemPrompt, err := renderPrompt("generate_debate_name.md", map[string]any{
		"Topic": input.Topic,
	})
	if err != nil {
		return GenerateDebateNameOutput{}, err
	}
	resp, err := llm.Generate(ctx, contract.LLMRequest{
		SystemInstruction: systemPrompt,
		Messages: []contract.LLMMessage{
			{Role: "user", Content: input.Topic},
		},
		Temperature: debateNameLLMTemperature,
		Model:       u.Model,
		MaxTokens:   debateNameLLMMaxTokens,
	})
	if err != nil {
		return GenerateDebateNameOutput{}, err
	}
	name := strings.TrimSpace(strings.Trim(resp, "\""))
	if name == "" {
		return GenerateDebateNameOutput{}, fmt.Errorf("generated name is empty")
	}
	return GenerateDebateNameOutput{Name: name}, nil
}
