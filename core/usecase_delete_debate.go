package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bentos-lab/parley/core/debate"
)

// DeleteDebateInput defines the inputs for deleting a debate.
type DeleteDebateInput struct {
	Filename string
}

// DeleteDebateUsecase deletes debates from storage.
type DeleteDebateUsecase struct{}

// Execute removes a debate file from disk.
func (u *DeleteDebateUsecase) Execute(input DeleteDebateInput) error {
	dir, err := debate.EnsureDebatesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, input.Filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete debate: %w", err)
	}
	return nil
}
