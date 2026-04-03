package core

import (
	"context"

	"github.com/bentos-lab/parley/core/debate"
)

// CreateDebateInput defines the inputs for creating a debate.
type CreateDebateInput struct {
	Name        string
	Topic       string
	Agents      []debate.DebateAgent
	TTSProvider string
}

// CreateDebateOutput is the result of a debate creation.
type CreateDebateOutput struct {
	Debate   *debate.Debate
	Filename string
}

// CreateDebateUsecase creates and stores debates.
type CreateDebateUsecase struct {
	DefaultTTSProvider string // default provider used when no provider is supplied.
}

// Execute creates a debate and persists it to storage.
func (u *CreateDebateUsecase) Execute(ctx context.Context, input CreateDebateInput) (CreateDebateOutput, error) {
	provider, err := resolveTTSProvider(input.TTSProvider, "", u.DefaultTTSProvider)
	if err != nil {
		return CreateDebateOutput{}, err
	}
	debateItem, err := debate.Create(ctx, debate.CreateDebateInput{
		Name:        input.Name,
		Topic:       input.Topic,
		Agents:      input.Agents,
		TTSProvider: provider,
	})
	if err != nil {
		return CreateDebateOutput{}, err
	}
	assignMissingAgentIDs(debateItem.Agents)
	filename, err := debateItem.Save()
	if err != nil {
		return CreateDebateOutput{}, err
	}
	return CreateDebateOutput{
		Debate:   debateItem,
		Filename: filename,
	}, nil
}
