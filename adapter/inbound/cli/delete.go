package cli

import (
	"context"
	"fmt"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/wiring"
)

// Delete removes a saved debate by ID.
// Parameters: ctx is the request context, usecases holds the debate usecases, debateID is the identifier to delete.
// Returns: an error if deletion fails.
func Delete(ctx context.Context, usecases *wiring.Usecases, debateID string) error {
	if debateID == "" {
		return fmt.Errorf("id is required")
	}
	if usecases == nil || usecases.DeleteDebate == nil {
		return fmt.Errorf("delete usecase is required")
	}
	return usecases.DeleteDebate.Execute(core.DeleteDebateInput{
		Filename: debate.FilenameFromID(debateID),
	})
}
