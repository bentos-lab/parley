package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/wiring"
)

// Resume continues an existing debate and generates the requested rounds.
// Parameters: ctx is the request context, usecases holds the debate usecases, output writes CLI output,
// runtime holds display values, id is the debate identifier, numRounds is the number of rounds to create.
// Returns: an error if resuming fails.

func Resume(ctx context.Context, usecases *wiring.Usecases, output ResumeOutput, runtime RuntimeInfo, debateID string, numRounds int) error {
	if debateID == "" {
		return fmt.Errorf("id is required")
	}
	filename := debate.FilenameFromID(debateID)
	loadOutput, err := usecases.LoadDebate.Execute(core.LoadDebateInput{Filename: filename})
	if err != nil {
		return err
	}
	if err := printDebateHeader(os.Stdout, output, loadOutput.Debate, filename, runtime); err != nil {
		return err
	}
	agentLookup := buildAgentMap(loadOutput.Debate.Agents)
	baseRound := len(loadOutput.Debate.Rounds)
	for i := range numRounds {
		roundOutput, err := usecases.CreateRound.Execute(ctx, core.CreateRoundInput{
			Filename: filename,
		})
		if err != nil {
			return err
		}
		roundNumber := baseRound + i + 1
		agentName := agentNameByID(roundOutput.Round.AgentID, agentLookup)
		if err := printRound(os.Stdout, output, roundNumber, agentName, roundOutput.Round.Message, roundOutput.Round.Summary, roundOutput.Round.Weakness, roundOutput.Round.NewPoint, roundOutput.Round.Rebuttal); err != nil {
			return err
		}
	}
	if usecases.GenerateDebateSummary == nil {
		return fmt.Errorf("summary generator is required")
	}
	summaryOutput, err := usecases.GenerateDebateSummary.Execute(ctx, core.GenerateDebateSummaryInput{
		Filename: filename,
		ForceNew: true,
	})
	if err != nil {
		return err
	}
	if err := output.DebateSummary(os.Stdout, DebateSummaryOutput{
		Summary: DebateSummaryDetail{
			Agents:     summaryOutput.Summary.Agents,
			Conclusion: summaryOutput.Summary.Conclusion,
		},
		Agents: buildAgentRows(loadOutput.Debate.Agents),
	}); err != nil {
		return err
	}
	filePath, err := debate.FilePathFromFilename(filename)
	if err != nil {
		return err
	}
	if err := output.DebateResult(os.Stdout, filePath, debateID); err != nil {
		return err
	}
	return nil
}
