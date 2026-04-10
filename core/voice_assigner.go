package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/bentos-lab/parley/core/contract"
)

// VoiceAssigner assigns voices to agents using an LLM.
type VoiceAssigner struct {
	llmResolver contract.LLMResolver
	defaults    LLMDefaults
}

const (
	// voiceAssignLLMTemperature sets the fixed sampling temperature for voice assignment.
	voiceAssignLLMTemperature = 0.7
	// voiceAssignLLMMaxTokens sets the fixed max token budget for voice assignment.
	voiceAssignLLMMaxTokens = 4096
)

// NewVoiceAssigner creates a VoiceAssigner with dependencies.
func NewVoiceAssigner(llmResolver contract.LLMResolver, defaults LLMDefaults) *VoiceAssigner {
	return &VoiceAssigner{llmResolver: llmResolver, defaults: defaults}
}

// AssignVoices assigns voices to agents based on the available catalog and existing assignments.
func (g *VoiceAssigner) AssignVoices(ctx context.Context, voices map[string]string, agents map[string]contract.AgentSpec) (map[string]string, error) {
	result := map[string]string{}
	if len(agents) == 0 {
		return result, nil
	}
	if g.llmResolver == nil {
		return nil, fmt.Errorf("llm resolver is required")
	}
	provider, model, err := ResolveEffectiveLLMSelection(ctx, "", "", g.defaults)
	if err != nil {
		return nil, err
	}
	llm, err := g.llmResolver.Resolve(provider, model)
	if err != nil {
		return nil, err
	}
	filteredVoices, remainingAgents := filterAssignedVoices(result, voices, agents)
	if len(filteredVoices) == 0 || len(remainingAgents) == 0 {
		return result, nil
	}
	voicesText := formatVoiceLines(filteredVoices)
	agentsText := formatAgentLines(remainingAgents)
	systemPrompt, err := renderPrompt("assign_debate_voices.md", map[string]any{
		"VoicesText": voicesText,
		"AgentsText": agentsText,
	})
	if err != nil {
		return nil, err
	}
	resp, err := llm.GenerateJSON(ctx, contract.LLMRequest{
		SystemInstruction: systemPrompt,
		Messages:          []contract.LLMMessage{{Role: "user", Content: "Assign voices to agents."}},
		Temperature:       voiceAssignLLMTemperature,
		MaxTokens:         voiceAssignLLMMaxTokens,
	}, voiceAssignmentSchema())
	if err != nil {
		assigned := normalizeVoiceAssignments(filteredVoices, remainingAgents, nil)
		return mergeAssignments(result, assigned), nil
	}
	parsed := map[string]string{}
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		assigned := normalizeVoiceAssignments(filteredVoices, remainingAgents, nil)
		return mergeAssignments(result, assigned), nil
	}
	assigned := normalizeVoiceAssignments(filteredVoices, remainingAgents, parsed)
	return mergeAssignments(result, assigned), nil
}

// voiceAssignmentSchema builds the JSON schema for voice assignment responses.
func voiceAssignmentSchema() *contract.LLMJSONSchema {
	return &contract.LLMJSONSchema{
		Name: "voice_assignments",
		Schema: map[string]any{
			"type":                 "object",
			"additionalProperties": map[string]any{"type": "string"},
		},
	}
}

// normalizeVoiceAssignments validates and completes assignments with uniqueness rules.
func normalizeVoiceAssignments(voices map[string]string, agents map[string]contract.AgentSpec, proposed map[string]string) map[string]string {
	agentIDs := sortedAgentIDs(agents)
	result := make(map[string]string, len(agents))
	used := make(map[string]bool)
	for _, agentID := range agentIDs {
		candidate := ""
		if proposed != nil {
			candidate = proposed[agentID]
		}
		if candidate == "" {
			candidate = agents[agentID].VoiceName
		}
		if candidate == "" {
			continue
		}
		if _, ok := voices[candidate]; !ok {
			continue
		}
		if used[candidate] {
			continue
		}
		result[agentID] = candidate
		used[candidate] = true
	}
	remainingAgents := missingAgentIDs(agentIDs, result)
	availableVoices := availableVoiceNames(voices, used)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, agentID := range remainingAgents {
		if len(availableVoices) > 0 {
			index := rng.Intn(len(availableVoices))
			selected := availableVoices[index]
			result[agentID] = selected
			used[selected] = true
			availableVoices = removeVoiceAt(availableVoices, index)
			continue
		}
		voice := randomVoiceName(rng, voices)
		if voice == "" {
			result[agentID] = ""
			continue
		}
		result[agentID] = voice
	}
	return result
}

// filterAssignedVoices builds the preset assignments and removes their voices from the pool.
func filterAssignedVoices(result map[string]string, voices map[string]string, agents map[string]contract.AgentSpec) (map[string]string, map[string]contract.AgentSpec) {
	filteredVoices := copyVoiceMap(voices)
	remainingAgents := make(map[string]contract.AgentSpec, len(agents))
	for id, agent := range agents {
		if id == "" {
			continue
		}
		if agent.VoiceName == "" {
			remainingAgents[id] = agent
			continue
		}
		if _, ok := filteredVoices[agent.VoiceName]; !ok {
			remainingAgents[id] = agent
			continue
		}
		result[id] = agent.VoiceName
		delete(filteredVoices, agent.VoiceName)
	}
	return filteredVoices, remainingAgents
}

// sortedAgentIDs returns agent IDs in stable order.
func sortedAgentIDs(agents map[string]contract.AgentSpec) []string {
	ids := make([]string, 0, len(agents))
	for id := range agents {
		if id == "" {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// missingAgentIDs returns agent IDs missing from the assignment map.
func missingAgentIDs(orderedIDs []string, assigned map[string]string) []string {
	var missing []string
	for _, id := range orderedIDs {
		if assigned[id] == "" {
			missing = append(missing, id)
		}
	}
	return missing
}

// availableVoiceNames returns unused voices in stable order.
func availableVoiceNames(voices map[string]string, used map[string]bool) []string {
	result := make([]string, 0, len(voices))
	for name := range voices {
		if used[name] {
			continue
		}
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// randomVoiceName selects a random voice name from the catalog.
func randomVoiceName(rng *rand.Rand, voices map[string]string) string {
	if len(voices) == 0 {
		return ""
	}
	names := make([]string, 0, len(voices))
	for name := range voices {
		names = append(names, name)
	}
	sort.Strings(names)
	return names[rng.Intn(len(names))]
}

// removeVoiceAt removes an item from a slice at the provided index.
func removeVoiceAt(items []string, index int) []string {
	if index < 0 || index >= len(items) {
		return items
	}
	return append(items[:index], items[index+1:]...)
}

// formatVoiceLines renders voice catalog entries for prompts.
func formatVoiceLines(voices map[string]string) string {
	names := make([]string, 0, len(voices))
	for name := range voices {
		names = append(names, name)
	}
	sort.Strings(names)
	lines := make([]string, 0, len(names))
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("- %s: %s", name, voices[name]))
	}
	return strings.Join(lines, "\n")
}

// formatAgentLines renders agent entries for prompts.
func formatAgentLines(agents map[string]contract.AgentSpec) string {
	ids := sortedAgentIDs(agents)
	lines := make([]string, 0, len(ids))
	for _, id := range ids {
		agent := agents[id]
		lines = append(lines, fmt.Sprintf("- %s: %s (%s)", agent.ID, agent.Name, agent.Stance))
	}
	return strings.Join(lines, "\n")
}

// copyVoiceMap copies a voice catalog map.
func copyVoiceMap(voices map[string]string) map[string]string {
	result := make(map[string]string, len(voices))
	for key, value := range voices {
		result[key] = value
	}
	return result
}

// mergeAssignments merges two assignment maps, preferring existing entries.
func mergeAssignments(base map[string]string, additions map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(additions))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range additions {
		if result[key] != "" {
			continue
		}
		result[key] = value
	}
	return result
}
