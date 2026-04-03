package core

import "github.com/bentos-lab/parley/core/debate"

// LoadDebateInput defines the inputs for loading a debate.
type LoadDebateInput struct {
	Filename string
}

// LoadDebateOutput is the result of loading a debate.
type LoadDebateOutput struct {
	Debate *debate.Debate
}

// LoadDebateUsecase loads debates from storage.
type LoadDebateUsecase struct{}

// Execute loads a debate by filename.
func (u *LoadDebateUsecase) Execute(input LoadDebateInput) (LoadDebateOutput, error) {
	debateItem, err := debate.LoadDebate(input.Filename)
	if err != nil {
		return LoadDebateOutput{}, err
	}
	return LoadDebateOutput{Debate: debateItem}, nil
}
