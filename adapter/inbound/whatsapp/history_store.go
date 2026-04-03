package whatsapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bentos-lab/parley/core"
)

const historyCapacity = 10

// historyStore manages per-chat history entries and persists them to disk.
type historyStore struct {
	mu      sync.Mutex
	entries map[string][]core.ParleyCommandHistoryMessage
	path    string
}

// newHistoryStore loads the persisted history from disk and prepares the in-memory cache.
// Parameters: none.
// Returns: initialized historyStore or an error if the file cannot be read or parsed.
func newHistoryStore() (*historyStore, error) {
	path, err := defaultHistoryPath()
	if err != nil {
		return nil, err
	}
	store := &historyStore{
		entries: make(map[string][]core.ParleyCommandHistoryMessage),
		path:    path,
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

// add inserts a new entry for the chat while enforcing prefix filtering and capacity limits.
// Parameters: chat identifies the WhatsApp chat and entry carries the role/content.
// Returns: error when writing the updated cache fails, nil otherwise.
func (h *historyStore) add(chat string, entry core.ParleyCommandHistoryMessage) error {
	content := strings.TrimSpace(entry.Content)
	if content == "" || !shouldStoreHistory(content) {
		return nil
	}
	entry.Content = content

	h.mu.Lock()
	existing := append(h.entries[chat], entry)
	if len(existing) > historyCapacity {
		existing = existing[len(existing)-historyCapacity:]
	}
	h.entries[chat] = existing
	snapshot := copyHistoryMap(h.entries)
	h.mu.Unlock()

	return h.persist(snapshot)
}

// snapshot returns an isolated copy of a chat's history to avoid data races.
// Parameters: chat identifies the WhatsApp chat whose history is requested.
// Returns: a slice containing the stored entries for that chat.
func (h *historyStore) snapshot(chat string) []core.ParleyCommandHistoryMessage {
	h.mu.Lock()
	defer h.mu.Unlock()
	entries := h.entries[chat]
	copyEntries := make([]core.ParleyCommandHistoryMessage, len(entries))
	copy(copyEntries, entries)
	return copyEntries
}

// load reads the persisted JSON file, filters entries, and populates the in-memory cache.
// Parameters: none.
// Returns: error when the file cannot be read (other than not-exist) or parsed.
func (h *historyStore) load() error {
	data, err := os.ReadFile(h.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read history file: %w", err)
	}
	var disk map[string][]core.ParleyCommandHistoryMessage
	if err := json.Unmarshal(data, &disk); err != nil {
		return fmt.Errorf("parse history file: %w", err)
	}
	for chat, entries := range disk {
		filtered := filterHistoryEntries(entries)
		if len(filtered) == 0 {
			continue
		}
		h.entries[chat] = filtered
	}
	return nil
}

// persist writes the provided cache snapshot to disk in JSON format.
// Parameters: snapshot is a deep copy of the history map that should be stored.
// Returns: error when the write fails.
func (h *historyStore) persist(snapshot map[string][]core.ParleyCommandHistoryMessage) error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0o755); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	if err := os.WriteFile(h.path, payload, 0o644); err != nil {
		return fmt.Errorf("write history file: %w", err)
	}
	return nil
}

// filterHistoryEntries trims, filters, and caps the provided entries, keeping only valid history entries.
// Parameters: entries is the raw slice loaded from disk.
// Returns: the sanitized slice limited to historyCapacity.
func filterHistoryEntries(entries []core.ParleyCommandHistoryMessage) []core.ParleyCommandHistoryMessage {
	var filtered []core.ParleyCommandHistoryMessage
	for _, entry := range entries {
		content := strings.TrimSpace(entry.Content)
		if content == "" || !shouldStoreHistory(content) {
			continue
		}
		filtered = append(filtered, core.ParleyCommandHistoryMessage{
			Role:    entry.Role,
			Content: content,
		})
	}
	if len(filtered) > historyCapacity {
		return filtered[len(filtered)-historyCapacity:]
	}
	return filtered
}

// shouldStoreHistory checks whether the provided content is allowed to persist.
// Parameters: content is the trimmed message string.
// Returns: true when the message starts with /parley or [parley].
func shouldStoreHistory(content string) bool {
	return strings.HasPrefix(content, "/parley") || strings.HasPrefix(content, "[parley]")
}

// copyHistoryMap creates a deep copy of the given history map so it can be persisted safely.
// Parameters: src is the source cache to clone.
// Returns: the deep-copied map.
func copyHistoryMap(src map[string][]core.ParleyCommandHistoryMessage) map[string][]core.ParleyCommandHistoryMessage {
	dest := make(map[string][]core.ParleyCommandHistoryMessage, len(src))
	for chat, entries := range src {
		copied := make([]core.ParleyCommandHistoryMessage, len(entries))
		copy(copied, entries)
		dest[chat] = copied
	}
	return dest
}

// defaultHistoryPath derives the hardcoded location for the WhatsApp history cache.
// Parameters: none.
// Returns: the default whatsapp.history.json path under the connect directory inside ~/.bentos.
func defaultHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine user home: %w", err)
	}
	return filepath.Join(home, connectDirName, "whatsapp.history.json"), nil
}
