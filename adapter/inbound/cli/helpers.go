package cli

import (
	"io"
	"net/url"
	"strings"

	"github.com/bentos-lab/parley/adapter/outbound/tts/native"
	"github.com/bentos-lab/parley/core/debate"
)

const appName = "Debate"

type RuntimeInfo struct {
	LLMBaseURL string
	LLMModel   string
	TTSModel   string
}

// printDebateHeader prints the debate header and agents list to the provided writer.
// Parameters: writer is the output destination, debateItem is the debate to summarize, filename is the debate file name.
// Returns: an error if writing fails.
func printDebateHeader(writer io.Writer, output DebateStreamOutput, debateItem *debate.Debate, filename string, runtime RuntimeInfo) error {
	ttsProvider := displayTTSProvider(debateItem.TTSProvider)
	summary := DebateSummary{
		Name:        debateItem.Name,
		Topic:       debateItem.Topic,
		TTSProvider: ttsProvider,
		File:        filename,
		AppName:     appName,
		AgentsCount: len(debateItem.Agents),
		LLMProvider: resolveLLMProvider(runtime.LLMBaseURL),
		LLMModel:    runtime.LLMModel,
		TTSModel:    resolveTTSModel(ttsProvider, runtime.TTSModel),
	}
	agentRows := buildAgentRows(debateItem.Agents)
	return output.DebateHeader(writer, summary, agentRows)
}

// printDebateBasics prints the initial debate basics immediately.
func printDebateBasics(writer io.Writer, output CreateOutput, topic string, ttsProvider string, runtime RuntimeInfo) error {
	displayProvider := displayTTSProvider(ttsProvider)
	basics := DebateBasics{
		Topic:       topic,
		TTSProvider: displayProvider,
		AppName:     appName,
		LLMProvider: resolveLLMProvider(runtime.LLMBaseURL),
		LLMModel:    runtime.LLMModel,
		TTSModel:    resolveTTSModel(displayProvider, runtime.TTSModel),
	}
	return output.DebateBasics(writer, basics)
}

// printDebateName prints the generated debate name.
func printDebateName(writer io.Writer, output CreateOutput, name string) error {
	return output.DebateName(writer, name)
}

// printDebateAgents prints the generated agents list.
func printDebateAgents(writer io.Writer, output CreateOutput, agents []debate.DebateAgent) error {
	return output.DebateAgents(writer, buildAgentRows(agents))
}

func buildAgentRows(agents []debate.DebateAgent) []AgentRow {
	agentRows := make([]AgentRow, 0, len(agents))
	for _, agent := range agents {
		agentRows = append(agentRows, AgentRow{
			ID:     agent.ID,
			Name:   agent.Name,
			Stance: agent.Stance,
			Voice:  displayVoiceName(agent.VoiceName),
		})
	}
	return agentRows
}

// buildAgentMap builds a lookup map of agents keyed by agent ID.
// Parameters: agents is the list of debate agents.
// Returns: a map keyed by agent ID.
func buildAgentMap(agents []debate.DebateAgent) map[string]debate.DebateAgent {
	lookup := make(map[string]debate.DebateAgent, len(agents))
	for _, agent := range agents {
		if agent.ID == "" {
			continue
		}
		lookup[agent.ID] = agent
	}
	return lookup
}

// agentNameByID resolves an agent name from an agent ID.
// Parameters: agentID is the agent identifier, agents is the lookup map of agents.
// Returns: the agent name, or "User" for empty IDs, or "Unknown" when missing.
func agentNameByID(agentID string, agents map[string]debate.DebateAgent) string {
	if agentID == "" {
		return "User"
	}
	agent, ok := agents[agentID]
	if !ok || agent.Name == "" {
		return "Unknown"
	}
	return agent.Name
}

// printRound prints a single round line to the provided writer.
// Parameters: writer is the output destination, roundNumber is the 1-based round index, agentName is the speaker name, message is the round content.
// Returns: an error if writing fails.
func printRound(writer io.Writer, output DebateStreamOutput, roundNumber int, agentName string, message string, summary string, weakness string, newPoint string, rebuttal string) error {
	return output.DebateRound(writer, roundNumber, agentName, message, summary, weakness, newPoint, rebuttal)
}

// displayTTSProvider returns a user-facing TTS provider label.
// Parameters: provider is the stored TTS provider value.
// Returns: the display value for the provider.
func displayTTSProvider(provider string) string {
	if provider == "" {
		return "native"
	}
	return provider
}

// displayVoiceName returns a user-facing voice name label.
// Parameters: voiceName is the stored voice name value.
// Returns: the display value for the voice name.
func displayVoiceName(voiceName string) string {
	if voiceName == "" {
		return "(none)"
	}
	return voiceName
}

func resolveLLMProvider(baseURL string) string {
	host := strings.ToLower(strings.TrimSpace(baseURL))
	if host == "" {
		return "custom"
	}
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		parsed, err := url.Parse(host)
		if err == nil && parsed.Host != "" {
			host = parsed.Host
		}
	}
	switch {
	case strings.Contains(host, "api.openai.com"):
		return "openai"
	case strings.Contains(host, "api.anthropic.com"):
		return "anthropic"
	case strings.Contains(host, "generativelanguage.googleapis.com"):
		return "gemini"
	default:
		return "custom"
	}
}

func resolveTTSModel(provider string, inworldModel string) string {
	if provider == "" || provider == "native" {
		return nativeEngineLabel()
	}
	return inworldModel
}

func nativeEngineLabel() string {
	label := native.CurrentInfo().Label
	start := strings.Index(label, "(")
	end := strings.LastIndex(label, ")")
	if start >= 0 && end > start {
		return strings.TrimSpace(label[start+1 : end])
	}
	return "default"
}
