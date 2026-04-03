package core

import "github.com/bentos-lab/parley/core/debate"

// ListDebatesOutput is the result of listing debates.
type ListDebatesOutput struct {
	Items []debate.DebateSummary
}

// ListDebatesUsecase lists saved debates.
type ListDebatesUsecase struct{}

// Execute returns all debates.
func (u *ListDebatesUsecase) Execute() (ListDebatesOutput, error) {
	items, err := debate.GetAllDebate()
	if err != nil {
		return ListDebatesOutput{}, err
	}
	return ListDebatesOutput{Items: items}, nil
}
