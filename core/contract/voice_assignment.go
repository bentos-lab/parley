package contract

import "context"

// AssignVoices defines a contract for assigning voices to agents.
type AssignVoices interface {
	// AssignVoices assigns voices to agents using the available voice catalog.
	// Parameters:
	// - ctx: context for request cancellation and deadlines.
	// - voices: map of voice name to voice description.
	// - agents: map of agent ID to agent specification (including voice_name).
	// Returns:
	// - map[string]string: complete map of agent ID to assigned voice name.
	// - error: non-nil when assignment fails.
	AssignVoices(ctx context.Context, voices map[string]string, agents map[string]AgentSpec) (map[string]string, error)
}
