package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/wiring"
)

// Get loads a debate by ID and prints full debate details.
// Parameters: ctx is the request context, usecases holds debate usecases, output writes
// debate details, runtime carries fallback model metadata, debateID is the debate identifier.
// Returns: an error if loading or output rendering fails.
func Get(ctx context.Context, usecases *wiring.Usecases, output GetOutput, runtime RuntimeInfo, debateID string) error {
	_ = ctx
	if debateID == "" {
		return fmt.Errorf("id is required")
	}
	if usecases == nil || usecases.LoadDebate == nil {
		return fmt.Errorf("load usecase is required")
	}
	if output == nil {
		return fmt.Errorf("output is required")
	}

	filename := debate.FilenameFromID(debateID)
	loadOutput, err := usecases.LoadDebate.Execute(core.LoadDebateInput{Filename: filename})
	if err != nil {
		return err
	}

	details := buildDebateDetailsOutput(debateID, loadOutput.Debate, runtime)
	return output.DebateDetails(os.Stdout, details)
}

// buildDebateDetailsOutput maps a debate model to the full get command output payload.
// Parameters: debateID is the debate identifier, debateItem is the loaded debate model,
// runtime holds fallback model metadata.
// Returns: the complete debate details payload.
func buildDebateDetailsOutput(debateID string, debateItem *debate.Debate, runtime RuntimeInfo) DebateDetailsOutput {
	llmProvider := firstNonEmpty(debateItem.LLMProvider, runtime.LLMProvider)
	llmModel := resolveLLMModelForGet(debateItem, runtime)
	ttsProvider := displayTTSProvider(debateItem.TTSProvider)
	agentRows := buildAgentRows(debateItem.Agents)

	rounds := make([]DebateRoundOutput, 0, len(debateItem.Rounds))
	agentLookup := buildAgentMap(debateItem.Agents)
	for index, round := range debateItem.Rounds {
		rounds = append(rounds, DebateRoundOutput{
			Number:    index + 1,
			AgentID:   round.AgentID,
			AgentName: agentNameByID(round.AgentID, agentLookup),
			Message:   round.Message,
			Summary:   round.Summary,
			Weakness:  round.Weakness,
			NewPoint:  round.NewPoint,
			Rebuttal:  round.Rebuttal,
		})
	}

	summary := summaryDetailForGet(debateItem.Summary, len(debateItem.Agents))

	return DebateDetailsOutput{
		Header: DebateHeaderOutput{
			ID:          debateID,
			AppName:     appName,
			LLMProvider: llmProvider,
			LLMModel:    llmModel,
			TTSProvider: ttsProvider,
			TTSModel:    resolveTTSModel(ttsProvider, runtime.TTSModel),
			AgentsCount: len(debateItem.Agents),
		},
		Topic:   debateItem.Topic,
		Name:    debateItem.Name,
		Agents:  agentRows,
		Rounds:  rounds,
		Summary: summary,
	}
}

// summaryDetailForGet returns summary output, including placeholders when summary
// is missing from stored debate data.
// Parameters: summary is the stored debate summary, agentsCount is the number of
// debate agents.
// Returns: a summary payload suitable for rendering.
func summaryDetailForGet(summary *debate.DebateSummaryDetail, agentsCount int) DebateSummaryDetail {
	if summary == nil {
		return DebateSummaryDetail{
			Agents: make([][]string, agentsCount),
		}
	}
	agents := summary.Agents
	if agents == nil {
		agents = make([][]string, agentsCount)
	}
	return DebateSummaryDetail{
		Agents:     agents,
		Conclusion: summary.Conclusion,
	}
}

// resolveLLMModelForGet resolves the LLM model with debate value first, then provider
// defaults from runtime information.
// Parameters: debateItem is the loaded debate, runtime carries provider defaults.
// Returns: the resolved model string for display.
func resolveLLMModelForGet(debateItem *debate.Debate, runtime RuntimeInfo) string {
	if debateItem.LLMModel != "" {
		return debateItem.LLMModel
	}
	provider := firstNonEmpty(debateItem.LLMProvider, runtime.LLMProvider)
	switch provider {
	case "openai":
		return runtime.OpenAIModel
	case "anthropic":
		return runtime.AnthropicModel
	case "gemini":
		return runtime.GeminiModel
	default:
		return runtime.OpenAIModel
	}
}

// firstNonEmpty returns the first non-empty string from the provided values.
// Parameters: values is the ordered set of candidate strings.
// Returns: the first non-empty value, or an empty string when none exist.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
