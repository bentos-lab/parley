package debate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/shared/audio"
)

// Debate stores information about a single debate session.
type Debate struct {
	Name           string        `json:"name"`
	NormalizedName string        `json:"normalized_name"`
	Topic          string        `json:"topic"`
	Agents         []DebateAgent `json:"agents"`
	Rounds         []DebateRound `json:"rounds"`
	TTSProvider    string        `json:"tts_provider"`
}

// DebateAgent represents a participant in the debate.
type DebateAgent struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Stance    string `json:"stance"`
	VoiceName string `json:"voice_name"`
}

// DebateRound represents a single speaking round in the debate.
type DebateRound struct {
	AgentID  string `json:"agent_id"`
	Message  string `json:"message"`
	Weakness string `json:"weakness"`
	NewPoint string `json:"new_point"`
	Rebuttal string `json:"rebuttal"`
	Summary  string `json:"summary"`
}

// DebateSummary stores the minimal information for listing debates.
type DebateSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

// CreateDebateInput defines inputs for constructing a debate.
type CreateDebateInput struct {
	Name        string
	Topic       string
	Agents      []DebateAgent
	TTSProvider string
}

const (
	// audioPaddingSeconds controls the silence between rounds in generated audio.
	audioPaddingDuration = 1000 * time.Millisecond
)

// Create initializes a new debate using generator contracts when needed.
func Create(ctx context.Context, input CreateDebateInput) (*Debate, error) {
	if input.Topic == "" {
		return nil, fmt.Errorf("topic is required")
	}
	name := input.Name
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	agents := input.Agents
	if len(agents) == 0 {
		return nil, fmt.Errorf("agents are required")
	}
	debate := &Debate{
		Name:           name,
		NormalizedName: normalizeName(name),
		Topic:          input.Topic,
		Agents:         agents,
		Rounds:         []DebateRound{},
		TTSProvider:    input.TTSProvider,
	}
	return debate, nil
}

// AppendRound appends a new round with the provided content.
func (d *Debate) AppendRound(agentID string, content string) DebateRound {
	round := DebateRound{
		AgentID: agentID,
		Message: content,
	}
	d.Rounds = append(d.Rounds, round)
	return round
}

// AppendRoundDetailed appends a new round with structured content.
func (d *Debate) AppendRoundDetailed(agentID string, message string, weakness string, newPoint string, rebuttal string, summary string) DebateRound {
	round := DebateRound{
		AgentID:  agentID,
		Message:  message,
		Weakness: weakness,
		NewPoint: newPoint,
		Rebuttal: rebuttal,
		Summary:  summary,
	}
	d.Rounds = append(d.Rounds, round)
	return round
}

// SelectAgentID chooses an agent according to the weighted appearance rules.
func (d *Debate) SelectAgentID() (string, error) {
	return d.selectAgentID()
}

// FindAgent returns the agent with the given ID.
func (d *Debate) FindAgent(agentID string) (DebateAgent, error) {
	return d.findAgent(agentID)
}

// FormatHistoryPrompt renders the debate history as a single user prompt.
func (d *Debate) FormatHistoryPrompt() string {
	return d.formatHistoryPrompt()
}

// OtherAgentNames returns the names of all other agents for the provided ID.
func (d *Debate) OtherAgentNames(agentID string) []string {
	return otherAgentNames(d.Agents, agentID)
}

// otherAgentNames collects the names of agents excluding the current agent ID.
// Parameters: agents is the full set of debate agents, currentAgentID is the active agent identifier.
// Returns: a slice of agent names for all other participants.
func otherAgentNames(agents []DebateAgent, currentAgentID string) []string {
	var names []string
	for _, agent := range agents {
		if agent.ID == currentAgentID {
			continue
		}
		if agent.Name == "" {
			continue
		}
		names = append(names, agent.Name)
	}
	return names
}

// Synthesize generates WAV audio for the debate rounds.
func (d *Debate) Synthesize(ctx context.Context, tts contract.TTS) (string, error) {
	if len(d.Rounds) == 0 {
		return "", fmt.Errorf("no rounds to synthesize")
	}
	var wavs []audio.WAV
	for _, round := range d.Rounds {
		expectedPath, err := d.roundAudioPath(round)
		if err != nil {
			return "", err
		}
		wav, ok, err := loadWAV(expectedPath)
		if err != nil {
			return "", err
		}
		if !ok {
			voiceName := d.voiceNameForRound(round)
			bytes, err := tts.Synthesize(ctx, round.Message, voiceName)
			if err != nil {
				return "", err
			}
			wav, err = audio.ParseWAV(bytes)
			if err != nil {
				return "", err
			}
			if err := audio.SaveWAV(expectedPath, wav); err != nil {
				return "", err
			}
		}
		wavs = append(wavs, wav)
	}
	combined, err := audio.Concat(wavs, audioPaddingDuration)
	if err != nil {
		return "", err
	}
	path, err := d.debateAudioPath()
	if err != nil {
		return "", err
	}
	if err := audio.SaveWAV(path, combined); err != nil {
		return "", err
	}
	return path, nil
}

// SynthesizeRound ensures a single round has audio and returns the audio path.
func (d *Debate) SynthesizeRound(ctx context.Context, tts contract.TTS, index int) (string, error) {
	if index < 0 || index >= len(d.Rounds) {
		return "", fmt.Errorf("round index out of range")
	}
	round := d.Rounds[index]
	expectedPath, err := d.roundAudioPath(round)
	if err != nil {
		return "", err
	}
	exists, err := fileExists(expectedPath)
	if err != nil {
		return "", err
	}
	if !exists {
		voiceName := d.voiceNameForRound(round)
		bytes, err := tts.Synthesize(ctx, round.Message, voiceName)
		if err != nil {
			return "", err
		}
		wav, err := audio.ParseWAV(bytes)
		if err != nil {
			return "", err
		}
		if err := audio.SaveWAV(expectedPath, wav); err != nil {
			return "", err
		}
	}
	return expectedPath, nil
}

// voiceNameForRound resolves the voice name for a round based on the agent ID.
// Parameters: round is the debate round to synthesize.
// Returns: the voice name, or an empty string when none is available.
func (d *Debate) voiceNameForRound(round DebateRound) string {
	if round.AgentID == "" {
		return ""
	}
	agent, err := d.findAgent(round.AgentID)
	if err != nil {
		return ""
	}
	return agent.VoiceName
}

// selectAgentID chooses an agent according to the weighted appearance rules.
func (d *Debate) selectAgentID() (string, error) {
	if len(d.Agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}
	counts := make(map[string]int)
	lastAgent := ""
	for i, round := range d.Rounds {
		if i == len(d.Rounds)-1 {
			lastAgent = round.AgentID
			break
		}
		if round.AgentID != "" {
			counts[round.AgentID]++
		}
	}
	maxCount := 0
	for _, agent := range d.Agents {
		if agent.ID == lastAgent {
			continue
		}
		if counts[agent.ID] > maxCount {
			maxCount = counts[agent.ID]
		}
	}
	type weightedAgent struct {
		ID     string
		Weight int
	}
	var pool []weightedAgent
	for _, agent := range d.Agents {
		if agent.ID == lastAgent {
			continue
		}
		weight := (maxCount - counts[agent.ID]) + 1
		pool = append(pool, weightedAgent{ID: agent.ID, Weight: weight})
	}
	if len(pool) == 0 {
		return d.Agents[0].ID, nil
	}
	var total int
	for _, agent := range pool {
		total += agent.Weight
	}
	seeded := rand.New(rand.NewSource(time.Now().UnixNano()))
	target := seeded.Intn(total)
	for _, agent := range pool {
		if target < agent.Weight {
			return agent.ID, nil
		}
		target -= agent.Weight
	}
	return pool[0].ID, nil
}

// findAgent returns the agent with the given ID.
func (d *Debate) findAgent(agentID string) (DebateAgent, error) {
	for _, agent := range d.Agents {
		if agent.ID == agentID {
			return agent, nil
		}
	}
	return DebateAgent{}, fmt.Errorf("agent not found: %s", agentID)
}

// formatHistoryPrompt renders the debate history as a single user prompt.
// Returns:
// - string: the full user prompt containing history and a turn guide.
func (d *Debate) formatHistoryPrompt() string {
	var history strings.Builder
	for _, round := range d.Rounds {
		name := round.AgentID
		if round.AgentID != "" {
			if agent, err := d.findAgent(round.AgentID); err == nil && agent.Name != "" {
				name = agent.Name
			}
		}
		if history.Len() > 0 {
			history.WriteString("\n")
		}
		fmt.Fprintf(&history, "%s: %s", name, round.Message)
	}
	if history.Len() == 0 {
		return "Since this is the opening turn, clearly establish your core thesis, define the " +
			"lens you will argue from, and introduce the strongest first argument that sets up " +
			"future rounds."
	}
	return fmt.Sprintf("%s\nThis is your turn to speak.", history.String())
}

const (
	debatesDirName = ".bentos/parley"
	fileTimeFormat = "2006-01-02-15-04-05"
	// defaultNormalizedName is used when a debate name has no valid characters.
	defaultNormalizedName = "debate"
)

// normalizeName normalizes the debate name to lowercase ASCII letters, digits, and dashes.
// Parameters: name is the raw debate name to normalize.
// Returns: the normalized name containing only a-z, 0-9, and '-' characters with single dashes.
func normalizeName(name string) string {
	if name == "" {
		return defaultNormalizedName
	}
	normalized := strings.ToLower(name)
	normalizedBytes := make([]byte, 0, len(normalized))
	lastWasDash := false
	for _, value := range normalized {
		if value >= 'a' && value <= 'z' {
			normalizedBytes = append(normalizedBytes, byte(value))
			lastWasDash = false
			continue
		}
		if value >= '0' && value <= '9' {
			normalizedBytes = append(normalizedBytes, byte(value))
			lastWasDash = false
			continue
		}
		if value == ' ' || value == '-' {
			if len(normalizedBytes) == 0 || lastWasDash {
				continue
			}
			normalizedBytes = append(normalizedBytes, '-')
			lastWasDash = true
		}
	}
	if lastWasDash && len(normalizedBytes) > 0 {
		normalizedBytes = normalizedBytes[:len(normalizedBytes)-1]
	}
	result := string(normalizedBytes)
	if result == "" {
		return defaultNormalizedName
	}
	return result
}

// debatesDir returns the path to the debates directory.
func debatesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, debatesDirName), nil
}

// EnsureDebatesDir creates the debates directory if it does not exist.
func EnsureDebatesDir() (string, error) {
	return ensureDebatesDir()
}

// ensureDebatesDir creates the debates directory if it does not exist.
func ensureDebatesDir() (string, error) {
	dir, err := debatesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create debates dir: %w", err)
	}
	return dir, nil
}

// Save persists the debate to a new file and returns the filename.
func (d *Debate) Save() (string, error) {
	if d.Name == "" {
		return "", fmt.Errorf("debate name is required")
	}
	d.NormalizedName = normalizeName(d.Name)
	dir, err := ensureDebatesDir()
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s.%s.json", d.NormalizedName, time.Now().Format(fileTimeFormat))
	path := filepath.Join(dir, filename)
	if err := writeDebate(path, d); err != nil {
		return "", err
	}
	return filename, nil
}

// SaveAs overwrites the debate at the specified filename.
func (d *Debate) SaveAs(filename string) error {
	d.NormalizedName = normalizeName(d.Name)
	dir, err := ensureDebatesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, filename)
	return writeDebate(path, d)
}

// LoadDebate loads a debate from the specified filename.
func LoadDebate(filename string) (*Debate, error) {
	dir, err := debatesDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read debate: %w", err)
	}
	var d Debate
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("decode debate: %w", err)
	}
	return &d, nil
}

// GetAllDebate loads all valid debate files from disk.
func GetAllDebate() ([]DebateSummary, error) {
	dir, err := debatesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []DebateSummary{}, nil
		}
		return nil, fmt.Errorf("read debates dir: %w", err)
	}
	type debateListItem struct {
		summary DebateSummary
		time    time.Time
		hasTime bool
	}
	var results []debateListItem
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		debateItem, err := LoadDebate(entry.Name())
		if err != nil {
			continue
		}
		timestamp, hasTime := parseDebateTimestamp(entry.Name())
		results = append(results, debateListItem{
			summary: DebateSummary{
				ID:    IDFromFilename(entry.Name()),
				Name:  debateItem.Name,
				Topic: debateItem.Topic,
			},
			time:    timestamp,
			hasTime: hasTime,
		})
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].hasTime != results[j].hasTime {
			return results[i].hasTime
		}
		if !results[i].hasTime {
			return false
		}
		if results[i].time.Equal(results[j].time) {
			return false
		}
		return results[i].time.After(results[j].time)
	})
	summaries := make([]DebateSummary, 0, len(results))
	for _, item := range results {
		summaries = append(summaries, item.summary)
	}
	return summaries, nil
}

// IDFromFilename derives a debate ID from a stored filename.
// Parameters: filename is the stored debate file name.
// Returns: the debate ID without the .json suffix.
func IDFromFilename(filename string) string {
	if strings.HasSuffix(filename, ".json") {
		return strings.TrimSuffix(filename, ".json")
	}
	return filename
}

// FilenameFromID builds the stored filename for a debate ID.
// Parameters: id is the debate identifier without the .json suffix.
// Returns: the stored filename that the JSON is saved as.
func FilenameFromID(id string) string {
	return fmt.Sprintf("%s.json", id)
}

// FilePathFromFilename builds the absolute path to the debate file for a stored filename.
// Parameters: filename is the stored debate file name.
// Returns: the full path to the file or an error when the debates directory cannot be resolved.
func FilePathFromFilename(filename string) (string, error) {
	dir, err := debatesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// parseDebateTimestamp extracts the timestamp from a debate filename.
// Parameters: filename is the debate file name to parse.
// Returns: the parsed timestamp and a boolean indicating whether it was found.
func parseDebateTimestamp(filename string) (time.Time, bool) {
	trimmed := strings.TrimSuffix(filename, ".json")
	lastDot := strings.LastIndex(trimmed, ".")
	if lastDot == -1 || lastDot == len(trimmed)-1 {
		return time.Time{}, false
	}
	segment := trimmed[lastDot+1:]
	parsed, err := time.Parse(fileTimeFormat, segment)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

// writeDebate writes debate data to disk with formatted JSON.
func writeDebate(path string, debate *Debate) error {
	data, err := json.MarshalIndent(debate, "", "  ")
	if err != nil {
		return fmt.Errorf("encode debate: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write debate: %w", err)
	}
	return nil
}

func (d *Debate) roundSignature(round DebateRound) (string, error) {
	payload := struct {
		Message     string `json:"message"`
		VoiceName   string `json:"voice_name"`
		TTSProvider string `json:"tts_provider"`
	}{
		Message:     round.Message,
		VoiceName:   d.voiceNameForRound(round),
		TTSProvider: d.TTSProvider,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode round signature payload: %w", err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (d *Debate) debateSignature() (string, error) {
	roundSignatures := make([]string, len(d.Rounds))
	for i, round := range d.Rounds {
		sig, err := d.roundSignature(round)
		if err != nil {
			return "", err
		}
		roundSignatures[i] = sig
	}
	data, err := json.Marshal(roundSignatures)
	if err != nil {
		return "", fmt.Errorf("encode debate signature payload: %w", err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (d *Debate) roundAudioPath(round DebateRound) (string, error) {
	signature, err := d.roundSignature(round)
	if err != nil {
		return "", err
	}
	dir, err := ensureDebatesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s.wav", signature)), nil
}

func (d *Debate) debateAudioPath() (string, error) {
	signature, err := d.debateSignature()
	if err != nil {
		return "", err
	}
	dir, err := ensureDebatesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s.wav", signature)), nil
}

// DebateAudioPath returns the expected audio path for the debate.
func (d *Debate) DebateAudioPath() (string, error) {
	return d.debateAudioPath()
}
