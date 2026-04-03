package cli

import (
	"fmt"
	"os"

	"github.com/bentos-lab/parley/wiring"
)

// List outputs the saved debate IDs using the provided formatter.
// Parameters: usecases holds the debate usecases, output writes the IDs.
// Returns: an error if listing fails.
func List(usecases *wiring.Usecases, output ListOutput) error {
	if usecases == nil || usecases.ListDebates == nil {
		return fmt.Errorf("list usecase is required")
	}
	result, err := usecases.ListDebates.Execute()
	if err != nil {
		return err
	}
	ids := make([]string, len(result.Items))
	for i, item := range result.Items {
		ids[i] = item.ID
	}
	return output.ListDebates(os.Stdout, ids)
}
