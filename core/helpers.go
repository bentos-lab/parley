package core

import (
	"fmt"
	"os"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// buildAgentSpecMap builds an agent map for voice assignment.
func buildAgentSpecMap(agents []debate.DebateAgent) map[string]contract.AgentSpec {
	result := make(map[string]contract.AgentSpec, len(agents))
	for _, agent := range agents {
		if agent.ID == "" {
			continue
		}
		result[agent.ID] = contract.AgentSpec{
			ID:        agent.ID,
			Name:      agent.Name,
			Stance:    agent.Stance,
			VoiceName: agent.VoiceName,
		}
	}
	return result
}

// applyAssignedVoices applies assigned voice names to agents.
func applyAssignedVoices(agents []debate.DebateAgent, assigned map[string]string) {
	for i := range agents {
		voiceName, ok := assigned[agents[i].ID]
		if !ok {
			continue
		}
		agents[i].VoiceName = voiceName
	}
}

// applyAgentVoices applies overrides to agents by ID.
func applyAgentVoices(agents []debate.DebateAgent, overrides map[string]string) error {
	if len(overrides) == 0 {
		return nil
	}
	lookup := make(map[string]int, len(agents))
	for i, agent := range agents {
		if agent.ID == "" {
			continue
		}
		lookup[agent.ID] = i
	}
	for agentID, voiceName := range overrides {
		index, ok := lookup[agentID]
		if !ok {
			return fmt.Errorf("unknown agent_id: %s", agentID)
		}
		agents[index].VoiceName = voiceName
	}
	return nil
}

// validateAgentVoices ensures overrides reference supported voices.
func validateAgentVoices(voices map[string]string, overrides map[string]string) error {
	for agentID, voiceName := range overrides {
		if voiceName == "" {
			continue
		}
		if _, ok := voices[voiceName]; !ok {
			return fmt.Errorf("invalid voice %q for agent %s", voiceName, agentID)
		}
	}
	return nil
}

// validateAgentsVoiceNames ensures existing agent voices are valid for the provider.
func validateAgentsVoiceNames(voices map[string]string, agents []debate.DebateAgent) error {
	for _, agent := range agents {
		if agent.VoiceName == "" {
			continue
		}
		if _, ok := voices[agent.VoiceName]; !ok {
			return fmt.Errorf("invalid voice %q for agent %s", agent.VoiceName, agent.ID)
		}
	}
	return nil
}

// needsVoiceAssignment determines if voice assignment should run.
func needsVoiceAssignment(voices map[string]string, agents []debate.DebateAgent) bool {
	seen := make(map[string]bool)
	hasDuplicate := false
	for _, agent := range agents {
		if agent.VoiceName == "" {
			return true
		}
		if _, ok := voices[agent.VoiceName]; !ok {
			return true
		}
		if seen[agent.VoiceName] {
			hasDuplicate = true
		}
		seen[agent.VoiceName] = true
	}
	if hasDuplicate && len(voices) >= len(agents) {
		return true
	}
	return false
}

// hasAnyVoiceName checks if any agent has a voice assigned.
func hasAnyVoiceName(agents []debate.DebateAgent) bool {
	for _, agent := range agents {
		if agent.VoiceName != "" {
			return true
		}
	}
	return false
}

// clearAgentVoices clears voice names for all agents.
func clearAgentVoices(debateItem *debate.Debate) {
	if debateItem == nil {
		return
	}
	for i := range debateItem.Agents {
		debateItem.Agents[i].VoiceName = ""
	}
}

// cloneDebate creates a deep copy of a debate.
func cloneDebate(source *debate.Debate) *debate.Debate {
	if source == nil {
		return nil
	}
	copyDebate := *source
	copyDebate.Agents = append([]debate.DebateAgent(nil), source.Agents...)
	copyDebate.Rounds = append([]debate.DebateRound(nil), source.Rounds...)
	return &copyDebate
}

// assignMissingAgentIDs fills empty agent IDs with sequential "agent-<n>" values.
func assignMissingAgentIDs(agents []debate.DebateAgent) {
	if len(agents) == 0 {
		return
	}
	used := make(map[string]bool, len(agents))
	for _, agent := range agents {
		if agent.ID == "" {
			continue
		}
		used[agent.ID] = true
	}
	nextIndex := 1
	for i := range agents {
		if agents[i].ID != "" {
			continue
		}
		for {
			candidate := fmt.Sprintf("agent-%d", nextIndex)
			nextIndex++
			if used[candidate] {
				continue
			}
			agents[i].ID = candidate
			used[candidate] = true
			break
		}
	}
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat %s: %w", path, err)
	}
	return !info.IsDir(), nil
}
