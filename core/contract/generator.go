package contract

import "context"

// GenerateDebateName defines an interface for generating debate names from a topic.
type GenerateDebateName interface {
	// GenerateName generates a debate name from the provided topic.
	GenerateName(ctx context.Context, topic string) (string, error)
}

// GenerateAgents defines an interface for generating debate agents.
type GenerateAgents interface {
	// GenerateAgents generates a list of agents from a topic and requested count.
	GenerateAgents(ctx context.Context, topic string, count int) ([]AgentSpec, error)
}

// AgentSpec represents an agent description produced by generators.
type AgentSpec struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Stance    string `json:"stance"`
	VoiceName string `json:"voice_name"`
}
