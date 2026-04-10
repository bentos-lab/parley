package jsonoutput

import (
	"encoding/json"
	"io"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
)

type Output struct{}

func New() *Output {
	return &Output{}
}

func (o *Output) DebateCreated(writer io.Writer, name string, file string) error {
	data := struct {
		File string `json:"file"`
	}{
		File: file,
	}
	return encodeLine(writer, "debate_created", data)
}

func (o *Output) DebateHeader(writer io.Writer, summary cli.DebateSummary, agents []cli.AgentRow) error {
	data := struct {
		Summary cli.DebateSummary `json:"summary"`
		Agents  []cli.AgentRow    `json:"agents"`
	}{
		Summary: summary,
		Agents:  agents,
	}
	return encodeLine(writer, "debate_header", data)
}

func (o *Output) DebateBasics(writer io.Writer, summary cli.DebateBasics) error {
	data := struct {
		Summary cli.DebateBasics `json:"summary"`
	}{
		Summary: summary,
	}
	return encodeLine(writer, "debate_basics", data)
}

func (o *Output) DebateName(writer io.Writer, name string) error {
	data := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}
	return encodeLine(writer, "debate_name", data)
}

func (o *Output) DebateAgents(writer io.Writer, agents []cli.AgentRow) error {
	data := struct {
		Agents []cli.AgentRow `json:"agents"`
	}{
		Agents: agents,
	}
	return encodeLine(writer, "debate_agents", data)
}

func (o *Output) DebateRound(writer io.Writer, roundNumber int, agentName string, message string, summary string, weakness string, newPoint string, rebuttal string) error {
	data := struct {
		Number    int    `json:"number"`
		AgentName string `json:"agent_name"`
		Message   string `json:"message"`
		Summary   string `json:"summary"`
		Weakness  string `json:"weakness"`
		NewPoint  string `json:"new_point"`
		Rebuttal  string `json:"rebuttal"`
	}{
		Number:    roundNumber,
		AgentName: agentName,
		Message:   message,
		Summary:   summary,
		Weakness:  weakness,
		NewPoint:  newPoint,
		Rebuttal:  rebuttal,
	}
	return encodeLine(writer, "round", data)
}

// DebateSummary emits the debate summary output.
// Parameters: writer is the output destination, summary is the summary payload with agents metadata.
// Returns: an error if writing fails.
func (o *Output) DebateSummary(writer io.Writer, summary cli.DebateSummaryOutput) error {
	data := struct {
		Summary cli.DebateSummaryDetail `json:"summary"`
		Agents  []cli.AgentRow          `json:"agents"`
	}{
		Summary: summary.Summary,
		Agents:  summary.Agents,
	}
	return encodeLine(writer, "debate_summary", data)
}

func (o *Output) DebateResult(writer io.Writer, file string, id string) error {
	data := struct {
		File string `json:"file"`
		ID   string `json:"id,omitempty"`
	}{
		File: file,
	}
	if id != "" {
		data.ID = id
	}
	return encodeLine(writer, "result", data)
}

func (o *Output) ListDebates(writer io.Writer, ids []string) error {
	data := struct {
		Debates []string `json:"debates"`
	}{
		Debates: ids,
	}
	return encodeLine(writer, "debate_list", data)
}

// DebateDetails emits full debate details for the get command.
// Parameters: writer is the output destination, details is the full debate payload.
// Returns: an error if writing fails.
func (o *Output) DebateDetails(writer io.Writer, details cli.DebateDetailsOutput) error {
	return encodeLine(writer, "debate_get", details)
}

func (o *Output) InstallGuide(writer io.Writer, title string, guide string) error {
	data := struct {
		Title string `json:"title"`
		Guide string `json:"guide"`
	}{
		Title: title,
		Guide: guide,
	}
	return encodeLine(writer, "install_guide", data)
}

func encodeLine(writer io.Writer, outputType string, data any) error {
	payload := struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}{
		Type: outputType,
		Data: data,
	}
	encoder := json.NewEncoder(writer)
	return encoder.Encode(payload)
}
