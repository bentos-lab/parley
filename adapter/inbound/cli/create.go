package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/wiring"
)

// Create creates a new debate and generates the requested rounds.
// Parameters: ctx is the request context, usecases holds the debate usecases, output writes CLI output,
// runtime holds display values, topic is the debate topic, numAgents is the number of agents,
// numRounds is the number of rounds to create, ttsProvider is the selected TTS provider.
// Returns: an error if creation fails.
func Create(ctx context.Context, usecases *wiring.Usecases, output CreateOutput, runtime RuntimeInfo, topic string, numAgents int, numRounds int, ttsProvider string) error {
	if topic == "" {
		return fmt.Errorf("topic is required")
	}
	if err := printDebateBasics(os.Stdout, output, topic, ttsProvider, runtime); err != nil {
		return err
	}

	if usecases.GenerateDebateName == nil {
		return fmt.Errorf("name generator is required")
	}
	nameOutput, err := usecases.GenerateDebateName.Execute(ctx, core.GenerateDebateNameInput{
		Topic: topic,
	})
	if err != nil {
		return err
	}
	if err := printDebateName(os.Stdout, output, nameOutput.Name); err != nil {
		return err
	}

	if numAgents <= 0 {
		return fmt.Errorf("num-agents must be greater than zero")
	}
	if usecases.GenerateDebateAgents == nil {
		return fmt.Errorf("agent generator is required")
	}
	agentsOutput, err := usecases.GenerateDebateAgents.Execute(ctx, core.GenerateAgentsInput{
		Topic: topic,
		Count: numAgents,
	})
	if err != nil {
		return err
	}
	agents := agentsOutput.Agents
	if usecases.AssignDebateVoices == nil {
		return fmt.Errorf("voice assignment usecase is required")
	}
	voicesOutput, err := usecases.AssignDebateVoices.Execute(ctx, core.AssignDebateVoicesInput{
		Agents:      agents,
		TTSProvider: ttsProvider,
	})
	if err != nil {
		return err
	}
	agents = voicesOutput.Agents

	result, err := usecases.CreateDebate.Execute(ctx, core.CreateDebateInput{
		Name:        nameOutput.Name,
		Topic:       topic,
		Agents:      agents,
		TTSProvider: ttsProvider,
	})
	if err != nil {
		return err
	}
	if err := printDebateAgents(os.Stdout, output, result.Debate.Agents); err != nil {
		return err
	}
	agentLookup := buildAgentMap(result.Debate.Agents)
	for i := 0; i < numRounds; i++ {
		roundOutput, err := usecases.CreateRound.Execute(ctx, core.CreateRoundInput{
			Filename: result.Filename,
		})
		if err != nil {
			return err
		}
		roundNumber := i + 1
		agentName := agentNameByID(roundOutput.Round.AgentID, agentLookup)
		if err := printRound(os.Stdout, output, roundNumber, agentName, roundOutput.Round.Message, roundOutput.Round.Summary, roundOutput.Round.Weakness, roundOutput.Round.NewPoint, roundOutput.Round.Rebuttal); err != nil {
			return err
		}
	}
	if usecases.GenerateDebateSummary == nil {
		return fmt.Errorf("summary generator is required")
	}
	summaryOutput, err := usecases.GenerateDebateSummary.Execute(ctx, core.GenerateDebateSummaryInput{
		Filename: result.Filename,
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
		Agents: buildAgentRows(result.Debate.Agents),
	}); err != nil {
		return err
	}
	id := debate.IDFromFilename(result.Filename)
	filePath, err := debate.FilePathFromFilename(result.Filename)
	if err != nil {
		return err
	}
	if err := output.DebateResult(os.Stdout, filePath, id); err != nil {
		return err
	}
	return nil
}
