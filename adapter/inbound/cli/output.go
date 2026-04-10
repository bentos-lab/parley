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

// GetOutput renders full debate details loaded by debate ID.
type GetOutput interface {
	// DebateDetails renders full debate information including header, topic, name,
	// agents, rounds, and summary.
	// Parameters: writer is the output destination, details is the full debate payload.
	// Returns: an error if writing fails.
	DebateDetails(writer io.Writer, details DebateDetailsOutput) error
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
	Agents     [][]string `json:"agents"`
	Conclusion string     `json:"final_conclusion"`
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

// DebateHeaderOutput stores header information shown in get command output.
type DebateHeaderOutput struct {
	ID          string `json:"id"`
	AppName     string `json:"app_name"`
	LLMProvider string `json:"llm_provider"`
	LLMModel    string `json:"llm_model"`
	TTSProvider string `json:"tts_provider"`
	TTSModel    string `json:"tts_model"`
	AgentsCount int    `json:"agents_count"`
}

// DebateRoundOutput stores one debate round in get command output.
type DebateRoundOutput struct {
	Number    int    `json:"number"`
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	Message   string `json:"message"`
	Summary   string `json:"summary"`
	Weakness  string `json:"weakness"`
	NewPoint  string `json:"new_point"`
	Rebuttal  string `json:"rebuttal"`
}

// DebateDetailsOutput stores full debate info for get command output.
type DebateDetailsOutput struct {
	Header  DebateHeaderOutput  `json:"header"`
	Topic   string              `json:"topic"`
	Name    string              `json:"name"`
	Agents  []AgentRow          `json:"agents"`
	Rounds  []DebateRoundOutput `json:"rounds"`
	Summary DebateSummaryDetail `json:"summary"`
}
