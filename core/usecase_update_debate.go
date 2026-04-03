package core

import (
	"fmt"

	"github.com/bentos-lab/parley/core/debate"
)

// UpdateDebateInput defines the inputs for updating a debate.
type UpdateDebateInput struct {
	Filename string
	Debate   *debate.Debate
}

// UpdateDebateUsecase updates a stored debate.
type UpdateDebateUsecase struct{}

// Execute overwrites a debate file with the provided content.
func (u *UpdateDebateUsecase) Execute(input UpdateDebateInput) error {
	if input.Debate == nil || input.Debate.Name == "" {
		return fmt.Errorf("debate name is required")
	}
	return input.Debate.SaveAs(input.Filename)
}
