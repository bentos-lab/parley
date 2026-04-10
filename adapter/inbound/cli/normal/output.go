package normal

import (
	"fmt"
	"io"
	"strings"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
)

// Output writes deterministic plain-text CLI output without styling.
type Output struct{}

// New constructs a plain-text output formatter.
// Parameters: none.
// Returns: a formatter for normal output.
func New() *Output {
	return &Output{}
}

// DebateDetails renders full debate details in plain labeled text.
// Parameters: writer is the output destination, details contains all debate information.
// Returns: an error if writing fails.
func (o *Output) DebateDetails(writer io.Writer, details cli.DebateDetailsOutput) error {
	lines := []string{
		"Header:",
		fmt.Sprintf("  ID: %s", details.Header.ID),
		fmt.Sprintf("  App: %s", details.Header.AppName),
		fmt.Sprintf("  Thinking: %s • %s", details.Header.LLMProvider, details.Header.LLMModel),
		fmt.Sprintf("  Speaking: %s • %s", details.Header.TTSProvider, details.Header.TTSModel),
		fmt.Sprintf("  Agents Count: %d", details.Header.AgentsCount),
		"",
		"Topic:",
		fmt.Sprintf("  %s", details.Topic),
		"",
		"Name:",
		fmt.Sprintf("  %s", details.Name),
		"",
		"Agents:",
	}
	if len(details.Agents) == 0 {
		lines = append(lines, "  (none)")
	} else {
		for _, agent := range details.Agents {
			lines = append(lines,
				fmt.Sprintf("  - ID: %s", agent.ID),
				fmt.Sprintf("    Name: %s", agent.Name),
				fmt.Sprintf("    Stance: %s", agent.Stance),
				fmt.Sprintf("    Voice: %s", agent.Voice),
			)
		}
	}

	lines = append(lines, "", "Rounds:")
	if len(details.Rounds) == 0 {
		lines = append(lines, "  (none)")
	} else {
		for _, round := range details.Rounds {
			lines = append(lines,
				fmt.Sprintf("  - Round: %d", round.Number),
				fmt.Sprintf("    Agent: %s (%s)", round.AgentName, round.AgentID),
				fmt.Sprintf("    Message: %s", round.Message),
				fmt.Sprintf("    Summary: %s", round.Summary),
				fmt.Sprintf("    Weakness: %s", round.Weakness),
				fmt.Sprintf("    New Point: %s", round.NewPoint),
				fmt.Sprintf("    Rebuttal: %s", round.Rebuttal),
			)
		}
	}

	lines = append(lines, "", "Summary:")
	for index, agent := range details.Agents {
		points := []string{"(no points)"}
		if index < len(details.Summary.Agents) && len(details.Summary.Agents[index]) > 0 {
			points = details.Summary.Agents[index]
		}
		lines = append(lines, fmt.Sprintf("  - %s (%s):", agent.Name, agent.Stance))
		for _, point := range points {
			lines = append(lines, fmt.Sprintf("    * %s", point))
		}
	}
	conclusion := strings.TrimSpace(details.Summary.Conclusion)
	if conclusion == "" {
		conclusion = "(none)"
	}
	lines = append(lines, fmt.Sprintf("  Conclusion: %s", conclusion))
	_, err := fmt.Fprintln(writer, strings.Join(lines, "\n"))
	return err
}
