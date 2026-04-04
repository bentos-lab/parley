package cli

import "io"

type DebateStreamOutput interface {
	DebateHeader(writer io.Writer, summary DebateSummary, agents []AgentRow) error
	DebateRound(writer io.Writer, roundNumber int, agentName string, message string, summary string, weakness string, newPoint string, rebuttal string) error
	DebateSummary(writer io.Writer, summary DebateSummaryOutput) error
	DebateResult(writer io.Writer, file string, id string) error
}

type CreateOutput interface {
	DebateStreamOutput
	DebateBasics(writer io.Writer, summary DebateBasics) error
	DebateName(writer io.Writer, name string) error
	DebateAgents(writer io.Writer, agents []AgentRow) error
}

type ResumeOutput interface {
	DebateStreamOutput
}

type ListOutput interface {
	// ListDebates renders the saved debate IDs to the writer.
	// Parameters: writer is the output destination, ids is the ordered list of debate IDs.
	// Returns: an error if writing fails.
	ListDebates(writer io.Writer, ids []string) error
}

type DebateSummary struct {
	Name        string `json:"name"`
	Topic       string `json:"topic"`
	TTSProvider string `json:"tts_provider"`
	File        string `json:"file"`

	AppName     string `json:"-"`
	AgentsCount int    `json:"-"`
	LLMProvider string `json:"-"`
	LLMModel    string `json:"-"`
	TTSModel    string `json:"-"`
}

type DebateSummaryDetail struct {
	Agents     map[string][]string `json:"agents"`
	Conclusion string              `json:"conclusion"`
}

type DebateSummaryOutput struct {
	Summary DebateSummaryDetail `json:"summary"`
	Agents  []AgentRow          `json:"agents"`
}

type DebateBasics struct {
	Topic       string `json:"topic"`
	TTSProvider string `json:"tts_provider"`

	AppName     string `json:"-"`
	LLMProvider string `json:"-"`
	LLMModel    string `json:"-"`
	TTSModel    string `json:"-"`
}

type AgentRow struct {
	ID     string
	Name   string
	Stance string
	Voice  string
}
